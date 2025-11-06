package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andrew/orquestador-notificacion/internal/config"
	"github.com/andrew/orquestador-notificacion/internal/handler"
	kafkaPkg "github.com/andrew/orquestador-notificacion/internal/kafka"
	"github.com/andrew/orquestador-notificacion/internal/logger"
	"github.com/andrew/orquestador-notificacion/internal/processor"
	"github.com/andrew/orquestador-notificacion/internal/service"

	kafka "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

const VERSION = "1.0.0"

var startTime = time.Now()

type HealthResponse struct {
	Status        string `json:"status"`
	Version       string `json:"version"`
	Uptime        string `json:"uptime"`
	UptimeSeconds int64  `json:"uptimeSeconds"`
}

func formatUptime(duration time.Duration) string {
	seconds := int64(duration.Seconds())
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, secs)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	} else {
		return fmt.Sprintf("%ds", secs)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(startTime)
	response := HealthResponse{
		Status:        "UP",
		Version:       VERSION,
		Uptime:        formatUptime(uptime),
		UptimeSeconds: int64(uptime.Seconds()),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(startTime)
	response := HealthResponse{
		Status:        "READY",
		Version:       VERSION,
		Uptime:        formatUptime(uptime),
		UptimeSeconds: int64(uptime.Seconds()),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func liveHandler(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(startTime)
	response := HealthResponse{
		Status:        "LIVE",
		Version:       VERSION,
		Uptime:        formatUptime(uptime),
		UptimeSeconds: int64(uptime.Seconds()),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func startHealthServer(port string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/health/ready", readyHandler)
	mux.HandleFunc("/health/live", liveHandler)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Error starting health server: %v", err)
		}
	}()

	return server
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Cargar configuración desde env
	cfg := config.LoadFromEnv()

	// 2. Iniciar logger
	z, err := logger.NewLogger()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer z.Sync()

	// 3. Verificar conexión a Kafka
	z.Info("checking kafka connectivity...", zap.Strings("brokers", cfg.KafkaBrokers))
	if err := checkKafkaConnectivity(cfg.KafkaBrokers); err != nil {
		z.Fatal("kafka not available", zap.Error(err))
	}
	z.Info("kafka connectivity confirmed")

	// 4. Crear producer para topic de salida (notificaciones)
	producerTopic := getEnv("KAFKA_PRODUCER_TOPIC", "notifications")
	producer := kafkaPkg.NewProducer(cfg.KafkaBrokers, producerTopic)

	// 5. Servicios y Handlers
	reg := handler.NewRegistry()
	userSvc := service.NewUserService(producer, z)

	// Cada handler interpreta un tipo de evento y llama al servicio
	reg.Register(handler.NewUserRegisteredHandler(userSvc, z))  // welcome
	reg.Register(handler.NewPasswordChangedHandler(userSvc, z)) // resetPassword
	reg.Register(handler.NewOtpRequestedHandler(userSvc, z))    // OTP
	reg.Register(handler.NewUserLoginHandler(userSvc, z))       // login_alert
	reg.Register(handler.NewUserVerifiedHandler(userSvc, z))    // verified_user

	// etc.

	proc := processor.NewProcessor(reg, z)

	// 6. Consumer - escucha el topic de entrada (user-events)
	rCfg := kafka.ReaderConfig{
		Brokers:  cfg.KafkaBrokers,
		Topic:    cfg.KafkaTopic, // ej: user-events
		GroupID:  cfg.GroupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	}
	consumer := kafkaPkg.NewConsumer(rCfg, proc, z)

	// 7. Iniciar servidor HTTP para health checks
	healthPort := getEnv("HEALTH_PORT", "8080")
	healthServer := startHealthServer(healthPort)
	z.Info("health server started", zap.String("port", healthPort))
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := healthServer.Shutdown(shutdownCtx); err != nil {
			z.Error("error shutting down health server", zap.Error(err))
		}
	}()

	// 8. Iniciar consumer
	go func() {
		defer func() {
			if r := recover(); r != nil {
				z.Error("consumer panic recovered", zap.Any("panic", r))
			}
		}()
		consumer.Start(ctx, 4) // 4 workers en paralelo
	}()

	// 9. Esperar señal para apagado
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	z.Info("shutdown requested")
	cancel()

	// 10. Cierre ordenado
	time.Sleep(3 * time.Second)
	_ = consumer.Close()
	_ = producer.Close()
	z.Info("bye")
}

// Verifica que Kafka esté disponible
func checkKafkaConnectivity(brokers []string) error {
	if len(brokers) == 0 {
		return fmt.Errorf("no kafka brokers configured")
	}

	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("failed to connect to broker %s: %w", brokers[0], err)
	}
	defer conn.Close()

	if _, err = conn.Brokers(); err != nil {
		return fmt.Errorf("failed to get brokers list: %w", err)
	}
	return nil
}

// Helper para valores por defecto
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
