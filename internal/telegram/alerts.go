package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// AlertmanagerWebhook — формат webhook от Alertmanager v4.
type AlertmanagerWebhook struct {
	Status string  `json:"status"`
	Alerts []Alert `json:"alerts"`
}

type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
}

// AlertHandler принимает webhook от Alertmanager и шлёт сообщения в Telegram.
type AlertHandler struct {
	client *Client
}

func NewAlertHandler(client *Client) *AlertHandler {
	return &AlertHandler{client: client}
}

func (h *AlertHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var payload AlertmanagerWebhook
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	text := formatAlerts(payload)
	if err := h.client.SendMessage(r.Context(), text); err != nil {
		slog.Error("не удалось отправить алерт в Telegram", "error", err)
		http.Error(w, "telegram error", http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func formatAlerts(payload AlertmanagerWebhook) string {
	if len(payload.Alerts) == 0 {
		return fmt.Sprintf("<b>%s</b>\nНет активных алертов", strings.ToUpper(payload.Status))
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("<b>Alertmanager: %s</b>\n\n", strings.ToUpper(payload.Status)))

	for i, alert := range payload.Alerts {
		if i > 0 {
			b.WriteString("\n")
		}
		name := alert.Labels["alertname"]
		severity := alert.Labels["severity"]
		b.WriteString(fmt.Sprintf("<b>%s</b> [%s]\n", name, strings.ToUpper(alert.Status)))
		if severity != "" {
			b.WriteString(fmt.Sprintf("Severity: %s\n", severity))
		}
		if summary := alert.Annotations["summary"]; summary != "" {
			b.WriteString(summary + "\n")
		}
		if desc := alert.Annotations["description"]; desc != "" {
			b.WriteString(desc + "\n")
		}
		if !alert.StartsAt.IsZero() {
			b.WriteString(fmt.Sprintf("Started: %s\n", alert.StartsAt.Format(time.RFC3339)))
		}
	}
	return b.String()
}

// RunCommandListener слушает команды /start и /chatid через long polling.
func RunCommandListener(ctx context.Context, token string, client *Client) {
	if token == "" {
		return
	}

	offset := 0
	httpClient := &http.Client{Timeout: 35 * time.Second}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		url := fmt.Sprintf("%s%s/getUpdates?timeout=30&offset=%d", apiBase, token, offset)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			slog.Error("telegram polling: запрос", "error", err)
			time.Sleep(3 * time.Second)
			continue
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Error("telegram polling: сеть", "error", err)
			time.Sleep(3 * time.Second)
			continue
		}

		var updates struct {
			OK     bool `json:"ok"`
			Result []struct {
				UpdateID int `json:"update_id"`
				Message  struct {
					Text string `json:"text"`
					Chat struct {
						ID int64 `json:"id"`
					} `json:"chat"`
				} `json:"message"`
			} `json:"result"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&updates); err != nil {
			resp.Body.Close()
			slog.Error("telegram polling: decode", "error", err)
			time.Sleep(3 * time.Second)
			continue
		}
		resp.Body.Close()

		for _, update := range updates.Result {
			offset = update.UpdateID + 1
			text := strings.TrimSpace(update.Message.Text)
			chatID := fmt.Sprintf("%d", update.Message.Chat.ID)

			switch text {
			case "/start", "/chatid":
				reply := fmt.Sprintf(
					"Бот алертов запущен.\nВаш chat_id: <code>%s</code>\nСкопируйте его в .env как TELEGRAM_CHAT_ID",
					chatID,
				)
				if err := client.SendToChat(ctx, chatID, reply); err != nil {
					slog.Error("telegram polling: ответ", "error", err)
				}
			}
		}
	}
}
