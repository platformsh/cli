package streaming

import (
	"encoding/json"
	"time"
)

const (
	MessageTypeLog         = "log"
	MessageTypeOutputChunk = "output_chunk"
	MessageTypeData        = "data"
	MessageTypeDone        = "done"
	MessageTypeKeepAlive   = "keep_alive"

	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelError = "error"
	LogLevelWarn  = "warn"
)

type Message struct {
	Type string    `json:"type"` // See MessageType constants.
	Time time.Time `json:"time,omitempty"`

	Message string   `json:"message,omitempty"` // For output or log messages.
	Level   string   `json:"level,omitempty"`   // For log messages: see LogLevel constants.
	Tags    []string `json:"tags,omitempty"`

	Key  string          `json:"key,omitempty"` // Used to identify data.
	Data json.RawMessage `json:"data,omitempty"`
}
