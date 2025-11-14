package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/andrew/orquestador-notificacion/internal/domain"
	"github.com/andrew/orquestador-notificacion/internal/logger"
	"github.com/andrew/orquestador-notificacion/internal/processor"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader    *kafka.Reader
	processor *processor.Processor
	logger    *logger.Logger
	shutdown  chan struct{}
}

func NewConsumer(cfg kafka.ReaderConfig, p *processor.Processor, log *logger.Logger) *Consumer {
	// Configuración mejorada del Reader
	cfg.MaxWait = 10 * time.Second
	cfg.ReadBackoffMin = 100 * time.Millisecond
	cfg.ReadBackoffMax = 1 * time.Second
	cfg.HeartbeatInterval = 3 * time.Second
	cfg.CommitInterval = 0 // Commit manual para mejor control

	r := kafka.NewReader(cfg)
	return &Consumer{
		reader:    r,
		processor: p,
		logger:    log,
		shutdown:  make(chan struct{}),
	}
}

func (c *Consumer) Start(ctx context.Context, workers int) {
	for i := 0; i < workers; i++ {
		go c.worker(ctx, i)
	}
}

func (c *Consumer) worker(ctx context.Context, id int) {
	c.logger.Info("Iniciando worker de consumer de Kafka", map[string]interface{}{
		"worker_id": id,
	})

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Worker deteniéndose por cancelación de contexto", map[string]interface{}{
				"worker_id": id,
			})
			return
		case <-c.shutdown:
			c.logger.Info("Worker deteniéndose por señal de apagado", map[string]interface{}{
				"worker_id": id,
			})
			return
		default:
			c.processMessage(ctx, id)
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, workerID int) {
	// Usar un contexto con timeout para evitar bloqueos eternos
	msgCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	m, err := c.reader.FetchMessage(msgCtx)
	if err != nil {
		if isTransientError(err) {
			c.logger.Warn("Error transitorio, se reintentará", map[string]interface{}{
				"worker_id": workerID,
				"error":     err.Error(),
			})
			time.Sleep(2 * time.Second)
			return
		}

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "context canceled") {
			return
		}

		c.logger.Error("Fallo al obtener mensaje de Kafka", map[string]interface{}{
			"worker_id": workerID,
			"error":     err.Error(),
		})
		return
	}

	var e domain.Event
	if err := json.Unmarshal(m.Value, &e); err != nil {
		c.logger.Error("JSON de evento inválido", map[string]interface{}{
			"worker_id": workerID,
			"error":     err.Error(),
			"raw":       string(m.Value),
		})

		// Commit para evitar procesar repetidamente mensajes inválidos
		if err := c.reader.CommitMessages(ctx, m); err != nil {
			c.logger.Error("Fallo al hacer commit después de mensaje inválido", map[string]interface{}{
				"worker_id": workerID,
				"error":     err.Error(),
			})
		}
		return
	}

	c.logger.Info("Procesando evento", map[string]interface{}{
		"worker_id": workerID,
		"event_type": e.Type,
		"event_id":   e.ID,
	})

	// Procesar el evento
	if err := c.processor.Process(ctx, &e); err != nil {
		c.logger.Error("Fallo al procesar evento", map[string]interface{}{
			"worker_id": workerID,
			"error":      err.Error(),
			"event_type": e.Type,
			"event_id":   e.ID,
		})

		// No commit para reintentar más tarde
		time.Sleep(5 * time.Second)
		return
	}

	// Commit después de procesamiento exitoso
	if err := c.reader.CommitMessages(ctx, m); err != nil {
		c.logger.Error("Fallo al hacer commit del mensaje", map[string]interface{}{
			"worker_id": workerID,
			"error":     err.Error(),
		})
	} else {
		c.logger.Info("Mensaje confirmado exitosamente", map[string]interface{}{
			"worker_id": workerID,
			"event_id":  e.ID,
		})
	}
}

// isTransientError identifica errores transitorios que merecen reintento
func isTransientError(err error) bool {
	if err == nil {
		return false
	}

	errorMsg := err.Error()
	return strings.Contains(errorMsg, "multiple Read calls return no data or error") ||
		strings.Contains(errorMsg, "connection reset by peer") ||
		strings.Contains(errorMsg, "i/o timeout") ||
		strings.Contains(errorMsg, "broker not available") ||
		strings.Contains(errorMsg, "network error")
}

func (c *Consumer) Close() error {
	c.logger.Info("Cerrando consumer de Kafka", nil)
	close(c.shutdown)
	return c.reader.Close()
}
