package slogtelegram

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/telebot.v4"
)

func TestNewSender(t *testing.T) {
	t.Run("pre-configured instance", func(t *testing.T) {
		instance := &senderMock{}

		sender := NewSender(SenderOptions{
			Instance: instance,
		})

		assert.Same(t, instance, sender)
	})
	t.Run("token required", func(t *testing.T) {
		assert.PanicsWithValue(t, "slogtelegram: token is required", func() {
			_ = NewSender(SenderOptions{})
		})
	})
	t.Run("chat id required", func(t *testing.T) {
		assert.PanicsWithValue(t, "slogtelegram: chat ID is required", func() {
			_ = NewSender(SenderOptions{Token: "test-token"})
		})
	})
	t.Run("telebot sender", func(t *testing.T) {
		sender := NewSender(SenderOptions{
			Token:  "test-token",
			ChatID: 1,
		})

		assert.IsType(t, &TelebotSender{}, sender)
	})
	t.Run("batch sender", func(t *testing.T) {
		sender := NewSender(SenderOptions{
			Token:         "test-token",
			ChatID:        1,
			BatchSize:     2,
			FlushInterval: time.Millisecond * 10,
		})

		assert.IsType(t, &BatchSender{}, sender)
	})
}

func TestTelebotSender(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		client := &clientMock{}
		chatID := int64(1)
		sender := NewTelebotSender(client, chatID)

		msg := "test-message"

		client.On("Send", telebot.ChatID(chatID), msg, []any(nil)).Once().Return(&telebot.Message{}, nil)

		assert.NoError(t, sender.Send(msg))
		assert.NoError(t, sender.Close())
	})
	t.Run("Failure", func(t *testing.T) {
		client := &clientMock{}
		chatID := int64(1)
		sender := NewTelebotSender(client, chatID)

		msg := "test-message"

		client.On("Send", telebot.ChatID(chatID), msg, []any(nil)).Once().Return(nil, errors.New("test-error"))

		assert.EqualError(t, sender.Send(msg), "test-error")
		assert.NoError(t, sender.Close())
	})
}

func TestBatchSender(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		parent := &senderMock{}
		sender := NewBatchSender(parent, 2, time.Millisecond*10)

		msg1 := "msg1"
		msg2 := "msg2"
		msg3 := "msg3"

		parent.On("Send", "msg1\n---\nmsg2\n---\n").Once().Return(nil)
		parent.On("Send", "msg3\n---\n").Once().Return(nil)
		parent.On("Close").Once().Return(nil)

		assert.NoError(t, sender.Send(msg1))
		assert.NoError(t, sender.Send(msg2))
		time.Sleep(time.Millisecond * 20)
		assert.NoError(t, sender.Send(msg3))

		assert.NoError(t, sender.Close())
	})
	t.Run("Failure", func(t *testing.T) {
		parent := &senderMock{}
		sender := NewBatchSender(parent, 2, time.Millisecond*10)

		msg1 := "msg1"
		msg2 := "msg2"
		msg3 := "msg3"

		parent.On("Send", "msg1\n---\nmsg2\n---\n").Once().Return(errors.New("test-error"))
		parent.On("Send", "msg3\n---\n").Once().Return(errors.New("test-error"))
		parent.On("Close").Once().Return(errors.New("test-error"))

		assert.NoError(t, sender.Send(msg1))
		assert.NoError(t, sender.Send(msg2))
		time.Sleep(time.Millisecond * 20)
		assert.NoError(t, sender.Send(msg3))

		assert.EqualError(t, sender.Close(), "test-error")
	})
}

type clientMock struct {
	mock.Mock
}

func (m *clientMock) Send(to telebot.Recipient, what any, opts ...any) (*telebot.Message, error) {
	args := m.Called(to, what, opts)

	res, _ := args.Get(0).(*telebot.Message)

	return res, args.Error(1)
}

type senderMock struct {
	mock.Mock
}

func (m *senderMock) Send(msg string) error {
	args := m.Called(msg)

	return args.Error(0)
}

func (m *senderMock) Close() error {
	args := m.Called()

	return args.Error(0)
}
