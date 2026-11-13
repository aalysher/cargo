// worker — отправка уведомлений из outbox в Telegram (polling).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cargo/internal/config"
	"cargo/internal/db"
	"cargo/internal/repo"
	"cargo/internal/telegram"
)

// notifPayload — формат payload уведомления.
type notifPayload struct {
	Text        string `json:"text"`
	OrderNumber string `json:"order_number"`
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	cfg := config.Load()
	if cfg.BotToken == "" {
		slog.Warn("BOT_TOKEN не задан — worker ждёт токен, уведомления не отправляются")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	notifications := repo.NewNotifications(pool)
	bot := telegram.New(cfg.BotToken)

	ticker := time.NewTicker(cfg.NotifyInterval)
	defer ticker.Stop()

	for {
		processOnce(ctx, notifications, bot, cfg)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// processOnce — один проход по pending-уведомлениям.
func processOnce(ctx context.Context, notifications *repo.Notifications, bot *telegram.Bot, cfg *config.Config) {
	pending, err := notifications.ListPending(ctx, 20)
	if err != nil {
		slog.Error("list pending", "err", err)
		return
	}
	for _, p := range pending {
		if !bot.Enabled() {
			return // ждём токен, не сжигаем попытки
		}

		var payload notifPayload
		if err := json.Unmarshal([]byte(p.Payload), &payload); err != nil {
			slog.Error("bad payload", "id", p.ID, "err", err)
			drop(ctx, notifications, p.ID)
			continue
		}

		deepLink := fmt.Sprintf("https://t.me/%s?startapp=order_%d", cfg.BotUsername, deref(p.OrderID))
		if err := bot.SendMessage(ctx, p.ChatID, payload.Text, deepLink); err != nil {
			slog.Warn("send failed", "id", p.ID, "err", err)
			status, ferr := notifications.MarkFailed(ctx, p.ID)
			if ferr != nil {
				slog.Error("mark failed", "err", ferr)
			} else if status == "failed" {
				slog.Error("notification dropped after 5 attempts", "id", p.ID)
			}
			continue
		}

		if err := notifications.MarkSent(ctx, p.ID); err != nil {
			slog.Error("mark sent", "err", err)
		}
	}
}

// drop помечает сломанное уведомление failed.
func drop(ctx context.Context, notifications *repo.Notifications, id int64) {
	if _, err := notifications.MarkFailed(ctx, id); err != nil {
		slog.Error("mark failed", "err", err)
	}
}

func deref(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}
