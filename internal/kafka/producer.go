package kafka

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) Send(ctx context.Context, key []byte, value []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

// -------------------- NUEVO --------------------

// NotificationEvent es el contrato de eventos
type NotificationEvent struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Template string                 `json:"template"`
	To       string                 `json:"to"`
	Data     map[string]interface{} `json:"data"`
}

// SendEvent construye el JSON y lo manda
func (p *Producer) SendEvent(ctx context.Context, eventType, template, to string, data map[string]interface{}) error {
	event := NotificationEvent{
		ID:       uuid.New().String(),
		Type:     eventType,
		Template: template,
		To:       to,
		Data:     data,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.Send(ctx, []byte(event.ID), payload)
}
