package config

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"github.com/andrew/orquestador-notificacion/internal/logger"
)

type Config struct {
	KafkaBrokers []string
	KafkaTopic   string
	GroupID      string
}

// LoadFromEnv carga configuración desde variables de entorno y usa logger estructurado
func LoadFromEnv(serviceName string) Config {
	log, err := logger.New(serviceName)
	if err != nil {
		panic("❌ No se pudo inicializar el logger: " + err.Error())
	}

	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:29092"
		log.Warn("KAFKA_BROKERS no definido, usando valor por defecto",
			zap.String("default", brokers))
	}

	topic := os.Getenv("KAFKA_CONSUMER_TOPIC")
	if topic == "" {
		topic = "user-events"
		log.Warn("KAFKA_CONSUMER_TOPIC no definido, usando valor por defecto",
			zap.String("default", topic))
	}

	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "kafka-listener-group"
		log.Warn("KAFKA_GROUP_ID no definido, usando valor por defecto",
			zap.String("default", groupID))
	}

	config := Config{
		KafkaBrokers: strings.Split(brokers, ","),
		KafkaTopic:   topic,
		GroupID:      groupID,
	}

	log.Info("Configuración de Kafka cargada exitosamente", zap.Any("config", config))

	return config
}
