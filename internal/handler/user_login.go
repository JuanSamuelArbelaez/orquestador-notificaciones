package handler

import (
	"context"

	"github.com/andrew/orquestador-notificacion/internal/domain"
	"github.com/andrew/orquestador-notificacion/internal/logger"
	"github.com/andrew/orquestador-notificacion/internal/service"
)

type UserLoginPayload struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
	Name  string `json:"name"`
	ID    int    `json:"id"`
}

type UserLoginHandler struct {
	userSvc service.UserService
	logger  *logger.Logger
}

func NewUserLoginHandler(us service.UserService, log *logger.Logger) *UserLoginHandler {
	return &UserLoginHandler{userSvc: us, logger: log}
}

func (h *UserLoginHandler) Types() []string {
	return []string{"USER_LOGIN"}
}

func (h *UserLoginHandler) Handle(ctx context.Context, e *domain.Event) error {
	var p UserLoginPayload
	if err := e.DecodePayload(&p); err != nil {
		h.logger.Error("Error al decodificar payload de USER_LOGIN", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Notificación por EMAIL
	if err := h.userSvc.SendNotification(ctx, p.ID, p.Email, p.Name, p.Phone, "EMAIL", "login_alert"); err != nil {
		h.logger.Error("Fallo al enviar notificación de login por email", map[string]interface{}{
			"error":   err.Error(),
			"user_id": p.ID,
			"email":   p.Email,
		})
		return err
	}

	// Notificación por SMS
	if err := h.userSvc.SendNotification(ctx, p.ID, p.Email, p.Name, p.Phone, "SMS", "login_alert"); err != nil {
		h.logger.Error("Fallo al enviar notificación de login por SMS", map[string]interface{}{
			"error":   err.Error(),
			"user_id": p.ID,
			"phone":   p.Phone,
		})
		return err
	}

	h.logger.Info("Evento USER_LOGIN procesado exitosamente", map[string]interface{}{
		"user_id": p.ID,
		"email":   p.Email,
		"phone":   p.Phone,
	})
	return nil
}
