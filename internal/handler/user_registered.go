package handler

import (
	"context"
	_ "fmt"

	"github.com/andrew/orquestador-notificacion/internal/domain"
	"github.com/andrew/orquestador-notificacion/internal/service"
	"go.uber.org/zap"
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
	logger  *zap.Logger
}

func NewUserRegisteredHandler(us service.UserService, logger *zap.Logger) *UserRegisteredHandler {
	return &UserRegisteredHandler{userSvc: us, logger: logger}
}

func (h *UserRegisteredHandler) Types() []string {
	return []string{"USER_REGISTERED"}
}

func (h *UserRegisteredHandler) Handle(ctx context.Context, e *domain.Event) error {
	var p UserRegisteredPayload
	if err := e.DecodePayload(&p); err != nil {
		return err
	}

	// Delegar la lógica de negocio al service (Single Responsibility)
	if err := h.userSvc.OnUserRegistered(ctx, p.ID, p.Email, p.Name, p.Phone, p.Url); err != nil {
		h.logger.Error("user service failed", zap.Error(err))
		return err
	}

	h.logger.Info("processed USER_REGISTERED",
		zap.Int("user_id", p.ID),
		zap.String("email", p.Email),
		zap.String("url", p.Url))
	return nil
}
