package auth

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

func (a *Api) Authenticate(ctx context.Context, in *models.AuthenticateRequest) (*models.User, error) {
	fn := "Authenticate"
	log := a.logger.With(sl.Method(fn))

	body, err := json.Marshal(map[string]any{
		"roles": in.Roles,
	})
	if err != nil {
		log.Error("failed to marshal request", sl.Err(err))
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/authenticate", a.host)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		log.Error("failed to create request", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", in.Token))

	log.Debug("requesting", slog.Any("body", string(body)), slog.String("token", in.Token))
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

	raw, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error("failed to read response body", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	user := new(models.User)
	if err := json.Unmarshal(raw, user); err != nil {
		log.Error("failed to unmarshal response body", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return user, nil
}
