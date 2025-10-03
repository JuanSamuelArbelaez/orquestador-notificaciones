package config

import (
	"os"
	"strings"
)

type Config struct {
	KafkaBrokers []string
	KafkaTopic   string
	GroupID      string
}

func LoadFromEnv() Config {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:29092"
	}

	topic := os.Getenv("KAFKA_CONSUMER_TOPIC")
	if topic == "" {
		topic = "user-events"
	}

	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "kafka-listener-group"
	}

	return Config{
		KafkaBrokers: strings.Split(brokers, ","),
		KafkaTopic:   topic,
		GroupID:      groupID,
	}
}
