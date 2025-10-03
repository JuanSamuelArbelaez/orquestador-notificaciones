package handler

import (
	"context"
	"errors"
	"github.com/andrew/orquestador-notificacion/internal/domain"
)

type EventHandler interface {
	// Handle realiza el procesamiento del evento
	Handle(ctx context.Context, e *domain.Event) error
	// Types retorna los tipos de evento que maneja (por ejemplo: ["USER_REGISTERED"])
	Types() []string
}

type Registry struct {
	handlers map[string][]EventHandler
}

func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string][]EventHandler)}
}

func (r *Registry) Register(h EventHandler) {
	for _, t := range h.Types() {
		r.handlers[t] = append(r.handlers[t], h)
	}
}

func (r *Registry) GetHandlers(eventType string) ([]EventHandler, error) {
	hs, ok := r.handlers[eventType]
	if !ok || len(hs) == 0 {
		return nil, errors.New("no handler registered for event type")
	}
	return hs, nil
}
