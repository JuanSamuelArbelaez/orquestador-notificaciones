package service

import "context"

type Producer interface {
	Send(ctx context.Context, key []byte, value []byte) error
	SendEvent(ctx context.Context, eventType, template, to string, data map[string]interface{}) error
}
