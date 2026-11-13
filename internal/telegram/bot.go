// Package telegram: минимальный клиент Bot API для уведомлений.
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Bot — клиент Telegram Bot API.
type Bot struct {
	token string
	http  *http.Client
}

// New создаёт бота.
func New(token string) *Bot {
	return &Bot{token: token, http: &http.Client{Timeout: 10 * time.Second}}
}

// Enabled — задан ли токен.
func (b *Bot) Enabled() bool { return b.token != "" }

// SendMessage отправляет сообщение; buttonURL — inline-кнопка «Открыть заказ» (deep link).
func (b *Bot) SendMessage(ctx context.Context, chatID int64, text, buttonURL string) error {
	if !b.Enabled() {
		return fmt.Errorf("telegram: BOT_TOKEN не задан")
	}

	msg := map[string]any{
		"chat_id":                  chatID,
		"text":                     text,
		"disable_web_page_preview": true,
	}
	if buttonURL != "" {
		msg["reply_markup"] = map[string]any{
			"inline_keyboard": [][]map[string]string{
				{{"text": "Открыть заказ", "url": buttonURL}},
			},
		}
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.token), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram: статус %d", resp.StatusCode)
	}
	return nil
}
