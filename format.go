package slogtelegram

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"text/template"
	"time"
)

const defaultTemplate = `{{level .}}
{{time .}}

{{message .}}{{if hasAttrs .}}
<tg-spoiler><blockquote expandable>{{attrs .}}</blockquote></tg-spoiler>{{end}}
`

type Formatter interface {
	Format(record slog.Record) (string, error)
}

type FormatterOptions struct {
	// Template is a Go template to format the log record.
	Template string

	// LevelFormatter is a custom function to format the log level.
	LevelFormatter func(slog.Record) string

	// TimeFormatter is a custom function to format the log time.
	TimeFormatter func(slog.Record) string

	// MessageFormatter is a custom function to format the log message.
	MessageFormatter func(slog.Record) string

	// AttrsFormatter is a custom function to format the log attributes.
	AttrsFormatter func(slog.Record) string

	// Instance is a custom formatter instance to use.
	Instance Formatter
}

type defaultFormatter struct {
	template *template.Template
}

func NewFormatter(opts FormatterOptions) Formatter {
	if opts.Instance != nil {
		return opts.Instance
	}

	f := &defaultFormatter{}

	if opts.Template == "" {
		opts.Template = defaultTemplate
	}

	if opts.LevelFormatter == nil {
		opts.LevelFormatter = f.formatLevel
	}

	if opts.TimeFormatter == nil {
		opts.TimeFormatter = f.formatTime
	}

	if opts.MessageFormatter == nil {
		opts.MessageFormatter = f.formatMessage
	}

	if opts.AttrsFormatter == nil {
		opts.AttrsFormatter = f.formatAttrs
	}

	f.template = template.Must(template.New("default").Funcs(template.FuncMap{
		"level":   opts.LevelFormatter,
		"time":    opts.TimeFormatter,
		"message": opts.MessageFormatter,
		"hasAttrs": func(record slog.Record) bool {
			return record.NumAttrs() > 0
		},
		"attrs": opts.AttrsFormatter,
	}).Parse(opts.Template))

	return f
}

func (f *defaultFormatter) Format(record slog.Record) (string, error) {
	var res bytes.Buffer
	if err := f.template.Execute(&res, record); err != nil {
		return "", err
	}

	return res.String(), nil
}

func (f *defaultFormatter) formatLevel(record slog.Record) string {
	switch record.Level {
	case slog.LevelDebug:
		return "<i>🐞 Debug</i>"
	case slog.LevelInfo:
		return "<i>ℹ️ Info</i>"
	case slog.LevelWarn:
		return "<i>⚠️ Warning</i>"
	case slog.LevelError:
		return "<i>⛔ Error</i>"
	}

	return "<i>❔ Unknown</i>"
}

func (f *defaultFormatter) formatTime(record slog.Record) string {
	return fmt.Sprintf("<i>⌚️ %s</i>", record.Time.Format(time.RFC3339))
}

func (f *defaultFormatter) formatMessage(record slog.Record) string {
	return fmt.Sprintf("<b>💬 %s</b>", record.Message)
}

func (f *defaultFormatter) formatAttrs(record slog.Record) string {
	if record.NumAttrs() == 0 {
		return ""
	}

	var res strings.Builder

	record.Attrs(func(attr slog.Attr) bool {
		res.WriteString(f.formatAttr(attr) + "\n")

		return true
	})

	return res.String()
}

func (f *defaultFormatter) formatAttr(attr slog.Attr) string {
	return fmt.Sprintf("🔘 %s: %v", attr.Key, attr.Value)
}
