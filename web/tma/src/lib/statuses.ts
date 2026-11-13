/** Статусы заказа и их русские названия (зеркало internal/domain). */
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
  new: 'Заявка создана',
  accepted: 'Заявка принята в работу',
  at_warehouse_cn: 'Получено на складе в Китае',
  shipped: 'Отправлено в Бишкек',
  in_transit: 'В пути',
  arrived_kg: 'Прибыло в Бишкек',
  ready_for_pickup: 'Готово к выдаче',
  delivered: 'Выдано',
  cancelled: 'Отменено',
};

/** Порядок статусов для прогресс-бара (8 сегментов, без cancelled). */
export const STATUS_ORDER: OrderStatus[] = [
  'new',
  'accepted',
  'at_warehouse_cn',
  'shipped',
  'in_transit',
  'arrived_kg',
  'ready_for_pickup',
  'delivered',
];

/** Прогресс 1..8; для cancelled — 0. */
export function statusProgress(s: OrderStatus): number {
  const i = STATUS_ORDER.indexOf(s);
  return i < 0 ? 0 : i + 1;
}

export const LINE_TITLES: Record<'road' | 'air', string> = {
  road: 'Авто (Китай → Бишкек)',
  air: 'Авиа (Китай → Бишкек)',
};

/** Фильтры списка заказов — скоупы клиентского API. */
export const SCOPES = [
  { key: 'active', label: 'Активные' },
  { key: 'delivering', label: 'В пути' },
  { key: 'done', label: 'Выданы' },
  { key: 'cancelled', label: 'Отменённые' },
  { key: 'all', label: 'Все' },
] as const;
