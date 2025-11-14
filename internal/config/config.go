package config

import (
	"os"
	"strings"

	"github.com/andrew/orquestador-notificacion/internal/logger"
)

type Config struct {
	KafkaBrokers []string
	KafkaTopic   string
	GroupID      string
}

// LoadFromEnv carga configuración desde variables de entorno y usa logger estructurado
func LoadFromEnv(loggerName string, log *logger.Logger) Config {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:29092"
		log.Warn("KAFKA_BROKERS no definido, usando valor por defecto", map[string]interface{}{
			"default": brokers,
		})
	}

	topic := os.Getenv("KAFKA_CONSUMER_TOPIC")
	if topic == "" {
		topic = "user-events"
		log.Warn("KAFKA_CONSUMER_TOPIC no definido, usando valor por defecto", map[string]interface{}{
			"default": topic,
		})
	}

	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "kafka-listener-group"
		log.Warn("KAFKA_GROUP_ID no definido, usando valor por defecto", map[string]interface{}{
			"default": groupID,
		})
	}

	config := Config{
		KafkaBrokers: strings.Split(brokers, ","),
		KafkaTopic:   topic,
		GroupID:      groupID,
	}

	log.Info("Configuración de Kafka cargada exitosamente", map[string]interface{}{
		"brokers": config.KafkaBrokers,
		"topic":   config.KafkaTopic,
		"groupID": config.GroupID,
	})

	return config
}
