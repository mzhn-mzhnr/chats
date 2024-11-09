package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mzhn/chats/internal/storage/models"
	"mzhn/chats/pkg/sl"
	"net/http"
)

type invokeResponse struct {
	Output struct {
		Result struct {
			Answer  string `json:"answer"`
			Sources []struct {
				FileId   string `json:"file_id"`
				FileName string `json:"file_name"`
				Page     int    `json:"page_number"`
			} `json:"sources"`
		} `json:"result"`
	} `json:"output"`
}

func (a *Api) Invoke(ctx context.Context, in *models.RagRequest) (*models.RagResponse, error) {

	fn := "Invoke"
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
		log.Error("unexpected status code", slog.Int("status", response.StatusCode))
		return nil, fmt.Errorf("%s: unexpected request error (%d)", fn, response.StatusCode)
	}

	resp, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error("failed to read response", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	var j invokeResponse
	if err := json.Unmarshal(resp, &j); err != nil {
		log.Error("failed to unmarshal response", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	out := &models.RagResponse{
		Answer:  j.Output.Result.Answer,
		Sources: make([]models.AnswerMeta, len(j.Output.Result.Sources)),
	}

	for i, s := range j.Output.Result.Sources {
		out.Sources[i] = models.AnswerMeta{
			FileId:   s.FileId,
			Filename: s.FileName,
			Slide:    s.Page,
		}
	}

	return out, nil
}
