package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const apiBase = "https://api.telegram.org/bot"

// Client отправляет сообщения через Telegram Bot API.
type Client struct {
	token  string
	chatID string
	http   *http.Client
}

func NewClient(token, chatID string) *Client {
	return &Client{
		token:  token,
		chatID: chatID,
		http:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) SendMessage(ctx context.Context, text string) error {
	payload := map[string]any{
		"chat_id":    c.chatID,
		"text":       text,
		"parse_mode": "HTML",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s%s/sendMessage", apiBase, c.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusMultipleChoices {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API: %s: %s", resp.Status, string(respBody))
	}
	return nil
}

func (c *Client) SendToChat(ctx context.Context, chatID, text string) error {
	prev := c.chatID
	c.chatID = chatID
	defer func() { c.chatID = prev }()
	return c.SendMessage(ctx, text)
}
