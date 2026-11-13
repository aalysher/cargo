package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Order — запись таблицы orders. Статусы — строки (enum приводится к ::text).
type Order struct {
	ID                 int64     `json:"id"`
	OrderNumber        string    `json:"order_number"`
	ClientID           int64     `json:"client_id"`
	Status             string    `json:"status"`
	TransportType      string    `json:"transport_type"`
	Contents           string    `json:"contents"`
	DeclaredWeightKg   float64   `json:"declared_weight_kg"`
	ActualWeightKg     *float64  `json:"actual_weight_kg"`
	ChargeableWeightKg *float64  `json:"chargeable_weight_kg"`
	EstimatedCost      *float64  `json:"estimated_cost"`
	FinalCost          *float64  `json:"final_cost"`
	Currency           string    `json:"currency"`
	Comment            string    `json:"comment"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// OrderTrack — трек-номер покупки, привязанный к заказу.
type OrderTrack struct {
	ID             int64     `json:"id"`
	OrderID        int64     `json:"order_id"`
	TrackingNumber string    `json:"tracking_number"`
	CreatedAt      time.Time `json:"created_at"`
}

// orderCols — колонки для выборки заказа.
const orderCols = `id, order_number, client_id, status::text, transport_type::text, contents,
	declared_weight_kg, actual_weight_kg, chargeable_weight_kg, estimated_cost, final_cost,
	currency, coalesce(comment, '') comment, created_at, updated_at`

// Orders — доступ к orders и order_tracks.
type Orders struct{ db DBTX }

// NewOrders конструирует репозиторий.
func NewOrders(db DBTX) *Orders { return &Orders{db: db} }

// Insert создаёт заказ (статус new по умолчанию БД) и возвращает его.
func (o *Orders) Insert(ctx context.Context, tx DBTX, ord Order) (Order, error) {
	row := tx.QueryRow(ctx, `
		INSERT INTO orders (order_number, client_id, transport_type, contents,
			declared_weight_kg, estimated_cost, currency, comment)
		VALUES ($1, $2, $3::transport_type, $4, $5, $6, $7, nullif($8, ''))
		RETURNING `+orderCols,
		ord.OrderNumber, ord.ClientID, ord.TransportType, ord.Contents,
		ord.DeclaredWeightKg, ord.EstimatedCost, ord.Currency, ord.Comment)

	var out Order
	err := scanOrder(row, &out)
	return out, err
}

// AddTrack добавляет трек; при дубликате возвращает pgx.ErrNoRows (сервис игнорирует).
func (o *Orders) AddTrack(ctx context.Context, tx DBTX, orderID int64, tracking string) (OrderTrack, error) {
	row := tx.QueryRow(ctx, `
		INSERT INTO order_tracks (order_id, tracking_number) VALUES ($1, $2)
		ON CONFLICT (order_id, tracking_number) DO NOTHING
		RETURNING id, order_id, tracking_number, created_at`, orderID, tracking)

	var t OrderTrack
	err := row.Scan(&t.ID, &t.OrderID, &t.TrackingNumber, &t.CreatedAt)
	return t, err
}

// TracksByOrder — треки заказа.
func (o *Orders) TracksByOrder(ctx context.Context, orderID int64) ([]OrderTrack, error) {
	rows, err := o.db.Query(ctx,
		`SELECT id, order_id, tracking_number, created_at FROM order_tracks WHERE order_id = $1 ORDER BY id`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []OrderTrack
	for rows.Next() {
		var t OrderTrack
		if err := rows.Scan(&t.ID, &t.OrderID, &t.TrackingNumber, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// GetForClient возвращает заказ клиента (изоляция по client_id).
func (o *Orders) GetForClient(ctx context.Context, id, clientID int64) (Order, error) {
	row := o.db.QueryRow(ctx, `SELECT `+orderCols+` FROM orders WHERE id = $1 AND client_id = $2`, id, clientID)
	var ord Order
	err := scanOrder(row, &ord)
	return ord, err
}

// ListByClient — постраничный список заказов клиента (cursor = id последнего элемента).
// statuses пустой — все статусы. Возвращает nextCursor (0 — дальше страниц нет).
func (o *Orders) ListByClient(ctx context.Context, clientID int64, statuses []string, cursor int64, limit int) ([]Order, int64, error) {
	q := `SELECT ` + orderCols + ` FROM orders WHERE client_id = $1`
	args := []any{clientID}
	ph := 2
	if cursor > 0 {
		q += fmt.Sprintf(` AND id < $%d`, ph)
		args = append(args, cursor)
		ph++
	}
	if len(statuses) > 0 {
		q += fmt.Sprintf(` AND status::text = ANY($%d)`, ph)
		args = append(args, statuses)
		ph++
	}
	q += fmt.Sprintf(` ORDER BY id DESC LIMIT $%d`, ph)
	args = append(args, limit+1)

	rows, err := o.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []Order
	for rows.Next() {
		var ord Order
		if err := scanOrder(rows, &ord); err != nil {
			return nil, 0, err
		}
		out = append(out, ord)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var next int64
	if len(out) > limit {
		out = out[:limit]
		next = out[len(out)-1].ID
	}
	return out, next, nil
}

// UpdateStatus меняет статус с защитой от гонок (AND status = from).
// pgx.ErrNoRows — статус уже изменился параллельным запросом.
func (o *Orders) UpdateStatus(ctx context.Context, tx DBTX, id int64, from, to string) (Order, error) {
	row := tx.QueryRow(ctx, `
		UPDATE orders SET status = $2::order_status, updated_at = now()
		WHERE id = $1 AND status = $3::order_status
		RETURNING `+orderCols, id, to, from)

	var ord Order
	err := scanOrder(row, &ord)
	return ord, err
}

// scanOrder сканирует строку с orderCols.
func scanOrder(row pgx.Row, ord *Order) error {
	return row.Scan(&ord.ID, &ord.OrderNumber, &ord.ClientID, &ord.Status, &ord.TransportType,
		&ord.Contents, &ord.DeclaredWeightKg, &ord.ActualWeightKg, &ord.ChargeableWeightKg,
		&ord.EstimatedCost, &ord.FinalCost, &ord.Currency, &ord.Comment, &ord.CreatedAt, &ord.UpdatedAt)
}
