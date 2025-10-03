package processor

import (
	"context"
	"github.com/andrew/orquestador-notificacion/internal/domain"
	"github.com/andrew/orquestador-notificacion/internal/handler"
	"go.uber.org/zap"
)

type Processor struct {
	registry *handler.Registry
	logger   *zap.Logger
}

func NewProcessor(reg *handler.Registry, logger *zap.Logger) *Processor {
	return &Processor{registry: reg, logger: logger}
}

func (p *Processor) Process(ctx context.Context, e *domain.Event) error {
	hs, err := p.registry.GetHandlers(e.Type)
	if err != nil {
		p.logger.Warn("no handlers found for event type", zap.String("type", e.Type))
		return nil // opcional: no es error si no hay handler; depende de tu política
	}

	// Llamar handlers en secuencia (podrías paralelizar si son independientes)
	for _, h := range hs {
		if err := h.Handle(ctx, e); err != nil {
			p.logger.Error("handler error", zap.Error(err))
			return err
		}
	}

	return nil
}
