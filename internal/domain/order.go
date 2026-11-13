// Package domain: бизнес-правила, не зависящие от HTTP и БД.
package domain

// OrderStatus — жизненный цикл заказа (enum в БД: order_status).
type OrderStatus string

const (
	StatusNew            OrderStatus = "new"
	StatusAccepted       OrderStatus = "accepted"
	StatusAtWarehouseCN  OrderStatus = "at_warehouse_cn"
	StatusShipped        OrderStatus = "shipped"
	StatusInTransit      OrderStatus = "in_transit"
	StatusArrivedKG      OrderStatus = "arrived_kg"
	StatusReadyForPickup OrderStatus = "ready_for_pickup"
	StatusDelivered      OrderStatus = "delivered"
	StatusCancelled      OrderStatus = "cancelled"
)

// transitions — карта допустимых переходов «только вперёд».
var transitions = map[OrderStatus][]OrderStatus{
	StatusNew:            {StatusAccepted, StatusCancelled},
	StatusAccepted:       {StatusAtWarehouseCN, StatusCancelled},
	StatusAtWarehouseCN:  {StatusShipped, StatusCancelled},
	StatusShipped:        {StatusInTransit},
	StatusInTransit:      {StatusArrivedKG},
	StatusArrivedKG:      {StatusReadyForPickup},
	StatusReadyForPickup: {StatusDelivered, StatusCancelled},
	StatusDelivered:      {},
	StatusCancelled:      {},
}

// CanTransition проверяет допустимость перехода статуса.
func CanTransition(from, to OrderStatus) bool {
	for _, s := range transitions[from] {
		if s == to {
			return true
		}
	}
	return false
}

// ClientCanCancel — клиент может отменить заявку до подтверждения оператором.
func ClientCanCancel(s OrderStatus) bool {
	return s == StatusNew || s == StatusAccepted
}

// StatusTitle — человекочитаемый заголовок статуса (для UI и уведомлений).
func StatusTitle(s OrderStatus) string {
	if t, ok := statusTitles[s]; ok {
		return t
	}
	return string(s)
}

var statusTitles = map[OrderStatus]string{
	StatusNew:            "Заявка создана",
	StatusAccepted:       "Заявка принята в работу",
	StatusAtWarehouseCN:  "Получено на складе в Китае",
	StatusShipped:        "Отправлено в Бишкек",
	StatusInTransit:      "В пути",
	StatusArrivedKG:      "Прибыло в Бишкек",
	StatusReadyForPickup: "Готово к выдаче",
	StatusDelivered:      "Выдано",
	StatusCancelled:      "Отменено",
}

// statusOrder — порядок статусов для прогресса на UI (без cancelled).
var statusOrder = []OrderStatus{
	StatusNew, StatusAccepted, StatusAtWarehouseCN, StatusShipped,
	StatusInTransit, StatusArrivedKG, StatusReadyForPickup, StatusDelivered,
}

// StatusProgress — позиция статуса в таймлайне (0..7), для прогресс-бара.
func StatusProgress(s OrderStatus) int {
	for i, st := range statusOrder {
		if st == s {
			return i
		}
	}
	return 0
}
