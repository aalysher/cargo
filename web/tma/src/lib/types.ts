/** Типы ответов клиентского API /api/v1. */
import type { OrderStatus } from './statuses';

export interface Profile {
  id: number;
  telegram_id: number;
  first_name: string;
  last_name: string;
  username: string;
  photo_url: string;
  phone: string;
  client_code: string;
}

export interface Tariff {
  transport_type: 'road' | 'air';
  price_per_kg: number;
  currency: string;
  min_chargeable_kg: number;
}

export interface Estimate {
  transport_type: string;
  declared_kg: number;
  chargeable_kg: number;
  price_per_kg: number;
  estimated_cost: number;
  currency: string;
}

export interface Track {
  id: number;
  tracking_number: string;
}

export interface OrderEvent {
  type: string;
  status?: string;
  title: string;
  created_at: string;
}

export interface Order {
  id: number;
  order_number: string;
  status: OrderStatus;
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
  tracks?: Track[];
  events?: OrderEvent[];
}

export interface OrderList {
  items: Order[];
  next_cursor: number;
}

export interface WarehouseAddress {
  contact: string;
  address: string;
  phone: string;
  city_cn: string;
  client_code: string;
}
