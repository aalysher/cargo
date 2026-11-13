package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"cargo/internal/domain"
	"cargo/internal/repo"
)

// scopeStatuses — фильтры списка заказов.
var scopeStatuses = map[string][]string{
	"active":     {"new", "accepted"},
	"delivering": {"at_warehouse_cn", "shipped", "in_transit", "arrived_kg", "ready_for_pickup"},
	"done":       {"delivered"},
	"cancelled":  {"cancelled"},
	"all":        nil,
}

// listLimit — размер страницы списка.
const listLimit = 20

// CreateOrderInput — данные заявки от клиента.
type CreateOrderInput struct {
	TransportType    string   `json:"transport_type"`
	Contents         string   `json:"contents"`
	DeclaredWeightKg float64  `json:"declared_weight_kg"`
	Tracks           []string `json:"tracks"`
	Comment          string   `json:"comment"`
}

// CreateOrder создаёт заявку: заказ + треки + событие + уведомление (одна транзакция).
func (s *Service) CreateOrder(ctx context.Context, clientID int64, in CreateOrderInput) (repo.Order, error) {
	in.TransportType = strings.ToLower(strings.TrimSpace(in.TransportType))
	if in.TransportType != "road" && in.TransportType != "air" {
		return repo.Order{}, domain.ValidationErr("укажите тип доставки: road или air")
	}
	in.Contents = strings.TrimSpace(in.Contents)
	if in.Contents == "" || len(in.Contents) > 200 {
		return repo.Order{}, domain.ValidationErr("опишите груз (1–200 символов)")
	}
	if in.DeclaredWeightKg <= 0 || in.DeclaredWeightKg > 1000 {
		return repo.Order{}, domain.ValidationErr("вес должен быть от 0.1 до 1000 кг")
	}
	if len(in.Tracks) > 10 {
		return repo.Order{}, domain.ValidationErr("не больше 10 трек-номеров")
	}
	in.Comment = strings.TrimSpace(in.Comment)
	if len(in.Comment) > 500 {
		return repo.Order{}, domain.ValidationErr("комментарий — до 500 символов")
	}

	tf, err := s.Tariffs.GetActiveByType(ctx, in.TransportType)
	if errors.Is(err, pgx.ErrNoRows) {
		return repo.Order{}, domain.ValidationErr("тариф недоступен, попробуйте позже")
	}
	if err != nil {
		return repo.Order{}, err
	}
	chargeable := domain.ChargeableWeight(in.DeclaredWeightKg, tf.MinChargeableKg)
	est := domain.Estimate(tf.PricePerKg, chargeable)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return repo.Order{}, err
	}
	defer tx.Rollback(ctx)

	var seq int64
	if err := tx.QueryRow(ctx, `SELECT nextval('order_number_seq')`).Scan(&seq); err != nil {
		return repo.Order{}, err
	}
	number := fmt.Sprintf("KG-%s-%04d", time.Now().Format("060102"), seq)

	ord, err := s.Orders.Insert(ctx, tx, repo.Order{
		OrderNumber:      number,
		ClientID:         clientID,
		TransportType:    in.TransportType,
		Contents:         in.Contents,
		DeclaredWeightKg: in.DeclaredWeightKg,
		EstimatedCost:    &est,
		Currency:         tf.Currency,
		Comment:          in.Comment,
	})
	if err != nil {
		return repo.Order{}, err
	}

	for _, tr := range in.Tracks {
		tr = strings.TrimSpace(tr)
		if tr == "" || len(tr) > 64 {
			continue
		}
		// дубликат трека игнорируем (pgx.ErrNoRows при DO NOTHING)
		if _, err := s.Orders.AddTrack(ctx, tx, ord.ID, tr); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return repo.Order{}, err
		}
	}

	if err := s.Events.Insert(ctx, tx, repo.OrderEvent{
		OrderID:   ord.ID,
		EventType: "created",
		Status:    string(domain.StatusNew),
		Title:     domain.StatusTitle(domain.StatusNew),
		ActorType: "client",
		ActorID:   clientID,
	}); err != nil {
		return repo.Order{}, err
	}

	if err := s.notify(ctx, tx, ord,
		"✅ Заявка "+ord.OrderNumber+" создана. Оператор проверит её и возьмёт в работу."); err != nil {
		return repo.Order{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return repo.Order{}, err
	}
	return ord, nil
}

// ListOrders — список заказов клиента по скоупу с курсорной пагинацией.
func (s *Service) ListOrders(ctx context.Context, clientID int64, scope string, cursor int64) ([]repo.Order, int64, error) {
	statuses, ok := scopeStatuses[scope]
	if !ok {
		return nil, 0, domain.ValidationErr("неизвестный scope")
	}
	return s.Orders.ListByClient(ctx, clientID, statuses, cursor, listLimit)
}

// GetOrder — деталь заказа с треками и таймлайном событий.
func (s *Service) GetOrder(ctx context.Context, clientID, id int64) (repo.Order, []repo.OrderTrack, []repo.OrderEvent, error) {
	ord, err := s.Orders.GetForClient(ctx, id, clientID)
	if errors.Is(err, pgx.ErrNoRows) {
		return repo.Order{}, nil, nil, domain.ErrNotFound
	}
	if err != nil {
		return repo.Order{}, nil, nil, err
	}
	tracks, err := s.Orders.TracksByOrder(ctx, ord.ID)
	if err != nil {
		return repo.Order{}, nil, nil, err
	}
	events, err := s.Events.ListForOrder(ctx, ord.ID)
	if err != nil {
		return repo.Order{}, nil, nil, err
	}
	return ord, tracks, events, nil
}

// CancelOrder — отмена клиентом (только new/accepted).
func (s *Service) CancelOrder(ctx context.Context, clientID, orderID int64) (repo.Order, error) {
	ord, err := s.Orders.GetForClient(ctx, orderID, clientID)
	if errors.Is(err, pgx.ErrNoRows) {
		return repo.Order{}, domain.ErrNotFound
	}
	if err != nil {
		return repo.Order{}, err
	}
	if !domain.ClientCanCancel(domain.OrderStatus(ord.Status)) {
		return repo.Order{}, fmt.Errorf(
			"%w: заявку можно отменить только до подтверждения оператором", domain.ErrInvalidTransition)
	}
	return s.ChangeStatus(ctx, ord, domain.StatusCancelled, "client", clientID, "")
}

// ChangeStatus — единственная точка смены статуса: валидация перехода,
// UPDATE с защитой от гонок, событие и уведомление в одной транзакции.
func (s *Service) ChangeStatus(ctx context.Context, ord repo.Order, to domain.OrderStatus,
	actorType string, actorID int64, note string) (repo.Order, error) {

	if !domain.CanTransition(domain.OrderStatus(ord.Status), to) {
		return repo.Order{}, fmt.Errorf("%w: из «%s» в «%s»", domain.ErrInvalidTransition, ord.Status, to)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return repo.Order{}, err
	}
	defer tx.Rollback(ctx)

	updated, err := s.Orders.UpdateStatus(ctx, tx, ord.ID, ord.Status, string(to))
	if errors.Is(err, pgx.ErrNoRows) {
		return repo.Order{}, fmt.Errorf("%w: заказ уже изменён", domain.ErrInvalidTransition)
	}
	if err != nil {
		return repo.Order{}, err
	}

	title := domain.StatusTitle(to)
	if err := s.Events.Insert(ctx, tx, repo.OrderEvent{
		OrderID:   ord.ID,
		EventType: "status_change",
		Status:    string(to),
		Title:     title,
		Payload:   notePayload(note),
		ActorType: actorType,
		ActorID:   actorID,
	}); err != nil {
		return repo.Order{}, err
	}

	if err := s.notify(ctx, tx, ord, "📦 Заказ "+ord.OrderNumber+" — "+title); err != nil {
		return repo.Order{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return repo.Order{}, err
	}
	return updated, nil
}

// notify кладёт сообщение клиенту в outbox.
func (s *Service) notify(ctx context.Context, tx repo.DBTX, ord repo.Order, text string) error {
	payload, err := json.Marshal(map[string]string{
		"text":         text,
		"order_number": ord.OrderNumber,
	})
	if err != nil {
		return err
	}
	return s.Notifications.InsertPending(ctx, tx, ord.ClientID, &ord.ID, "status_change", string(payload))
}

// notePayload — JSON события с комментарием оператора (или пустая строка).
func notePayload(note string) string {
	if note == "" {
		return ""
	}
	b, _ := json.Marshal(map[string]string{"note": note})
	return string(b)
}
