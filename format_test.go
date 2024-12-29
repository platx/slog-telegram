package slogtelegram

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDefaultFormatter_Format(t *testing.T) {
	formatter := NewFormatter(FormatterOptions{})

	record := slog.Record{
		Level:   slog.LevelInfo,
		Time:    time.Date(2024, 12, 29, 15, 0, 0, 0, time.UTC),
		Message: "Test message",
	}

	record.AddAttrs(slog.String("key1", "value1"))
	record.AddAttrs(slog.Int("key2", 42))

	formatted, err := formatter.Format(record)
	assert.NoError(t, err)

	expected := `<i>ℹ️ Info</i>
<i>⌚️ 2024-12-29T15:00:00Z</i>

<b>💬 Test message</b>
<tg-spoiler><blockquote expandable>🔘 key1: value1
🔘 key2: 42
</blockquote></tg-spoiler>`
	assert.Equal(t, expected, strings.TrimSpace(formatted))
}

func TestCustomTemplate(t *testing.T) {
	customTemplate := `{{time .}} | {{level .}} | {{message .}}`
	formatter := NewFormatter(FormatterOptions{
		Template: customTemplate,
	})

	record := slog.Record{
		Level:   slog.LevelWarn,
		Time:    time.Date(2024, 12, 29, 15, 0, 0, 0, time.UTC),
		Message: "Warning message",
	}

	formatted, err := formatter.Format(record)
	assert.NoError(t, err)

	expected := `<i>⌚️ 2024-12-29T15:00:00Z</i> | <i>⚠️ Warning</i> | <b>💬 Warning message</b>`
	assert.Equal(t, expected, strings.TrimSpace(formatted))
}

func TestCustomFormatters(t *testing.T) {
	formatter := NewFormatter(FormatterOptions{
		LevelFormatter: func(record slog.Record) string {
			return "CustomLevel"
		},
		TimeFormatter: func(record slog.Record) string {
			return "CustomTime"
		},
		MessageFormatter: func(record slog.Record) string {
			return "CustomMessage"
		},
		AttrsFormatter: func(record slog.Record) string {
			return "CustomAttrs"
		},
	})

	record := slog.Record{
		Level:   slog.LevelDebug,
		Time:    time.Now(),
		Message: "Debug message",
	}
	record.AddAttrs(slog.String("key", "value"))
	formatted, err := formatter.Format(record)
	assert.NoError(t, err)

	expected := `CustomLevel
CustomTime

CustomMessage
<tg-spoiler><blockquote expandable>CustomAttrs</blockquote></tg-spoiler>`
	assert.Equal(t, expected, strings.TrimSpace(formatted))
}

func TestEmptyAttrs(t *testing.T) {
	formatter := NewFormatter(FormatterOptions{})

	record := slog.Record{
		Level:   slog.LevelInfo,
		Time:    time.Date(2024, 12, 29, 15, 0, 0, 0, time.UTC),
		Message: "Message with no attrs",
	}

	formatted, err := formatter.Format(record)
	assert.NoError(t, err)

	expected := `<i>ℹ️ Info</i>
<i>⌚️ 2024-12-29T15:00:00Z</i>

<b>💬 Message with no attrs</b>`
	assert.Equal(t, expected, strings.TrimSpace(formatted))
}

func TestCustomInstance(t *testing.T) {
	customFormatter := &formatterMock{}
	formatter := NewFormatter(FormatterOptions{
		Instance: customFormatter,
	})

	record := slog.Record{}

	customFormatter.On("Format", record).Return("custom", nil)

	_, err := formatter.Format(record)
	assert.NoError(t, err)

	customFormatter.AssertExpectations(t)
}

type formatterMock struct {
	mock.Mock
}

func (m *formatterMock) Format(record slog.Record) (string, error) {
	args := m.Called(record)

	return args.String(0), args.Error(1)
}
