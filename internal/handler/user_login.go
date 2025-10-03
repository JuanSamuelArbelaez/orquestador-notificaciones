package handler

import (
	"context"
	"github.com/andrew/orquestador-notificacion/internal/domain"
	"github.com/andrew/orquestador-notificacion/internal/service"
	"go.uber.org/zap"
)

type UserLoginPayload struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
	Name  string `json:"name"`
	ID    int    `json:"id"`
}

type UserLoginHandler struct {
	userSvc service.UserService
	logger  *zap.Logger
}

func NewUserLoginHandler(us service.UserService, logger *zap.Logger) *UserLoginHandler {
	return &UserLoginHandler{userSvc: us, logger: logger}
}

func (h *UserLoginHandler) Types() []string {
	return []string{"USER_LOGIN"}
}

func (h *UserLoginHandler) Handle(ctx context.Context, e *domain.Event) error {
	var p UserLoginPayload
	if err := e.DecodePayload(&p); err != nil {
		return err
	}

	// Notificación por EMAIL
	if err := h.userSvc.SendNotification(ctx, p.ID, p.Email, p.Name, p.Phone, "EMAIL", "login_alert"); err != nil {
		h.logger.Error("failed to send login email notification", zap.Error(err))
		return err
	}

	// Notificación por SMS
	if err := h.userSvc.SendNotification(ctx, p.ID, p.Email, p.Name, p.Phone, "SMS", "login_alert"); err != nil {
		h.logger.Error("failed to send login sms notification", zap.Error(err))
		return err
	}

	h.logger.Info("processed USER_LOGIN", zap.Int("user_id", p.ID), zap.String("email", p.Email), zap.String("phone", p.Phone))
	return nil
}
