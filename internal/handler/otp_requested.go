package handler

import (
	"context"

	"github.com/andrew/orquestador-notificacion/internal/domain"
	"github.com/andrew/orquestador-notificacion/internal/service"
	"github.com/andrew/orquestador-notificacion/internal/logger"
	"go.uber.org/zap"
)

type OtpRequestedPayload struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
	Name  string `json:"name"`
	ID    int    `json:"id"`
	Url   string `json:"url-recovery"`
}

type OtpRequestedHandler struct {
	userSvc service.UserService
	logger  *zap.Logger
}

func NewOtpRequestedHandler(us service.UserService, logger *zap.Logger) *OtpRequestedHandler {
	return &OtpRequestedHandler{userSvc: us, logger: logger}
}

func (h *OtpRequestedHandler) Types() []string {
	return []string{"OTP_REQUESTED"}
}

func (h *OtpRequestedHandler) Handle(ctx context.Context, e *domain.Event) error {
	var p OtpRequestedPayload
	if err := e.DecodePayload(&p); err != nil {
		return err
	}

	// Usar el método especializado de OTP
	if err := h.userSvc.SendOtpRecovery(ctx, p.ID, p.Email, p.Name, p.Url); err != nil {
		h.logger.Error("failed to send otp recovery email", zap.Error(err))
		return err
	}

	h.logger.Info("processed OTP_REQUESTED",
		zap.Int("user_id", p.ID),
		zap.String("email", p.Email),
		zap.String("url", p.Url),
	)

	return nil
}
