package logger

import (
	"context"
	"log"
)

type Logger struct{}

func New() *Logger {
	return &Logger{}
}

func (l *Logger) Info(ctx context.Context, msg string) {
	log.Println(ctx, "I", msg)
}

func (l *Logger) Warn(ctx context.Context, msg string) {
	log.Println(ctx, "W", msg)
}

func (l *Logger) Error(ctx context.Context, msg string) {
	log.Println(ctx, "E", msg)
}

func (l *Logger) Fatal(ctx context.Context, msg string) {
	log.Fatalln(ctx, "F", msg)
}
