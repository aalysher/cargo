/** Типы админского API. */

export interface AdminClient {
  id: number;
  first_name: string;
  last_name: string;
  username: string;
  phone: string;
  code: string;
}

export interface AdminOrder {
  id: number;
  order_number: string;
  status: string;
  transport_type: 'road' | 'air';
  contents: string;
  declared_weight_kg: number;
  actual_weight_kg: number | null;
  chargeable_weight_kg: number | null;
  estimated_cost: number | null;
  final_cost: number | null;
  currency: string;
  comment: string;
  created_at: string;
  updated_at: string;
  client: AdminClient;
  tracks?: Track[];
  events?: OrderEvent[];
}

export interface Track {
  id: number;
  tracking_number: string;
}

export interface AdminOrderList {
  items: AdminOrder[];
  next_cursor: number;
}

export interface Tariff {
  id: number;
  transport_type: 'road' | 'air';
  price_per_kg: number;
  currency: string;
  min_chargeable_kg: number;
  is_active: boolean;
}

export interface AdminClientFull extends AdminClient {
  telegram_id: number;
  orders_count: number;
  orders?: ClientOrder[];
}

export interface ClientOrder {
  id: number;
  order_number: string;
  status: string;
  declared_weight_kg: number;
  estimated_cost: number | null;
  final_cost: number | null;
  currency: string;
  created_at: string;
}

export type OrderEvent = {
  type: string;
  status?: string;
  title: string;
  created_at: string;
};
