package gomappergen

import (
	"context"
	"log/slog"
	"strings"
)

func replacePlaceholders(template string, vars map[string]string) string {
	result := template
	for k, v := range vars {
		result = strings.ReplaceAll(result, k, v)
	}
	return result
}

func NewNoopLogger() *slog.Logger {
	return slog.New(&noopSlogHandler{})
}

type noopSlogHandler struct{}

func (n *noopSlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return false
}

func (n *noopSlogHandler) Handle(ctx context.Context, record slog.Record) error {
	return nil
}

func (n *noopSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return n
}

func (n *noopSlogHandler) WithGroup(name string) slog.Handler {
	return n
}

var _ slog.Handler = (*noopSlogHandler)(nil)
