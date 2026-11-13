package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"cargo/internal/auth"
	"cargo/internal/domain"
	"cargo/internal/repo"
)

// AdminLogin проверяет логин/пароль и выпускает JWT оператора.
func (s *Service) AdminLogin(ctx context.Context, username, password string) (repo.AdminUser, string, error) {
	u, err := s.Admins.GetByUsername(ctx, username)
	if errors.Is(err, pgx.ErrNoRows) {
		return repo.AdminUser{}, "", domain.ErrUnauthorized
	}
	if err != nil {
		return repo.AdminUser{}, "", err
	}
	if !u.IsActive {
		return repo.AdminUser{}, "", domain.ErrUnauthorized
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return repo.AdminUser{}, "", domain.ErrUnauthorized
	}
	token, err := auth.NewAdminToken(s.cfg.JWTSecret, u.ID, u.Role)
	if err != nil {
		return repo.AdminUser{}, "", err
	}
	return u, token, nil
}

// ListAdminOrders — таблица заказов: фильтр по статусу, поиск, пагинация.
func (s *Service) ListAdminOrders(ctx context.Context, status, search string, cursor int64) ([]repo.AdminOrder, int64, error) {
	if status != "" && !validStatus(status) {
		return nil, 0, domain.ValidationErr("неизвестный статус: " + status)
	}
	return s.AdminOrders.ListAdmin(ctx, status, search, cursor, listLimit)
}

// validStatus — проверка enum-значения статуса.
func validStatus(s string) bool {
	switch domain.OrderStatus(s) {
	case domain.StatusNew, domain.StatusAccepted, domain.StatusAtWarehouseCN,
		domain.StatusShipped, domain.StatusInTransit, domain.StatusArrivedKG,
		domain.StatusReadyForPickup, domain.StatusDelivered, domain.StatusCancelled:
		return true
	}
	return false
}

// GetAdminOrder — деталь заказа с клиентом, треками и таймлайном.
func (s *Service) GetAdminOrder(ctx context.Context, id int64) (repo.AdminOrder, []repo.OrderTrack, []repo.OrderEvent, error) {
	ao, err := s.AdminOrders.GetAdminByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return repo.AdminOrder{}, nil, nil, domain.ErrNotFound
	}
	if err != nil {
		return repo.AdminOrder{}, nil, nil, err
	}
	tracks, err := s.Orders.TracksByOrder(ctx, ao.ID)
	if err != nil {
		return repo.AdminOrder{}, nil, nil, err
	}
	events, err := s.Events.ListForOrder(ctx, ao.ID)
	if err != nil {
		return repo.AdminOrder{}, nil, nil, err
	}
	return ao, tracks, events, nil
}

// SetOrderWeight — фактический вес на складе: пересчёт стоимости + событие.
func (s *Service) SetOrderWeight(ctx context.Context, orderID int64, weightKg float64, actorID int64) (repo.AdminOrder, error) {
	if weightKg <= 0 || weightKg > 1000 {
		return repo.AdminOrder{}, domain.ValidationErr("вес должен быть от 0.1 до 1000 кг")
	}
	ao, err := s.AdminOrders.GetAdminByID(ctx, orderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return repo.AdminOrder{}, domain.ErrNotFound
	}
	if err != nil {
		return repo.AdminOrder{}, err
	}
	tf, err := s.Tariffs.GetActiveByType(ctx, ao.TransportType)
	if errors.Is(err, pgx.ErrNoRows) {
		return repo.AdminOrder{}, domain.ValidationErr("нет активного тарифа для линии " + ao.TransportType)
	}
	if err != nil {
		return repo.AdminOrder{}, err
	}

	chargeable := domain.ChargeableWeight(weightKg, tf.MinChargeableKg)
	est := domain.Estimate(tf.PricePerKg, chargeable)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return repo.AdminOrder{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := s.AdminOrders.UpdateWeight(ctx, tx, ao.ID, weightKg, chargeable, est); err != nil {
		return repo.AdminOrder{}, err
	}
	if err := s.Events.Insert(ctx, tx, repo.OrderEvent{
		OrderID:   ao.ID,
		EventType: "weight_set",
		Title:     fmt.Sprintf("Вес установлен: %.2f кг", weightKg),
		Payload:   fmt.Sprintf(`{"weight_kg":%.2f,"chargeable_kg":%.2f}`, weightKg, chargeable),
		ActorType: "operator",
		ActorID:   actorID,
	}); err != nil {
		return repo.AdminOrder{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return repo.AdminOrder{}, err
	}
	return s.AdminOrders.GetAdminByID(ctx, ao.ID)
}

// UpdateOrderMeta — финальная стоимость и комментарий оператора.
func (s *Service) UpdateOrderMeta(ctx context.Context, orderID int64,
	finalCost *float64, comment *string, actorID int64) (repo.AdminOrder, error) {

	ao, err := s.AdminOrders.GetAdminByID(ctx, orderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return repo.AdminOrder{}, domain.ErrNotFound
	}
	if err != nil {
		return repo.AdminOrder{}, err
	}
	if comment != nil && len(*comment) > 500 {
		return repo.AdminOrder{}, domain.ValidationErr("комментарий — до 500 символов")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return repo.AdminOrder{}, err
	}
	defer tx.Rollback(ctx)

	if finalCost != nil {
		if *finalCost < 0 {
			return repo.AdminOrder{}, domain.ValidationErr("стоимость не может быть отрицательной")
		}
		if _, err := s.AdminOrders.UpdateFinalCost(ctx, tx, ao.ID, *finalCost); err != nil {
			return repo.AdminOrder{}, err
		}
		if err := s.Events.Insert(ctx, tx, repo.OrderEvent{
			OrderID:   ao.ID,
			EventType: "note",
			Title:     fmt.Sprintf("Финальная стоимость: %.2f %s", *finalCost, ao.Currency),
			ActorType: "operator",
			ActorID:   actorID,
		}); err != nil {
			return repo.AdminOrder{}, err
		}
	}
	if comment != nil {
		if _, err := s.AdminOrders.UpdateComment(ctx, tx, ao.ID, *comment); err != nil {
			return repo.AdminOrder{}, err
		}
		if err := s.Events.Insert(ctx, tx, repo.OrderEvent{
			OrderID:   ao.ID,
			EventType: "note",
			Title:     "Комментарий оператора обновлён",
			ActorType: "operator",
			ActorID:   actorID,
		}); err != nil {
			return repo.AdminOrder{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return repo.AdminOrder{}, err
	}
	return s.AdminOrders.GetAdminByID(ctx, ao.ID)
}

// AdminChangeStatus — смена статуса оператором (валидация + событие + уведомление).
func (s *Service) AdminChangeStatus(ctx context.Context, orderID int64, to, note string, actorID int64) (repo.AdminOrder, error) {
	if !validStatus(to) {
		return repo.AdminOrder{}, domain.ValidationErr("неизвестный статус: " + to)
	}
	ao, err := s.AdminOrders.GetAdminByID(ctx, orderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return repo.AdminOrder{}, domain.ErrNotFound
	}
	if err != nil {
		return repo.AdminOrder{}, err
	}

	updated, err := s.ChangeStatus(ctx, ao.Order, domain.OrderStatus(to), "operator", actorID, note)
	if err != nil {
		return repo.AdminOrder{}, err
	}
	return s.AdminOrders.GetAdminByID(ctx, updated.ID)
}

// AdminAddTrack — привязать трек покупки (событие track_added).
func (s *Service) AdminAddTrack(ctx context.Context, orderID int64, tracking string, actorID int64) (repo.AdminOrder, error) {
	tracking = strings.TrimSpace(tracking)
	if tracking == "" || len(tracking) > 64 {
		return repo.AdminOrder{}, domain.ValidationErr("трек-номер — 1..64 символов")
	}
	ao, err := s.AdminOrders.GetAdminByID(ctx, orderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return repo.AdminOrder{}, domain.ErrNotFound
	}
	if err != nil {
		return repo.AdminOrder{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return repo.AdminOrder{}, err
	}
	defer tx.Rollback(ctx)

	_, err = s.Orders.AddTrack(ctx, tx, ao.ID, tracking)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return repo.AdminOrder{}, err
	}
	if err == nil { // дубликат не логируем как событие
		if err := s.Events.Insert(ctx, tx, repo.OrderEvent{
			OrderID:   ao.ID,
			EventType: "track_added",
			Title:     "Добавлен трек: " + tracking,
			Payload:   fmt.Sprintf(`{"tracking_number":%q}`, tracking),
			ActorType: "operator",
			ActorID:   actorID,
		}); err != nil {
			return repo.AdminOrder{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return repo.AdminOrder{}, err
	}
	return s.AdminOrders.GetAdminByID(ctx, ao.ID)
}

// AdminRemoveTrack — отвязать трек покупки (событие track_removed).
func (s *Service) AdminRemoveTrack(ctx context.Context, orderID, trackID int64, actorID int64) (repo.AdminOrder, error) {
	ao, err := s.AdminOrders.GetAdminByID(ctx, orderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return repo.AdminOrder{}, domain.ErrNotFound
	}
	if err != nil {
		return repo.AdminOrder{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return repo.AdminOrder{}, err
	}
	defer tx.Rollback(ctx)

	tracking, err := s.AdminOrders.DeleteTrack(ctx, tx, ao.ID, trackID)
	if errors.Is(err, pgx.ErrNoRows) {
		return repo.AdminOrder{}, domain.ErrNotFound
	}
	if err != nil {
		return repo.AdminOrder{}, err
	}
	if err := s.Events.Insert(ctx, tx, repo.OrderEvent{
		OrderID:   ao.ID,
		EventType: "track_removed",
		Title:     "Удалён трек: " + tracking,
		Payload:   fmt.Sprintf(`{"tracking_number":%q}`, tracking),
		ActorType: "operator",
		ActorID:   actorID,
	}); err != nil {
		return repo.AdminOrder{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return repo.AdminOrder{}, err
	}
	return s.AdminOrders.GetAdminByID(ctx, ao.ID)
}

// ListAdminTariffs — все тарифы, включая неактивные.
func (s *Service) ListAdminTariffs(ctx context.Context) ([]repo.Tariff, error) {
	return s.Tariffs.ListAll(ctx)
}

// UpdateTariff — правка тарифа (валидация значений).
func (s *Service) UpdateTariff(ctx context.Context, id int64,
	pricePerKg, minChargeable *float64, isActive *bool) (repo.Tariff, error) {

	if pricePerKg != nil && (*pricePerKg <= 0 || *pricePerKg > 10000) {
		return repo.Tariff{}, domain.ValidationErr("ставка за кг: 0.01–10000")
	}
	if minChargeable != nil && (*minChargeable <= 0 || *minChargeable > 1000) {
		return repo.Tariff{}, domain.ValidationErr("минимальный вес: 0.01–1000 кг")
	}
	tf, err := s.Tariffs.Update(ctx, id, pricePerKg, minChargeable, isActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return repo.Tariff{}, domain.ErrNotFound
	}
	return tf, err
}

// ListClientsAdmin — поиск клиентов для оператора.
func (s *Service) ListClientsAdmin(ctx context.Context, search string) ([]repo.Client, error) {
	return s.Clients.ListForAdmin(ctx, search, 50)
}

// GetClientAdmin — клиент + его последние заказы.
func (s *Service) GetClientAdmin(ctx context.Context, id int64) (repo.Client, []repo.Order, error) {
	cl, err := s.Clients.GetByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return repo.Client{}, nil, domain.ErrNotFound
	}
	if err != nil {
		return repo.Client{}, nil, err
	}
	orders, err := s.AdminOrders.OrdersByClient(ctx, id, 20)
	return cl, orders, err
}

// adminSettingsWhitelist — какие настройки доступны через API.
var adminSettingsWhitelist = map[string]bool{"warehouse_address": true}

// GetSettingsAdmin — все настройки (ключ → JSON).
func (s *Service) GetSettingsAdmin(ctx context.Context) (map[string]json.RawMessage, error) {
	out := map[string]json.RawMessage{}
	for key := range adminSettingsWhitelist {
		raw, err := s.Settings.Get(ctx, key)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return nil, err
		}
		out[key] = json.RawMessage(raw)
	}
	return out, nil
}

// PutSettingAdmin — обновить настройку (только ключи из whitelist).
func (s *Service) PutSettingAdmin(ctx context.Context, key string, value json.RawMessage) error {
	if !adminSettingsWhitelist[key] {
		return domain.ValidationErr("неизвестный ключ настройки")
	}
	// warehouse_address валидируем структурой
	if key == "warehouse_address" {
		var wa WarehouseAddress
		if err := json.Unmarshal(value, &wa); err != nil {
			return domain.ValidationErr("формат warehouse_address: {contact, address, phone, city_cn}")
		}
	}
	return s.Settings.Put(ctx, key, string(value))
}
