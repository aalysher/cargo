package repo

import (
	"context"
	"fmt"
)

// AdminOrder — заказ + сводка клиента для админки (JOIN clients).
type AdminOrder struct {
	Order
	ClientFirstName string `json:"client_first_name"`
	ClientLastName  string `json:"client_last_name"`
	ClientUsername  string `json:"client_username"`
	ClientPhotoURL  string `json:"client_photo_url"`
	ClientPhone     string `json:"client_phone"`
	ClientCode      string `json:"client_code"`
}

// adminOrderCols — колонки JOIN-выборки (префиксы o/c).
const adminOrderCols = `o.id, o.order_number, o.client_id, o.status::text, o.transport_type::text,
	o.contents, o.declared_weight_kg, o.actual_weight_kg, o.chargeable_weight_kg,
	o.estimated_cost, o.final_cost, o.currency, coalesce(o.comment, '') comment,
	o.created_at, o.updated_at,
	c.first_name, c.last_name, coalesce(c.username, '') username, coalesce(c.photo_url, '') photo_url,
	coalesce(c.phone, '') phone, c.client_code`

// AdminOrders — административные запросы к orders.
type AdminOrders struct{ db DBTX }

// NewAdminOrders конструирует репозиторий.
func NewAdminOrders(db DBTX) *AdminOrders { return &AdminOrders{db: db} }

// ListAdmin — постраничный список с фильтрами: точный статус и поиск (номер, содержимое, код, телефон).
func (a *AdminOrders) ListAdmin(ctx context.Context, status, search string, cursor int64, limit int) ([]AdminOrder, int64, error) {
	q := `SELECT ` + adminOrderCols + ` FROM orders o JOIN clients c ON c.id = o.client_id WHERE true`
	args := []any{}
	ph := 1
	if status != "" {
		q += fmt.Sprintf(` AND o.status::text = $%d`, ph)
		args = append(args, status)
		ph++
	}
	if search != "" {
		pattern := "%" + search + "%"
		q += fmt.Sprintf(` AND (o.order_number ILIKE $%d OR o.contents ILIKE $%d
			OR c.client_code ILIKE $%d OR coalesce(c.phone, '') ILIKE $%d)`, ph, ph, ph, ph)
		args = append(args, pattern)
		ph++
	}
	if cursor > 0 {
		q += fmt.Sprintf(` AND o.id < $%d`, ph)
		args = append(args, cursor)
		ph++
	}
	q += fmt.Sprintf(` ORDER BY o.id DESC LIMIT $%d`, ph)
	args = append(args, limit+1)

	rows, err := a.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []AdminOrder
	for rows.Next() {
		var ao AdminOrder
		if err := scanAdminOrder(rows, &ao); err != nil {
			return nil, 0, err
		}
		out = append(out, ao)
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

// GetAdminByID — заказ с клиентом по id.
func (a *AdminOrders) GetAdminByID(ctx context.Context, id int64) (AdminOrder, error) {
	row := a.db.QueryRow(ctx,
		`SELECT `+adminOrderCols+` FROM orders o JOIN clients c ON c.id = o.client_id WHERE o.id = $1`, id)
	var ao AdminOrder
	err := scanAdminOrder(row, &ao)
	return ao, err
}

// OrdersByClient — последние заказы клиента (карточка клиента в админке).
func (a *AdminOrders) OrdersByClient(ctx context.Context, clientID int64, limit int) ([]Order, error) {
	rows, err := a.db.Query(ctx,
		`SELECT `+orderCols+` FROM orders WHERE client_id = $1 ORDER BY id DESC LIMIT $2`,
		clientID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Order
	for rows.Next() {
		var ord Order
		if err := scanOrder(rows, &ord); err != nil {
			return nil, err
		}
		out = append(out, ord)
	}
	return out, rows.Err()
}

// UpdateWeight сохраняет фактический вес, расчётный вес и пересчитанную стоимость.
func (a *AdminOrders) UpdateWeight(ctx context.Context, tx DBTX, id int64,
	actual, chargeable, estimate float64) (Order, error) {

	row := tx.QueryRow(ctx, `
		UPDATE orders SET actual_weight_kg = $2, chargeable_weight_kg = $3,
			estimated_cost = $4, updated_at = now()
		WHERE id = $1
		RETURNING `+orderCols, id, actual, chargeable, estimate)

	var ord Order
	err := scanOrder(row, &ord)
	return ord, err
}

// UpdateFinalCost сохраняет финальную стоимость (при выдаче).
func (a *AdminOrders) UpdateFinalCost(ctx context.Context, tx DBTX, id int64, cost float64) (Order, error) {
	row := tx.QueryRow(ctx, `
		UPDATE orders SET final_cost = $2, updated_at = now() WHERE id = $1
		RETURNING `+orderCols, id, cost)

	var ord Order
	err := scanOrder(row, &ord)
	return ord, err
}

// UpdateComment заменяет комментарий оператора.
func (a *AdminOrders) UpdateComment(ctx context.Context, tx DBTX, id int64, comment string) (Order, error) {
	row := tx.QueryRow(ctx, `
		UPDATE orders SET comment = nullif($2, ''), updated_at = now() WHERE id = $1
		RETURNING `+orderCols, id, comment)

	var ord Order
	err := scanOrder(row, &ord)
	return ord, err
}

// DeleteTrack удаляет трек заказа; ErrNoRows — трек не найден. Возвращает удалённый номер.
func (a *AdminOrders) DeleteTrack(ctx context.Context, tx DBTX, orderID, trackID int64) (string, error) {
	var tracking string
	err := tx.QueryRow(ctx,
		`DELETE FROM order_tracks WHERE id = $1 AND order_id = $2 RETURNING tracking_number`,
		trackID, orderID).Scan(&tracking)
	return tracking, err
}

// scanAdminOrder сканирует строку с adminOrderCols.
func scanAdminOrder(row interface{ Scan(dest ...any) error }, ao *AdminOrder) error {
	return row.Scan(
		&ao.ID, &ao.OrderNumber, &ao.ClientID, &ao.Status, &ao.TransportType,
		&ao.Contents, &ao.DeclaredWeightKg, &ao.ActualWeightKg, &ao.ChargeableWeightKg,
		&ao.EstimatedCost, &ao.FinalCost, &ao.Currency, &ao.Comment, &ao.CreatedAt, &ao.UpdatedAt,
		&ao.ClientFirstName, &ao.ClientLastName, &ao.ClientUsername, &ao.ClientPhotoURL,
		&ao.ClientPhone, &ao.ClientCode,
	)
}
