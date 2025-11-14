// logger/logger.go
package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Logger estructura para logging en formato JSON
type Logger struct {
	loggerName string
}

// LogPayload estructura del log JSON (igual a Logger.js)
type LogPayload struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Logger    string                 `json:"logger"`
	Message   string                 `json:"message"`
	Thread    string                 `json:"thread"`
	Meta      map[string]interface{} `json:"-"`
}

// New crea una nueva instancia de logger
func New(loggerName string) *Logger {
	return &Logger{loggerName: loggerName}
}

// log función interna que genera el JSON
func (l *Logger) log(level, message string, meta map[string]interface{}) {
	payload := LogPayload{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Logger:    l.loggerName,
		Message:   message,
		Thread:    fmt.Sprintf("%d", os.Getpid()),
	}

	// Construir el JSON final con meta incluido
	result := map[string]interface{}{
		"timestamp": payload.Timestamp,
		"level":     payload.Level,
		"logger":    payload.Logger,
		"message":   payload.Message,
		"thread":    payload.Thread,
	}

	// Agregar campos meta
	for k, v := range meta {
		result[k] = v
	}

	// Serializar a JSON
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		// Fallback si hay error serializando
		fmt.Printf(`{"timestamp":"%s","level":"error","logger":"Logger","message":"Error serializando log: %s","thread":"%s"}`+"\n",
			time.Now().UTC().Format(time.RFC3339), err.Error(), fmt.Sprintf("%d", os.Getpid()))
		return
	}

	fmt.Println(string(jsonBytes))
}

// Info registra un log de nivel info
func (l *Logger) Info(message string, meta map[string]interface{}) {
	if meta == nil {
		meta = make(map[string]interface{})
	}
	l.log("info", message, meta)
}

// Debug registra un log de nivel debug
func (l *Logger) Debug(message string, meta map[string]interface{}) {
	if meta == nil {
		meta = make(map[string]interface{})
	}
	l.log("debug", message, meta)
}

// Warn registra un log de nivel warn
func (l *Logger) Warn(message string, meta map[string]interface{}) {
	if meta == nil {
		meta = make(map[string]interface{})
	}
	l.log("warn", message, meta)
}

// Error registra un log de nivel error
func (l *Logger) Error(message string, meta map[string]interface{}) {
	if meta == nil {
		meta = make(map[string]interface{})
	}
	l.log("error", message, meta)
}

// Fatal registra un log de nivel error y termina el programa
func (l *Logger) Fatal(message string, meta map[string]interface{}) {
	if meta == nil {
		meta = make(map[string]interface{})
	}
	l.log("error", message, meta)
	os.Exit(1)
}
