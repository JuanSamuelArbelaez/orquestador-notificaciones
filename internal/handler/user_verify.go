package handler

import (
	"context"

	"github.com/andrew/orquestador-notificacion/internal/domain"
	"github.com/andrew/orquestador-notificacion/internal/service"
	"go.uber.org/zap"
)

// Payload específico para el evento USER_VERIFIED
type UserVerifiedPayload struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
	Name  string `json:"name"`
	ID    int    `json:"id"`
}

// Handler para USER_VERIFIED
type UserVerifiedHandler struct {
	userSvc service.UserService // inyección del servicio de negocio
	logger  *zap.Logger
}

// Constructor
func NewUserVerifiedHandler(us service.UserService, logger *zap.Logger) *UserVerifiedHandler {
	return &UserVerifiedHandler{userSvc: us, logger: logger}
}

// Tipos de eventos que maneja este handler
func (h *UserVerifiedHandler) Types() []string {
	return []string{"USER_VERIFIED"}
}

// Lógica de procesamiento del evento
func (h *UserVerifiedHandler) Handle(ctx context.Context, e *domain.Event) error {
	var p UserVerifiedPayload
	if err := e.DecodePayload(&p); err != nil {
		return err
	}

	// Delegamos al UserService (por ejemplo, enviar email de cuenta verificada)
	if err := h.userSvc.OnUserVerified(ctx, p.ID, p.Email, p.Name, p.Phone); err != nil {
		h.logger.Error("user service failed", zap.Error(err))
		return err
	}

	h.logger.Info("processed USER_VERIFIED",
		zap.Int("user_id", p.ID),
		zap.String("email", p.Email),
	)

	return nil
}
