package handler

import (
	"context"

	"github.com/andrew/orquestador-notificacion/internal/domain"
	"github.com/andrew/orquestador-notificacion/internal/logger"
	"github.com/andrew/orquestador-notificacion/internal/service"
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
	logger  *logger.Logger
}

func NewOtpRequestedHandler(us service.UserService, log *logger.Logger) *OtpRequestedHandler {
	return &OtpRequestedHandler{userSvc: us, logger: log}
}

func (h *OtpRequestedHandler) Types() []string {
	return []string{"OTP_REQUESTED"}
}

func (h *OtpRequestedHandler) Handle(ctx context.Context, e *domain.Event) error {
	var p OtpRequestedPayload
	if err := e.DecodePayload(&p); err != nil {
		h.logger.Error("Error al decodificar payload de OTP_REQUESTED", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Usar el método especializado de OTP
	if err := h.userSvc.SendOtpRecovery(ctx, p.ID, p.Email, p.Name, p.Url); err != nil {
		h.logger.Error("Fallo al enviar email de recuperación de contraseña con OTP", map[string]interface{}{
			"error":   err.Error(),
			"user_id": p.ID,
			"email":   p.Email,
		})
		return err
	}

	h.logger.Info("Evento OTP_REQUESTED procesado exitosamente", map[string]interface{}{
		"user_id": p.ID,
		"email":   p.Email,
		"url":     p.Url,
	})

	return nil
}
