package handler

import (
	"context"

	"github.com/andrew/orquestador-notificacion/internal/domain"
	"github.com/andrew/orquestador-notificacion/internal/logger"
	"github.com/andrew/orquestador-notificacion/internal/service"
)

type PasswordChangedPayload struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
	Name  string `json:"name"`
	ID    int    `json:"id"`
}

type PasswordChangedHandler struct {
	userSvc service.UserService
	logger  *logger.Logger
}

func NewPasswordChangedHandler(us service.UserService, log *logger.Logger) *PasswordChangedHandler {
	return &PasswordChangedHandler{userSvc: us, logger: log}
}

func (h *PasswordChangedHandler) Types() []string {
	return []string{"PASSWORD_CHANGED"}
}

func (h *PasswordChangedHandler) Handle(ctx context.Context, e *domain.Event) error {
	var p PasswordChangedPayload
	if err := e.DecodePayload(&p); err != nil {
		h.logger.Error("Error al decodificar payload de PASSWORD_CHANGED", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Enviar alerta por email
	if err := h.userSvc.SendNotification(ctx, p.ID, p.Email, p.Name, p.Phone, "EMAIL", "password_changed_alert"); err != nil {
		h.logger.Error("Fallo al enviar alerta de cambio de contraseña por email", map[string]interface{}{
			"error":   err.Error(),
			"user_id": p.ID,
			"email":   p.Email,
		})
		return err
	}

	// Enviar alerta por sms
	if err := h.userSvc.SendNotification(ctx, p.ID, p.Email, p.Name, p.Phone, "SMS", "password_changed_alert"); err != nil {
		h.logger.Error("Fallo al enviar alerta de cambio de contraseña por SMS", map[string]interface{}{
			"error":   err.Error(),
			"user_id": p.ID,
			"phone":   p.Phone,
		})
		return err
	}

	h.logger.Info("Evento PASSWORD_CHANGED procesado exitosamente", map[string]interface{}{
		"user_id": p.ID,
		"email":   p.Email,
	})
	return nil
}
