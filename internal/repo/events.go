package repo

import (
	"context"
	"time"
)

// OrderEvent — запись таймлайна/аудита (order_events).
type OrderEvent struct {
	ID        int64     `json:"id"`
	OrderID   int64     `json:"order_id"`
	EventType string    `json:"event_type"`
	Status    string    `json:"status"`
	Title     string    `json:"title"`
	Payload   string    `json:"payload"`
	ActorType string    `json:"actor_type"`
	ActorID   int64     `json:"actor_id"`
	CreatedAt time.Time `json:"created_at"`
}

// eventCols — колонки для выборки события (NULL → нейтральные значения).
const eventCols = `id, order_id, event_type::text,
	case when status is null then '' else status::text end,
	title, payload::text, actor_type::text, coalesce(actor_id, 0), created_at`

// Events — доступ к order_events.
type Events struct{ db DBTX }

// NewEvents конструирует репозиторий.
func NewEvents(db DBTX) *Events { return &Events{db: db} }

// Insert добавляет событие (payload — JSON-строка, может быть пустой).
func (e *Events) Insert(ctx context.Context, tx DBTX, ev OrderEvent) error {
	var statusParam, actorParam any
	if ev.Status != "" {
		statusParam = ev.Status
	}
	if ev.ActorID != 0 {
		actorParam = ev.ActorID
	}
	payload := ev.Payload
	if payload == "" {
		payload = "{}"
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO order_events (order_id, event_type, status, title, payload, actor_type, actor_id)
		VALUES ($1, $2::event_type, $3::order_status, $4, $5::jsonb, $6::actor_type, $7)`,
		ev.OrderID, ev.EventType, statusParam, ev.Title, payload, ev.ActorType, actorParam)
	return err
}

// ListForOrder — события заказа в хронологическом порядке (таймлайн).
func (e *Events) ListForOrder(ctx context.Context, orderID int64) ([]OrderEvent, error) {
	rows, err := e.db.Query(ctx, `SELECT `+eventCols+` FROM order_events WHERE order_id = $1 ORDER BY id`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []OrderEvent
	for rows.Next() {
		var ev OrderEvent
		if err := rows.Scan(&ev.ID, &ev.OrderID, &ev.EventType, &ev.Status,
			&ev.Title, &ev.Payload, &ev.ActorType, &ev.ActorID, &ev.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	return out, rows.Err()
}
