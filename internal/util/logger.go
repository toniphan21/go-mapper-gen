package util

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"
)

type SlogHandler struct {
	writer *os.File
	level  slog.Level
}

func NewSlogHandler(writer *os.File, level string) slog.Handler {
	h := &SlogHandler{writer: writer, level: slog.LevelInfo}
	h.SetLevel(level)
	return h
}

func (h *SlogHandler) SetLevel(level string) {
	l := slog.LevelDebug
	switch level {
	case "debug":
		l = slog.LevelDebug
	case "info":
		l = slog.LevelInfo
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	}
	h.level = l
}

func (h *SlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *SlogHandler) Handle(_ context.Context, r slog.Record) error {
	showTime := false
	timeStr := r.Time.Format(time.RFC3339)
	levelStr := ""
	switch h.level {
	case slog.LevelError:
		levelStr = "ERR"
	case slog.LevelWarn:
		levelStr = "WRN"
	case slog.LevelInfo:
		levelStr = "INF"
	case slog.LevelDebug:
		showTime = true
		levelStr = "DBG"
	default:
		levelStr = r.Level.String()
	}
	msg := r.Message

	attrs := ""
	r.Attrs(func(a slog.Attr) bool {
		attrs += fmt.Sprintf(" %s=%v", a.Key, a.Value)
		return true
	})

	if showTime {
		_, err := fmt.Fprintf(h.writer, "[%s] %s: %s%s\n", timeStr, levelStr, msg, attrs)
		return err
	} else {
		_, err := fmt.Fprintf(h.writer, "%s%s\n", msg, attrs)
		return err
	}
}

func (h *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *SlogHandler) WithGroup(name string) slog.Handler {
	return h
}
