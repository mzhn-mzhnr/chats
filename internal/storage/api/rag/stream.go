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
	"mzhn/chats/internal/storage/models"
	"mzhn/chats/pkg/sl"
	"net/http"
	"strings"
)

type Metainfo struct {
	FileId     string `json:"file_id"`
	FileName   string `json:"file_name"`
	PageNumber int    `json:"page_number"`
}

type Response struct {
	Response string `json:"response"`
}

type data struct {
	Metainfo *Metainfo `json:"metainfo"`
	*Response
}

func (a *Api) Stream(ctx context.Context, in *models.RagRequest, eventCh chan<- []byte) (*models.AnswerMeta, error) {
	defer close(eventCh)

	fn := "Stream"
	log := a.logger.With(sl.Method(fn))

	chathistory := make([][2]string, len(in.ChatHistory))

	for i, entry := range in.ChatHistory {
		who := "ai"
		if entry.IsUser {
			who = "human"
		}

		chathistory[i][0] = who
		chathistory[i][1] = entry.Body
	}

	body, err := json.Marshal(map[string]any{
		"input": map[string]any{
			"input":        in.Input,
			"chat_history": chathistory,
		},
	})
	if err != nil {
		log.Error("failed to marshal request", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	endpoint := fmt.Sprintf("%s/chat/stream", a.host)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		log.Error("failed to create request", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	request.Header.Add("Content-Type", "application/json")

	log.Debug("requesting", slog.Any("body", string(body)))
	response, err := a.client.Do(request)
	if err != nil {
		log.Error("failed to send request", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Error("response status is not OK", slog.Int("status", response.StatusCode))
		return nil, fmt.Errorf("%s: %w", fn, fmt.Errorf("failed request (%d)", response.StatusCode))
	}

	meta := new(models.AnswerMeta)
	reader := bufio.NewReader(response.Body)
	for {
		event, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Info("stream ended")
				break
			}
			log.Error("failed to read meta of response", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", fn, err)
		}
		slog.Info("received event", slog.String("event", event))
		event = strings.TrimPrefix(event, "event: ")

		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Info("stream ended")
				break
			}
			log.Error("failed to read data of response", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", fn, err)
		}

		if _, err := reader.ReadString('\n'); err != nil {
			if errors.Is(err, io.EOF) {
				log.Info("stream ended")
				break
			}
			log.Error("failed to read data of response", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", fn, err)
		}

		if strings.Compare(event, "metadata") == 0 {
			log.Info("metadata received")
			continue
		} else if strings.Compare(event, "end") == 0 {
			log.Info("stream ended")
			break
		}

		slog.Info("received line", slog.String("line", line))

		line = strings.TrimPrefix(line, "data: ")
		line = strings.Trim(line, "\r\n")

		var j data
		if err := json.Unmarshal([]byte(line), &j); err != nil {
			// log.Error("failed to unmarshal response", sl.Err(err))
			// return nil, fmt.Errorf("%s: %w", fn, err)
			continue
		}

		log.Info(
			"received event",
			slog.String("event", event),
			slog.String("line", line),
			slog.Any("data", j),
		)

		if j.Metainfo == nil && j.Response == nil {
			continue
		} else if j.Metainfo != nil {
			m := j.Metainfo
			meta.FileId = m.FileId
			meta.Filename = m.FileName
			meta.Slide = m.PageNumber
			log.Info("saving meta", slog.Any("meta", meta), slog.Any("m", m))
		} else if j.Response != nil {
			eventCh <- []byte(j.Response.Response)
		}
	}

	log.Info("stream ended. returning meta", slog.Any("meta", meta))
	return meta, nil
}
