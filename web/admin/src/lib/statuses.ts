/** Статусы/переходы — зеркало internal/domain ( переходы только вперёд). */
export type OrderStatus =
  | 'new'
  | 'accepted'
  | 'at_warehouse_cn'
  | 'shipped'
  | 'in_transit'
  | 'arrived_kg'
  | 'ready_for_pickup'
  | 'delivered'
  | 'cancelled';

export const STATUS_TITLES: Record<OrderStatus, string> = {
  new: 'Новая',
  accepted: 'Принята',
  at_warehouse_cn: 'На складе (Китай)',
  shipped: 'Отгружено',
  in_transit: 'В пути',
  arrived_kg: 'Прибыло в Бишкек',
  ready_for_pickup: 'Готово к выдаче',
  delivered: 'Выдано',
  cancelled: 'Отменено',
};

export const ALL_STATUSES = Object.keys(STATUS_TITLES) as OrderStatus[];

/** Допустимые следующие статусы (mirror domain.transitions). */
export const NEXT_STATUSES: Record<OrderStatus, OrderStatus[]> = {
  new: ['accepted', 'cancelled'],
  accepted: ['at_warehouse_cn', 'cancelled'],
  at_warehouse_cn: ['shipped', 'cancelled'],
  shipped: ['in_transit'],
  in_transit: ['arrived_kg'],
  arrived_kg: ['ready_for_pickup'],
  ready_for_pickup: ['delivered', 'cancelled'],
  delivered: [],
  cancelled: [],
};

export const LINE_TITLES: Record<'road' | 'air', string> = {
  road: 'Авто',
  air: 'Авиа',
};
