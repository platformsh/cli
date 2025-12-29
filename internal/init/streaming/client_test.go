package streaming_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"

	"github.com/upsun/cli/internal/init/streaming"
)

func TestStreaming(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		stream, err := NewServerHandler(w, &ServerConfig{KeepAliveInterval: 10 * time.Second})
		if err != nil {
			t.Fatal(err)
		}
		defer stream.Close()
		stream.Info("first message")
		time.Sleep(20 * time.Millisecond)
		stream.Warn("warning")
		stream.Debug("debug")
		stream.Output("output chunk 1\n")
		time.Sleep(20 * time.Millisecond)
		stream.Output("output chunk 2\n")
		stream.Output("output chunk 3\n")
		stream.LogWithTags(streaming.LogLevelInfo, "tagged message", "example")
		stream.Error("error")
		stream.Info("more output")
	}))
	t.Cleanup(s.Close)

	resp, err := http.Get(s.URL)
	if err != nil {
		t.Fatal(err)
	}

	expectedMessages := []streaming.Message{
		{Type: streaming.MessageTypeLog, Level: streaming.LogLevelInfo, Message: "first message"},
		{Type: streaming.MessageTypeLog, Level: streaming.LogLevelWarn, Message: "warning"},
		{Type: streaming.MessageTypeLog, Level: streaming.LogLevelDebug, Message: "debug"},
		{Type: streaming.MessageTypeOutputChunk, Message: "output chunk 1\n"},
		{Type: streaming.MessageTypeOutputChunk, Message: "output chunk 2\n"},
		{Type: streaming.MessageTypeOutputChunk, Message: "output chunk 3\n"},
		{Type: streaming.MessageTypeLog, Level: streaming.LogLevelInfo,
			Message: "tagged message", Tags: []string{"example"}},
		{Type: streaming.MessageTypeLog, Level: streaming.LogLevelError, Message: "error"},
		{Type: streaming.MessageTypeLog, Level: streaming.LogLevelInfo, Message: "more output"},
	}

	var errGroup errgroup.Group
	var msgChan = make(chan streaming.Message)
	errGroup.Go(func() error {
		defer close(msgChan)
		return streaming.HandleResponse(t.Context(), resp, msgChan)
	})
	errGroup.Go(func() error {
		i := 0
		for msg := range msgChan {
			assert.NotZero(t, msg.Time)
			if i >= len(expectedMessages) {
				t.Fatalf("expected %d messages but another was received: %s", len(expectedMessages), msg.Message)
			}
			expected := expectedMessages[i]
			assert.NotEmpty(t, expected)
			expected.Time = msg.Time
			assert.EqualValues(t, expected, msg)
			i++
		}
		return nil
	})

	if err := errGroup.Wait(); err != nil {
		t.Fatal(err)
	}
}

type ServerHandler struct {
	w    http.ResponseWriter
	f    http.Flusher
	enc  *json.Encoder
	done chan struct{}
	mux  sync.Mutex
}

type ServerConfig struct {
	KeepAliveInterval time.Duration
}

// NewServerHandler creates an HTTP streaming server handler and sets headers.
func NewServerHandler(w http.ResponseWriter, cnf *ServerConfig) (*ServerHandler, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming not supported: the writer must implement http.Flusher")
	}
	// See: https://github.com/ndjson/ndjson-spec
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	h := &ServerHandler{
		w:    w,
		f:    flusher,
		enc:  json.NewEncoder(w),
		done: make(chan struct{}),
	}

	go func() {
		ticker := time.NewTicker(cnf.KeepAliveInterval)
		defer ticker.Stop()
		for {
			select {
			case <-h.done:
				return
			case <-ticker.C:
				h.send(&streaming.Message{Type: streaming.MessageTypeKeepAlive})
			}
		}
	}()

	return h, nil
}

func (h *ServerHandler) Close() {
	close(h.done)
	h.send(&streaming.Message{Type: streaming.MessageTypeDone})
}

func (h *ServerHandler) Output(chunk string, tags ...string) {
	h.send(&streaming.Message{Type: streaming.MessageTypeOutputChunk, Message: chunk, Tags: tags})
}

func (h *ServerHandler) SendData(data json.RawMessage, key string) {
	h.send(&streaming.Message{Type: streaming.MessageTypeData, Data: data, Key: key})
}

func (h *ServerHandler) send(msg *streaming.Message) {
	if msg.Time.IsZero() {
		msg.Time = time.Now()
	}
	h.mux.Lock()
	if err := h.enc.Encode(msg); err != nil {
		panic(fmt.Sprintf("failed to encode message: %v", err))
	}
	h.f.Flush()
	h.mux.Unlock()
}

func (h *ServerHandler) log(level, format string, args ...any) {
	h.send(&streaming.Message{
		Type:    streaming.MessageTypeLog,
		Level:   level,
		Message: fmt.Sprintf(format, args...),
	})
}

func (h *ServerHandler) LogWithTags(level, message string, tags ...string) {
	h.send(&streaming.Message{
		Type:    streaming.MessageTypeLog,
		Level:   level,
		Message: message,
		Tags:    tags,
	})
}

func (h *ServerHandler) Debug(format string, args ...any) {
	h.log(streaming.LogLevelDebug, format, args...)
}

func (h *ServerHandler) Info(format string, args ...any) {
	h.log(streaming.LogLevelInfo, format, args...)
}

func (h *ServerHandler) Warn(format string, args ...any) {
	h.log(streaming.LogLevelWarn, format, args...)
}

func (h *ServerHandler) Error(format string, args ...any) {
	h.log(streaming.LogLevelError, format, args...)
}
