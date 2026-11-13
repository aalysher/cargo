package domain

import "testing"

func TestCanTransition(t *testing.T) {
	valid := [][2]OrderStatus{
		{StatusNew, StatusAccepted},
		{StatusAccepted, StatusAtWarehouseCN},
		{StatusAtWarehouseCN, StatusShipped},
		{StatusShipped, StatusInTransit},
		{StatusInTransit, StatusArrivedKG},
		{StatusArrivedKG, StatusReadyForPickup},
		{StatusReadyForPickup, StatusDelivered},
		{StatusNew, StatusCancelled},
		{StatusAccepted, StatusCancelled},
		{StatusAtWarehouseCN, StatusCancelled},
		{StatusReadyForPickup, StatusCancelled},
	}
	for _, p := range valid {
		if !CanTransition(p[0], p[1]) {
			t.Errorf("ожидается допустимый переход %s -> %s", p[0], p[1])
		}
	}

	invalid := [][2]OrderStatus{
		{StatusNew, StatusDelivered},
		{StatusShipped, StatusAtWarehouseCN},
		{StatusDelivered, StatusNew},
		{StatusCancelled, StatusAccepted},
		{StatusInTransit, StatusCancelled},
		{StatusNew, StatusShipped},
		{StatusAccepted, StatusReadyForPickup},
	}
	for _, p := range invalid {
		if CanTransition(p[0], p[1]) {
			t.Errorf("ожидается недопустимый переход %s -> %s", p[0], p[1])
		}
	}
}

func TestClientCanCancel(t *testing.T) {
	if !ClientCanCancel(StatusNew) || !ClientCanCancel(StatusAccepted) {
		t.Fatal("клиент должен мочь отменить new/accepted")
	}
	if ClientCanCancel(StatusShipped) {
		t.Fatal("клиент не должен мочь отменить shipped")
	}
}

func TestPricing(t *testing.T) {
	if got := ChargeableWeight(0.5, 1); got != 1 {
		t.Fatalf("минимальный расчётный вес: got %v, want 1", got)
	}
	if got := ChargeableWeight(4.5, 1); got != 4.5 {
		t.Fatalf("расчётный вес: got %v, want 4.5", got)
	}
	if got := Estimate(2.90, 4.5); got != 13.05 {
		t.Fatalf("расчёт стоимости: got %v, want 13.05", got)
	}
	if got := VolumetricWeight(40, 30, 25, 5000); got != 6 {
		t.Fatalf("объёмный вес: got %v, want 6", got)
	}
}

func TestStatusTitle(t *testing.T) {
	if StatusTitle(StatusReadyForPickup) != "Готово к выдаче" {
		t.Fatalf("неожиданный заголовок статуса: %s", StatusTitle(StatusReadyForPickup))
	}
}
