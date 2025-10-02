package init

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/platformsh/cli/internal/init/streaming"
	"github.com/platformsh/cli/internal/tui"
)

type dataHandler func(data json.RawMessage, key string) error

func handleMessage(msg *streaming.Message, stdout, stderr io.Writer, spinr *tui.Spinner, handleData dataHandler) error { //nolint:lll
	logger := &logPrinter{spinr: spinr, stderr: stderr}

	spinr.Stop()

	switch msg.Type {
	case streaming.MessageTypeLog:
		logger.print(msg.Level, msg.Message, msg.Tags...)
	case streaming.MessageTypeOutputChunk:
		fmt.Fprint(stdout, msg.Message)
	case streaming.MessageTypeData:
		if err := handleData(msg.Data, msg.Key); err != nil {
			return err
		}
	default:
		logger.print(streaming.LogLevelError, fmt.Sprintf("Unknown message type: %v\n", msg.Type))
	}

	return nil
}
