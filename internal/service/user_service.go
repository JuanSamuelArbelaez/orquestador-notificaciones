package service

import (
	"context"

	"github.com/andrew/orquestador-notificacion/internal/logger"
)

type UserService interface {
	OnUserRegistered(ctx context.Context, id int, email, name, phone, url string) error
	SendNotification(ctx context.Context, id int, email, name, phone, channel, template string) error
	SendOtpRecovery(ctx context.Context, id int, email, name, url string) error
	OnUserVerified(ctx context.Context, id int, email, name, phone string) error
}

type userServiceImpl struct {
	producer Producer // usa la interfaz, no la implementación concreta
	logger   *logger.Logger
}

func NewUserService(producer Producer, log *logger.Logger) UserService {
	return &userServiceImpl{producer: producer, logger: log}
}

func (s *userServiceImpl) OnUserRegistered(ctx context.Context, id int, email, name, phone, url string) error {
	data := map[string]interface{}{
		"user_id": id,
		"name":    name,
		"phone":   phone,
		"url":     url,
	}

	err := s.producer.SendEvent(ctx, "EMAIL", "welcome", email, data)
	if err != nil {
		s.logger.Error("Fallo al enviar notificación de bienvenida", map[string]interface{}{
			"error": err.Error(),
			"user_id": id,
			"email": email,
		})
		return err
	}

	s.logger.Info("Evento de notificación de bienvenida publicado", map[string]interface{}{
		"type":     "EMAIL",
		"template": "welcome",
		"to":       email,
		"user_id":  id,
	})
	return nil
}

func (s *userServiceImpl) SendNotification(ctx context.Context, id int, email, name, phone, channel, template string) error {
	to := chooseTarget(channel, email, phone)
	data := map[string]interface{}{
		"user_id": id,
		"name":    name,
		"phone":   phone,
	}

	err := s.producer.SendEvent(ctx, channel, template, to, data)
	if err != nil {
		s.logger.Error("Fallo al enviar notificación", map[string]interface{}{
			"error":    err.Error(),
			"channel":  channel,
			"template": template,
			"user_id":  id,
		})
		return err
	}

	s.logger.Info("Notificación enviada exitosamente", map[string]interface{}{
		"channel":  channel,
		"template": template,
		"to":       to,
		"user_id":  id,
	})
	return nil
}

func chooseTarget(channel, email, phone string) string {
	if channel == "EMAIL" {
		return email
	}
	return phone
}

func (s *userServiceImpl) SendOtpRecovery(ctx context.Context, id int, email, name, url string) error {
	data := map[string]interface{}{
		"user_id": id,
		"name":    name,
		"url":     url,
	}

	err := s.producer.SendEvent(ctx, "EMAIL", "password_recovery", email, data)
	if err != nil {
		s.logger.Error("Fallo al enviar notificación de recuperación de contraseña", map[string]interface{}{
			"error":   err.Error(),
			"user_id": id,
			"email":   email,
		})
		return err
	}

	s.logger.Info("Notificación de recuperación de contraseña enviada", map[string]interface{}{
		"to":      email,
		"user_id": id,
	})
	return nil
}

func (s *userServiceImpl) OnUserVerified(ctx context.Context, id int, email, name, phone string) error {
	data := map[string]interface{}{
		"user_id": id,
		"name":    name,
		"phone":   phone,
	}

	err := s.producer.SendEvent(ctx, "EMAIL", "account_verified", email, data)
	if err != nil {
		s.logger.Error("Fallo al enviar notificación de cuenta verificada", map[string]interface{}{
			"error":   err.Error(),
			"user_id": id,
			"email":   email,
		})
		return err
	}

	s.logger.Info("Notificación de cuenta verificada enviada", map[string]interface{}{
		"type":     "EMAIL",
		"template": "account_verified",
		"to":       email,
		"user_id":  id,
	})
	return nil
}
