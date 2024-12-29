package slogtelegram

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

const errPrefix = "slogtelegram: "

var _ slog.Handler = (*Handler)(nil)

type HandlerOptions struct {
	// Level reports the minimum record level that will be logged.
	// The handler discards records with lower levels.
	// If Level is nil, the handler assumes LevelInfo.
	// The handler calls Level.Level for each record processed;
	// to adjust the minimum level dynamically, use a LevelVar.
	Level slog.Leveler

	// Formatter is a config to format the log record before sending.
	Formatter FormatterOptions

	// Sender is a config to send the formatted log record to a telegram chat.
	Sender SenderOptions
}

type Handler struct {
	formatter Formatter
	sender    Sender
	minLevel  slog.Leveler
	attrs     []slog.Attr
	groups    []string
}

func DefaultHandler(token string, chatID int64) *Handler {
	return NewHandler(HandlerOptions{
		Sender: SenderOptions{
			Token:  token,
			ChatID: chatID,
		},
	})
}

func NewHandler(opts HandlerOptions) *Handler {
	if opts.Level == nil {
		opts.Level = slog.LevelInfo
	}

	return &Handler{
		formatter: NewFormatter(opts.Formatter),
		sender:    NewSender(opts.Sender),
		minLevel:  opts.Level,
	}
}

func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.minLevel.Level()
}

func (h *Handler) Handle(_ context.Context, record slog.Record) error {
	record.AddAttrs(h.attrs...)
	if len(h.groups) > 0 {
		record.AddAttrs(slog.String("groups", strings.Join(h.groups, ", ")))
	}

	text, err := h.formatter.Format(record)
	if err != nil {
		return fmt.Errorf("%s%w", errPrefix, err)
	}

	if err := h.sender.Send(text); err != nil {
		return fmt.Errorf("%s%w", errPrefix, err)
	}

	return nil
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	c := h.clone()
	c.attrs = append(c.attrs, attrs...)

	return c
}

func (h *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	c := h.clone()
	c.groups = append(h.groups, name)

	return c
}

func (h *Handler) Close() error {
	return h.sender.Close()
}

func (h *Handler) clone() *Handler {
	c := *h
	c.attrs = append([]slog.Attr(nil), h.attrs...)
	c.groups = append([]string(nil), h.groups...)

	return &c
}
