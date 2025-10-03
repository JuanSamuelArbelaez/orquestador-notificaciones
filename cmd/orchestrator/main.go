package main

import (
	"context"
	"fmt"
	"log"
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

	// 7. Iniciar consumer
	go func() {
		defer func() {
			if r := recover(); r != nil {
				z.Error("consumer panic recovered", zap.Any("panic", r))
			}
		}()
		consumer.Start(ctx, 4) // 4 workers en paralelo
	}()

	// 8. Esperar señal para apagado
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	z.Info("shutdown requested")
	cancel()

	// 9. Cierre ordenado
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
