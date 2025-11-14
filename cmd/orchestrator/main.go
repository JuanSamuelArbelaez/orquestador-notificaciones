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

	// 1. Iniciar logger
	log := logger.New("[OrchestratorMain]")

	// 2. Cargar configuración desde env
	cfg := config.LoadFromEnv("[Config]", log)

	// 3. Verificar conexión a Kafka
	log.Info("Verificando conectividad con Kafka...", map[string]interface{}{
		"brokers": cfg.KafkaBrokers,
	})
	if err := checkKafkaConnectivity(cfg.KafkaBrokers); err != nil {
		log.Fatal("Kafka no disponible", map[string]interface{}{
			"error": err.Error(),
		})
	}
	log.Info("Conectividad con Kafka confirmada", nil)

	// 4. Crear producer para topic de salida (notificaciones)
	producerTopic := getEnv("KAFKA_PRODUCER_TOPIC", "notifications")
	producer := kafkaPkg.NewProducer(cfg.KafkaBrokers, producerTopic)
	log.Info("Producer de Kafka inicializado", map[string]interface{}{
		"topic": producerTopic,
	})

	// 5. Servicios y Handlers
	reg := handler.NewRegistry()
	userSvc := service.NewUserService(producer, log)

	// Cada handler interpreta un tipo de evento y llama al servicio
	reg.Register(handler.NewUserRegisteredHandler(userSvc, log))  // welcome
	reg.Register(handler.NewPasswordChangedHandler(userSvc, log)) // resetPassword
	reg.Register(handler.NewOtpRequestedHandler(userSvc, log))    // OTP
	reg.Register(handler.NewUserLoginHandler(userSvc, log))       // login_alert
	reg.Register(handler.NewUserVerifiedHandler(userSvc, log))    // verified_user

	log.Info("Handlers registrados exitosamente", map[string]interface{}{
		"handlers": []string{"USER_REGISTERED", "PASSWORD_CHANGED", "OTP_REQUESTED", "USER_LOGIN", "USER_VERIFIED"},
	})

	proc := processor.NewProcessor(reg, log)

	// 6. Consumer - escucha el topic de entrada (user-events)
	rCfg := kafka.ReaderConfig{
		Brokers:  cfg.KafkaBrokers,
		Topic:    cfg.KafkaTopic, // ej: user-events
		GroupID:  cfg.GroupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	}
	consumer := kafkaPkg.NewConsumer(rCfg, proc, log)

	log.Info("Consumer de Kafka configurado", map[string]interface{}{
		"topic":   cfg.KafkaTopic,
		"groupID": cfg.GroupID,
	})

	// 7. Iniciar servidor HTTP para health checks
	healthPort := getEnv("HEALTH_PORT", "8080")
	healthServer := startHealthServer(healthPort)
	log.Info("Servidor de health checks iniciado", map[string]interface{}{
		"port": healthPort,
	})
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := healthServer.Shutdown(shutdownCtx); err != nil {
			log.Error("Error al cerrar servidor de health checks", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	// 8. Iniciar consumer (después del health server)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("Panic recuperado en consumer", map[string]interface{}{
					"panic": fmt.Sprintf("%v", r),
				})
			}
		}()
		log.Info("Iniciando consumer con workers", map[string]interface{}{
			"workers": 4,
		})
		consumer.Start(ctx, 4) // 4 workers en paralelo
	}()

	// 8. Esperar señal para apagado
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	log.Info("Orquestador iniciado, esperando eventos...", nil)
	<-sig
	log.Info("Solicitud de apagado recibida", nil)
	cancel()

	// 9. Cierre ordenado
	log.Info("Iniciando cierre ordenado...", nil)
	time.Sleep(3 * time.Second)
	_ = consumer.Close()
	_ = producer.Close()
	log.Info("Orquestador finalizado correctamente", nil)
}

// Verifica que Kafka esté disponible
func checkKafkaConnectivity(brokers []string) error {
	if len(brokers) == 0 {
		return fmt.Errorf("no hay brokers de Kafka configurados")
	}

	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("fallo al conectar con broker %s: %w", brokers[0], err)
	}
	defer conn.Close()

	if _, err = conn.Brokers(); err != nil {
		return fmt.Errorf("fallo al obtener lista de brokers: %w", err)
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
