package handler

import (
	"context"

	"github.com/andrew/orquestador-notificacion/internal/domain"
	"github.com/andrew/orquestador-notificacion/internal/logger"
	"github.com/andrew/orquestador-notificacion/internal/service"
)

type UserRegisteredPayload struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
	Name  string `json:"name"`
	ID    int    `json:"id"`
	Url   string `json:"url"`
}

type UserRegisteredHandler struct {
	userSvc service.UserService // interfaz para lógica de negocio (inyección)
	logger  *logger.Logger
}

func NewUserRegisteredHandler(us service.UserService, log *logger.Logger) *UserRegisteredHandler {
	return &UserRegisteredHandler{userSvc: us, logger: log}
}

func (h *UserRegisteredHandler) Types() []string {
	return []string{"USER_REGISTERED"}
}

func (h *UserRegisteredHandler) Handle(ctx context.Context, e *domain.Event) error {
	var p UserRegisteredPayload
	if err := e.DecodePayload(&p); err != nil {
		h.logger.Error("Error al decodificar payload de USER_REGISTERED", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Delegar la lógica de negocio al service (Single Responsibility)
	if err := h.userSvc.OnUserRegistered(ctx, p.ID, p.Email, p.Name, p.Phone, p.Url); err != nil {
		h.logger.Error("Fallo en servicio de usuario para USER_REGISTERED", map[string]interface{}{
			"error":   err.Error(),
			"user_id": p.ID,
			"email":   p.Email,
		})
		return err
	}

	h.logger.Info("Evento USER_REGISTERED procesado exitosamente", map[string]interface{}{
		"user_id": p.ID,
		"email":   p.Email,
		"url":      p.Url,
	})
	return nil
}
