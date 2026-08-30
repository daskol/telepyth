package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/cenkalti/backoff/v5"
)

const TelegramBotEndpoint = `https://api.telegram.org/bot%s/%s`

func renderEndpoint(token, method string) string {
	return fmt.Sprintf(TelegramBotEndpoint, token, method)
}

var ErrBadRequest = errors.New("bot: request failed")

type Result struct {
	OK          bool   `json:"ok"`
	ErrorCode   any    `json:"error_code"`
	Description string `json:"description"`
	Result      any    `json:"result"`
}

type InputFile struct{}

type SetWebhook struct {
	URL                string     `json:"url"`
	Certificate        InputFile  `json:"certificate,omitzero"`
	IPAddr             net.IPAddr `json:"ip_address,omitzero"`
	MaxConnections     int        `json:"max_connections,omitempty"`
	AllowedUpdates     []string   `json:"allowed_updates,omitempty"`
	DropPendingUpdates bool       `json:"drop_pending_updates,omitempty"`
	SecretToken        string     `json:"secret_token,omitempty"`
}

func (t *TelegramBotApi) SetWebhook(ctx context.Context, params SetWebhook) error {
	client := http.DefaultClient
	_, err := request[Result](ctx, client, t.token, "setWebhook", params)
	return err
}

func request[T any](ctx context.Context, client *http.Client, token, method string, params any) (T, error) {
	var result T

	buf := bytes.Buffer{}
	err := json.NewEncoder(&buf).Encode(params)
	if err != nil {
		return result, err
	}

	url := renderEndpoint(token, method)

	bo := backoff.NewExponentialBackOff()
	bo.Reset()

	var res *http.Response
	for {
		res, err = client.Post(url, "application/json", &buf)
		if err == nil {
			defer res.Body.Close()
			break
		}

		if dur := bo.NextBackOff(); dur != backoff.Stop {
			return result, fmt.Errorf("%w: %s: retries", ErrBadRequest, method)
		}
	}

	buf.Reset()
	io.Copy(&buf, res.Body)
	log.Printf("setWebhook: %s", buf.String())

	if res.StatusCode != http.StatusOK {
		return result, fmt.Errorf("%w: %s", ErrBadRequest, method)
	}

	if err := json.NewDecoder(&buf).Decode(&result); err != nil {
		return result, fmt.Errorf("%w: %s: %w", ErrBadRequest, method, err)
	}
	return result, nil
}
