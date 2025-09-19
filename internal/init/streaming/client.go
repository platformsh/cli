package streaming

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleResponse(ctx context.Context, resp *http.Response, msgChan chan<- Message) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	// Note: Transfer-Encoding can't be checked like this as it is a hop-by-hop header.
	if ct := resp.Header.Get("Content-Type"); ct != "application/x-ndjson" {
		return fmt.Errorf("unexpected content type: %s", ct)
	}

	scanner := bufio.NewScanner(resp.Body)
	// Start with default buffer, expand to 1MB if needed for large lines
	scanner.Buffer(nil, 1024*1024)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		var msg Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			return fmt.Errorf("failed to decode line: %w: %s", err, string(scanner.Bytes()))
		}
		if msg.Type == MessageTypeKeepAlive {
			continue
		} else if msg.Type == MessageTypeDone {
			break
		}
		msgChan <- msg
	}

	return scanner.Err()
}
