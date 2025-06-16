package silentlog

import (
	"context"
	"log/slog"
)

// SilentHandler — хендлер, который игнорирует все логи
type SilentHandler struct{}

func (h SilentHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return false // Никогда не логируем
}

func (h SilentHandler) Handle(_ context.Context, _ slog.Record) error {
	return nil // Ничего не делаем
}

func (h SilentHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h // Возвращаем тот же хендлер
}

func (h SilentHandler) WithGroup(_ string) slog.Handler {
	return h // Возвращаем тот же хендлер
}

func NewSilentLogger() *slog.Logger {
	return slog.New(SilentHandler{})
}
