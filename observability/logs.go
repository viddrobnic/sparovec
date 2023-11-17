package observability

import (
	"context"
	"errors"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/viddrobnic/sparovec/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}

	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	errs := make([]error, 0, len(m.handlers))
	for _, h := range m.handlers {
		if err := h.Handle(ctx, r); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	hanlders := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		hanlders[i] = h.WithAttrs(attrs)
	}

	return &multiHandler{hanlders}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	hanlders := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		hanlders[i] = h.WithGroup(name)
	}

	return &multiHandler{hanlders}
}

func newRollingFile(conf *config.Config) io.Writer {
	directory := filepath.Dir(conf.Observability.Path)
	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return &lumberjack.Logger{
		Filename:   conf.Observability.Path,
		MaxSize:    conf.Observability.MaxSize,
		MaxBackups: conf.Observability.MaxBackups,
		Compress:   true,
	}
}

func NewLogger(conf *config.Config) *slog.Logger {
	handler := &multiHandler{
		handlers: []slog.Handler{},
	}

	if conf.Observability.WriteToConsole {
		handler.handlers = append(handler.handlers, slog.Default().Handler())
	}

	if conf.Observability.WriteToFile {
		writer := newRollingFile(conf)
		handler.handlers = append(handler.handlers, slog.NewJSONHandler(writer, nil))
	}

	return slog.New(handler)
}
