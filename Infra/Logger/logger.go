package logging

import (
	"io"
	"log"
)

type Logger interface {
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type DefaultLogger struct {
	logger *log.Logger
}

func New(w io.Writer) DefaultLogger {
	return DefaultLogger{
		log.New(w, "", log.Ltime),
	}
}

func (l DefaultLogger) Info(msg string, args ...any) {
	l.logger.Printf("[INFO] "+msg, args...)
}

func (l DefaultLogger) Debug(msg string, args ...any) {
	l.logger.Printf("[DEBUG] "+msg, args...)
}

func (l DefaultLogger) Warn(msg string, args ...any) {
	l.logger.Printf("[WARN] "+msg, args...)
}

func (l DefaultLogger) Error(msg string, args ...any) {
	l.logger.Printf("[ERROR] "+msg, args...)
}
