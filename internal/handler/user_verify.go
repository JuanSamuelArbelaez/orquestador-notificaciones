package handler

import (
	"context"

	"github.com/andrew/orquestador-notificacion/internal/domain"
	"github.com/andrew/orquestador-notificacion/internal/logger"
	"github.com/andrew/orquestador-notificacion/internal/service"
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
	logger  *logger.Logger
}

// Constructor
func NewUserVerifiedHandler(us service.UserService, log *logger.Logger) *UserVerifiedHandler {
	return &UserVerifiedHandler{userSvc: us, logger: log}
}

// Tipos de eventos que maneja este handler
func (h *UserVerifiedHandler) Types() []string {
	return []string{"USER_VERIFIED"}
}

// Lógica de procesamiento del evento
func (h *UserVerifiedHandler) Handle(ctx context.Context, e *domain.Event) error {
	var p UserVerifiedPayload
	if err := e.DecodePayload(&p); err != nil {
		h.logger.Error("Error al decodificar payload de USER_VERIFIED", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Delegamos al UserService (por ejemplo, enviar email de cuenta verificada)
	if err := h.userSvc.OnUserVerified(ctx, p.ID, p.Email, p.Name, p.Phone); err != nil {
		h.logger.Error("Fallo en servicio de usuario para USER_VERIFIED", map[string]interface{}{
			"error":   err.Error(),
			"user_id": p.ID,
			"email":   p.Email,
		})
		return err
	}

	h.logger.Info("Evento USER_VERIFIED procesado exitosamente", map[string]interface{}{
		"user_id": p.ID,
		"email":   p.Email,
	})

	return nil
}
