package slogtelegram

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_Enabled(t *testing.T) {
	handler := NewHandler(HandlerOptions{
		Level: slog.LevelWarn,
		Sender: SenderOptions{
			Token:  "test-token",
			ChatID: 1,
		},
	})

	assert.False(t, handler.Enabled(context.Background(), slog.LevelInfo))
	assert.True(t, handler.Enabled(context.Background(), slog.LevelWarn))
	assert.True(t, handler.Enabled(context.Background(), slog.LevelError))
}

func TestHandler_Handle_Success(t *testing.T) {
	mockFormatter := &formatterMock{}
	mockFormatter.On("Format", mock.Anything).Return("formatted log", nil)

	mockSender := &senderMock{}
	mockSender.On("Send", mock.Anything).Return(nil)

	handler := &Handler{
		formatter: mockFormatter,
		sender:    mockSender,
		minLevel:  slog.LevelInfo,
	}

	record := slog.Record{
		Level:   slog.LevelInfo,
		Message: "Test message",
	}

	err := handler.Handle(context.Background(), record)
	assert.NoError(t, err)
	mockSender.AssertCalled(t, "Send", "formatted log")
}

func TestHandler_Handle_FormatError(t *testing.T) {
	mockFormatter := &formatterMock{}
	mockFormatter.On("Format", mock.Anything).Return("", assert.AnError)

	handler := &Handler{
		formatter: mockFormatter,
		sender:    &senderMock{},
		minLevel:  slog.LevelInfo,
	}

	record := slog.Record{
		Level:   slog.LevelInfo,
		Message: "Test message",
	}

	err := handler.Handle(context.Background(), record)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "slogtelegram: ")
}

func TestHandler_Handle_SendError(t *testing.T) {
	mockSender := &senderMock{}
	mockSender.On("Send", mock.Anything).Return(assert.AnError)

	mockFormatter := &formatterMock{}
	mockFormatter.On("Format", mock.Anything).Return("formatted log", nil)

	handler := &Handler{
		formatter: mockFormatter,
		sender:    mockSender,
		minLevel:  slog.LevelInfo,
	}

	record := slog.Record{
		Level:   slog.LevelInfo,
		Message: "Test message",
	}

	err := handler.Handle(context.Background(), record)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "slogtelegram: ")
	mockSender.AssertCalled(t, "Send", "formatted log")
}

func TestHandler_WithAttrs(t *testing.T) {
	handler := DefaultHandler("test-token", 1)
	newHandler := handler.WithAttrs([]slog.Attr{
		slog.Any("key1", "value1"),
		slog.Any("key2", 42),
	})

	assert.NotSame(t, handler, newHandler)
	assert.Len(t, newHandler.(*Handler).attrs, 2)
}

func TestHandler_WithGroup(t *testing.T) {
	handler := DefaultHandler("test-token", 1)
	newHandler := handler.WithGroup("group1").WithGroup("group2")

	assert.NotSame(t, handler, newHandler)
	assert.Equal(t, []string{"group1", "group2"}, newHandler.(*Handler).groups)
}

func TestHandler_Close(t *testing.T) {
	mockSender := &senderMock{}
	mockSender.On("Close").Return(nil)

	handler := &Handler{
		sender: mockSender,
	}

	err := handler.Close()
	assert.NoError(t, err)
	mockSender.AssertCalled(t, "Close")
}
