package slogtelegram

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"gopkg.in/telebot.v4"
)

var (
	messageSeparator = "\n---\n"
	messageMaxSize   = 4096
)

type Sender interface {
	Send(msg string) error
	Close() error
}

type SenderOptions struct {
	// Token is the Telegram bot token to send messages. It is required, please generate it from @BotFather.
	Token string

	// ChatID is the chat ID to send messages to (private, group or channel).
	ChatID int64

	// HTTPClient is the client used to send messages to the Telegram API.
	HTTPClient *http.Client

	// BaseURL is the Telegram API basic url to send messages.
	BaseURL string

	// BatchSize is the maximum number of messages to send in a single batch.
	BatchSize uint64

	// FlushInterval is the maximum duration to wait before sending a batch.
	FlushInterval time.Duration

	// Verbose specifies whether to print the Telegram API requests and responses.
	Verbose bool

	// Instance is a custom sender instance to use.
	Instance Sender
}

func NewSender(opts SenderOptions) Sender {
	if opts.Instance != nil {
		return opts.Instance
	}

	if opts.Token == "" {
		panic(fmt.Sprintf("%stoken is required", errPrefix))
	}

	if opts.ChatID == 0 {
		panic(fmt.Sprintf("%schat ID is required", errPrefix))
	}

	if opts.HTTPClient == nil {
		opts.HTTPClient = http.DefaultClient
	}

	client, err := telebot.NewBot(telebot.Settings{
		URL:       opts.BaseURL,
		Token:     opts.Token,
		Verbose:   opts.Verbose,
		Client:    opts.HTTPClient,
		ParseMode: telebot.ModeHTML,
		Offline:   true,
	})
	if err != nil {
		panic(fmt.Sprintf("%s%v", errPrefix, err))
	}

	sender := NewTelebotSender(client, opts.ChatID)
	if opts.BatchSize == 0 && opts.FlushInterval == 0 {
		return sender
	}

	return NewBatchSender(sender, opts.BatchSize, opts.FlushInterval)
}

type client interface {
	Send(to telebot.Recipient, what any, opts ...any) (*telebot.Message, error)
}

type TelebotSender struct {
	client client
	chatID int64
}

func NewTelebotSender(client client, chatID int64) *TelebotSender {
	return &TelebotSender{
		client: client,
		chatID: chatID,
	}
}

func (s *TelebotSender) Send(msg string) error {
	if _, err := s.client.Send(telebot.ChatID(s.chatID), msg); err != nil {
		return err
	}

	return nil
}

func (s *TelebotSender) Close() error {
	return nil
}

type BatchSender struct {
	parent        Sender
	items         []string
	wg            sync.WaitGroup
	mutex         sync.Mutex
	batchSize     uint64
	flushInterval time.Duration
	cancel        func()
}

func NewBatchSender(parent Sender, batchSize uint64, flushInterval time.Duration) *BatchSender {
	s := &BatchSender{
		parent:        parent,
		items:         make([]string, 0, batchSize),
		batchSize:     batchSize,
		flushInterval: flushInterval,
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		s.run(ctx)
	}()

	return s
}

func (s *BatchSender) Send(msg string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.items = append(s.items, msg)
	if uint64(len(s.items)) >= s.batchSize {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			if err := s.flush(); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
			}
		}()
	}

	return nil
}

func (s *BatchSender) Close() error {
	s.cancel()
	s.wg.Wait()

	return s.parent.Close()
}

func (s *BatchSender) run(ctx context.Context) {
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if err := s.flush(); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
			}
			return
		case <-ticker.C:
			if err := s.flush(); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
			}
		}
	}
}

func (s *BatchSender) flush() error {
	s.mutex.Lock()
	if len(s.items) == 0 {
		s.mutex.Unlock()

		return nil
	}
	items := append([]string(nil), s.items...)
	s.items = s.items[:0]
	s.mutex.Unlock()

	for chunk := range s.chunks(items) {
		if err := s.parent.Send(chunk); err != nil {
			return err
		}
	}

	return nil
}

func (s *BatchSender) chunks(items []string) iter.Seq[string] {
	return func(yield func(string) bool) {
		var b strings.Builder
		for _, item := range items {
			if b.Len() > 0 && b.Len()+len(item)+len(messageSeparator) > messageMaxSize {
				if !yield(b.String()) {
					return
				}

				b.Reset()
			}

			b.WriteString(item)
			b.WriteString(messageSeparator)
		}

		if b.Len() > 0 {
			yield(b.String())
		}
	}
}
