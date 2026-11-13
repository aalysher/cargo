package repo

import (
	"context"
	"fmt"
)

// Tariff — строка справочника тарифов.
type Tariff struct {
	ID                int64   `json:"id"`
	TransportType     string  `json:"transport_type"`
	PricePerKg        float64 `json:"price_per_kg"`
	Currency          string  `json:"currency"`
	VolumetricDivisor int     `json:"volumetric_divisor"`
	MinChargeableKg   float64 `json:"min_chargeable_kg"`
	IsActive          bool    `json:"is_active"`
}

// tariffCols — колонки для выборки тарифа.
const tariffCols = `id, transport_type::text, price_per_kg, currency, volumetric_divisor, min_chargeable_kg, is_active`

// Tariffs — доступ к tariffs.
type Tariffs struct{ db DBTX }

// NewTariffs конструирует репозиторий.
func NewTariffs(db DBTX) *Tariffs { return &Tariffs{db: db} }

// ListActive — активные тарифы (для калькулятора).
func (t *Tariffs) ListActive(ctx context.Context) ([]Tariff, error) {
	rows, err := t.db.Query(ctx, `SELECT `+tariffCols+` FROM tariffs WHERE is_active ORDER BY transport_type`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Tariff
	for rows.Next() {
		var tf Tariff
		if err := rows.Scan(&tf.ID, &tf.TransportType, &tf.PricePerKg, &tf.Currency,
			&tf.VolumetricDivisor, &tf.MinChargeableKg, &tf.IsActive); err != nil {
			return nil, err
		}
		out = append(out, tf)
	}
	return out, rows.Err()
}

// ListAll — все тарифы для админки.
func (t *Tariffs) ListAll(ctx context.Context) ([]Tariff, error) {
	rows, err := t.db.Query(ctx, `SELECT `+tariffCols+` FROM tariffs ORDER BY transport_type`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Tariff
	for rows.Next() {
		var tf Tariff
		if err := rows.Scan(&tf.ID, &tf.TransportType, &tf.PricePerKg, &tf.Currency,
			&tf.VolumetricDivisor, &tf.MinChargeableKg, &tf.IsActive); err != nil {
			return nil, err
		}
		out = append(out, tf)
	}
	return out, rows.Err()
}

// GetActiveByType — активный тариф по типу линии (road|air).
func (t *Tariffs) GetActiveByType(ctx context.Context, transport string) (Tariff, error) {
	row := t.db.QueryRow(ctx,
		`SELECT `+tariffCols+` FROM tariffs WHERE transport_type::text = $1 AND is_active`, transport)
	var tf Tariff
	err := row.Scan(&tf.ID, &tf.TransportType, &tf.PricePerKg, &tf.Currency,
		&tf.VolumetricDivisor, &tf.MinChargeableKg, &tf.IsActive)
	return tf, err
}

// GetByID — тариф по id.
func (t *Tariffs) GetByID(ctx context.Context, id int64) (Tariff, error) {
	row := t.db.QueryRow(ctx, `SELECT `+tariffCols+` FROM tariffs WHERE id = $1`, id)
	var tf Tariff
	err := row.Scan(&tf.ID, &tf.TransportType, &tf.PricePerKg, &tf.Currency,
		&tf.VolumetricDivisor, &tf.MinChargeableKg, &tf.IsActive)
	return tf, err
}

// Update правит тариф (nil — не менять); деактивация снимает деферред-уникальность корректно,
// т.к. arbiter используется только при INSERT.
func (t *Tariffs) Update(ctx context.Context, id int64, pricePerKg, minChargeable *float64, isActive *bool) (Tariff, error) {
	set := []string{}
	args := []any{id}
	ph := 2
	if pricePerKg != nil {
		set = append(set, fmt.Sprintf("price_per_kg = $%d", ph))
		args = append(args, *pricePerKg)
		ph++
	}
	if minChargeable != nil {
		set = append(set, fmt.Sprintf("min_chargeable_kg = $%d", ph))
		args = append(args, *minChargeable)
		ph++
	}
	if isActive != nil {
		set = append(set, fmt.Sprintf("is_active = $%d", ph))
		args = append(args, *isActive)
		ph++
	}
	if len(set) == 0 {
		return t.GetByID(ctx, id)
	}
	q := `UPDATE tariffs SET ` + joinComma(set) + `, updated_at = now()
		WHERE id = $1
		RETURNING ` + tariffCols

	row := t.db.QueryRow(ctx, q, args...)
	var tf Tariff
	err := row.Scan(&tf.ID, &tf.TransportType, &tf.PricePerKg, &tf.Currency,
		&tf.VolumetricDivisor, &tf.MinChargeableKg, &tf.IsActive)
	return tf, err
}

// joinComma — склеивает SET-выражения.
func joinComma(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += ", "
		}
		out += p
	}
	return out
}
