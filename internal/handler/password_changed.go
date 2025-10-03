package handler

import (
	"context"

	"github.com/andrew/orquestador-notificacion/internal/domain"
	"github.com/andrew/orquestador-notificacion/internal/service"
	"go.uber.org/zap"
)

type PasswordChangedPayload struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
	Name  string `json:"name"`
	ID    int    `json:"id"`
}

type PasswordChangedHandler struct {
	userSvc service.UserService
	logger  *zap.Logger
}

func NewPasswordChangedHandler(us service.UserService, logger *zap.Logger) *PasswordChangedHandler {
	return &PasswordChangedHandler{userSvc: us, logger: logger}
}

func (h *PasswordChangedHandler) Types() []string {
	return []string{"PASSWORD_CHANGED"}
}

func (h *PasswordChangedHandler) Handle(ctx context.Context, e *domain.Event) error {
	var p PasswordChangedPayload
	if err := e.DecodePayload(&p); err != nil {
		return err
	}

	// Enviar alerta por email
	if err := h.userSvc.SendNotification(ctx, p.ID, p.Email, p.Name, p.Phone, "EMAIL", "password_changed_alert"); err != nil {
		h.logger.Error("failed to send password-changed email alert", zap.Error(err))
		return err
	}

	// Enviar alerta por sms
	if err := h.userSvc.SendNotification(ctx, p.ID, p.Email, p.Name, p.Phone, "SMS", "password_changed_alert"); err != nil {
		h.logger.Error("failed to send password-changed sms alert", zap.Error(err))
		return err
	}

	h.logger.Info("processed PASSWORD_CHANGED", zap.Int("user_id", p.ID), zap.String("email", p.Email))
	return nil
}
