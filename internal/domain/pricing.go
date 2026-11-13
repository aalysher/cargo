package domain

import "math"

// Tariff — тариф (строка справочника tariffs).
type Tariff struct {
	ID                int64
	TransportType     string
	PricePerKg        float64
	Currency          string
	VolumetricDivisor int
	MinChargeableKg   float64
}

// ChargeableWeight — расчётный вес: не ниже минимального тарифа.
// MVP: тарифицируем фактический вес; объёмный вес включаем в v2 (VolumetricWeight).
func ChargeableWeight(grossKg, minKg float64) float64 {
	if grossKg < minKg {
		return minKg
	}
	return grossKg
}

// VolumetricWeight — объёмный вес по габаритам в см (задел на v2).
func VolumetricWeight(length, width, height, divisor float64) float64 {
	if divisor <= 0 || length <= 0 || width <= 0 || height <= 0 {
		return 0
	}
	return round2(length * width * height / divisor)
}

// Estimate — предварительная стоимость: ставка × расчётный вес, до копеек.
func Estimate(pricePerKg, chargeableKg float64) float64 {
	return round2(pricePerKg * chargeableKg)
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
