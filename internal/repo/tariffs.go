package repo

import (
	"context"
)

// Tariff — строка справочника тарифов.
type Tariff struct {
	ID                int64   `json:"id"`
	TransportType     string  `json:"transport_type"`
	PricePerKg        float64 `json:"price_per_kg"`
	Currency          string  `json:"currency"`
	VolumetricDivisor int     `json:"volumetric_divisor"`
	MinChargeableKg   float64 `json:"min_chargeable_kg"`
}

// tariffCols — колонки для выборки тарифа.
const tariffCols = `id, transport_type::text, price_per_kg, currency, volumetric_divisor, min_chargeable_kg`

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
			&tf.VolumetricDivisor, &tf.MinChargeableKg); err != nil {
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
		&tf.VolumetricDivisor, &tf.MinChargeableKg)
	return tf, err
}
