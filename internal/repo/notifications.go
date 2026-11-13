package repo

import (
	"context"
	"time"
)

// Notification — outbox-запись для отправки ботом.
type Notification struct {
	ID        int64     `json:"id"`
	ClientID  int64     `json:"client_id"`
	OrderID   *int64    `json:"order_id"`
	Type      string    `json:"type"`
	Payload   string    `json:"payload"`
	Status    string    `json:"status"`
	Attempts  int       `json:"attempts"`
	CreatedAt time.Time `json:"created_at"`
}

// PendingNotification — сообщение из outbox с chat_id получателя.
type PendingNotification struct {
	ID      int64
	OrderID *int64
	Payload string
	ChatID  int64
}

// Notifications — доступ к notifications (outbox).
type Notifications struct{ db DBTX }

// NewNotifications конструирует репозиторий.
func NewNotifications(db DBTX) *Notifications { return &Notifications{db: db} }

// InsertPending добавляет сообщение в outbox (payload — JSON-строка).
func (n *Notifications) InsertPending(ctx context.Context, tx DBTX, clientID int64, orderID *int64, typ, payload string) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO notifications (client_id, order_id, type, payload) VALUES ($1, $2, $3, $4::jsonb)`,
		clientID, orderID, typ, payload)
	return err
}

// ListPending — первые N ожидающих отправки (отправку ведёт один воркер).
func (n *Notifications) ListPending(ctx context.Context, limit int) ([]PendingNotification, error) {
	rows, err := n.db.Query(ctx, `
		SELECT n.id, n.order_id, n.payload::text, c.telegram_id
		FROM notifications n JOIN clients c ON c.id = n.client_id
		WHERE n.status = 'pending'
		ORDER BY n.id LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PendingNotification
	for rows.Next() {
		var p PendingNotification
		if err := rows.Scan(&p.ID, &p.OrderID, &p.Payload, &p.ChatID); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// MarkSent помечает сообщение отправленным.
func (n *Notifications) MarkSent(ctx context.Context, id int64) error {
	_, err := n.db.Exec(ctx,
		`UPDATE notifications SET status = 'sent', sent_at = now() WHERE id = $1`, id)
	return err
}

// MarkFailed увеличивает счётчик попыток; после 5 неудач — статус failed.
// Возвращает итоговый статус.
func (n *Notifications) MarkFailed(ctx context.Context, id int64) (string, error) {
	var status string
	err := n.db.QueryRow(ctx, `
		UPDATE notifications SET attempts = attempts + 1,
			status = CASE WHEN attempts + 1 >= 5 THEN 'failed' ELSE 'pending' END
		WHERE id = $1 RETURNING status::text`, id).Scan(&status)
	return status, err
}
