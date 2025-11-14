package processor

import (
	"context"
	"github.com/andrew/orquestador-notificacion/internal/domain"
	"github.com/andrew/orquestador-notificacion/internal/handler"
	"github.com/andrew/orquestador-notificacion/internal/logger"
)

type Processor struct {
	registry *handler.Registry
	logger   *logger.Logger
}

func NewProcessor(reg *handler.Registry, log *logger.Logger) *Processor {
	return &Processor{registry: reg, logger: log}
}

func (p *Processor) Process(ctx context.Context, e *domain.Event) error {
	hs, err := p.registry.GetHandlers(e.Type)
	if err != nil {
		p.logger.Warn("No se encontraron handlers para el tipo de evento", map[string]interface{}{
			"type": e.Type,
		})
		return nil // opcional: no es error si no hay handler; depende de tu política
	}

	// Llamar handlers en secuencia (podrías paralelizar si son independientes)
	for _, h := range hs {
		if err := h.Handle(ctx, e); err != nil {
			p.logger.Error("Error en handler", map[string]interface{}{
				"error":     err.Error(),
				"event_type": e.Type,
			})
			return err
		}
	}

	return nil
}
