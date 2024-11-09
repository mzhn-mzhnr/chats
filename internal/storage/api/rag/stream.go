package rag

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mzhn/chats/pkg/sl"
	"net/http"
)

func (a *Api) Stream(ctx context.Context, input string, eventCh chan<- []byte) error {
	defer close(eventCh)

	fn := "Authenticate"
	log := a.logger.With(sl.Method(fn))

	body, err := json.Marshal(map[string]any{
		"input": input,
	})
	if err != nil {
		log.Error("failed to marshal request", sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}

	endpoint := fmt.Sprintf("%s/chat/stream", a.host)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		log.Error("failed to create request", sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}

	request.Header.Add("Content-Type", "application/json")

	log.Debug("requesting", slog.Any("body", string(body)))
	response, err := a.client.Do(request)
	if err != nil {
		log.Error("failed to send request", sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Error("response status is not OK", slog.Int("status", response.StatusCode))
		return fmt.Errorf("%s: %w", fn, fmt.Errorf("failed request (%d)", response.StatusCode))
	}

	reader := bufio.NewReader(response.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Info("stream ended")
				break
			}
			log.Error("failed to read response", sl.Err(err))
			return fmt.Errorf("%s: %w", fn, err)
		}

		log.Info("received line", slog.String("line", line))

		eventCh <- []byte(line)
	}

	return nil
}
