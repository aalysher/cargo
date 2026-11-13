package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"cargo/internal/domain"
	"cargo/internal/repo"
)

// ListTariffs — активные тарифы для калькулятора.
func (s *Service) ListTariffs(ctx context.Context) ([]repo.Tariff, error) {
	return s.Tariffs.ListActive(ctx)
}

// EstimateResult — расчёт калькулятора.
type EstimateResult struct {
	TransportType string  `json:"transport_type"`
	DeclaredKg    float64 `json:"declared_kg"`
	ChargeableKg  float64 `json:"chargeable_kg"`
	PricePerKg    float64 `json:"price_per_kg"`
	EstimatedCost float64 `json:"estimated_cost"`
	Currency      string  `json:"currency"`
}

// Estimate — расчёт стоимости по тарифу (единая формула бэка и фронта).
func (s *Service) Estimate(ctx context.Context, transport string, weightKg float64) (EstimateResult, error) {
	tf, err := s.Tariffs.GetActiveByType(ctx, transport)
	if errors.Is(err, pgx.ErrNoRows) {
		return EstimateResult{}, domain.ValidationErr("неизвестный тип доставки: road или air")
	}
	if err != nil {
		return EstimateResult{}, err
	}
	if weightKg <= 0 || weightKg > 1000 {
		return EstimateResult{}, domain.ValidationErr("вес должен быть от 0.1 до 1000 кг")
	}
	chargeable := domain.ChargeableWeight(weightKg, tf.MinChargeableKg)
	return EstimateResult{
		TransportType: tf.TransportType,
		DeclaredKg:    weightKg,
		ChargeableKg:  chargeable,
		PricePerKg:    tf.PricePerKg,
		EstimatedCost: domain.Estimate(tf.PricePerKg, chargeable),
		Currency:      tf.Currency,
	}, nil
}
