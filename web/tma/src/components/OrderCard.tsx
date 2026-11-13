/** Карточка заказа в списке: номер, дата, вес, бейдж, мини-прогресс, chevron. */
import { ChevronRight } from 'lucide-react';
import type { Order } from '../lib/types';
import { dateShort, kg, money } from '../lib/format';
import { hapticImpact } from '../lib/telegram';
import { Progress } from './Progress';
import { StatusBadge } from './primitives';

export function OrderCard({ order, onClick }: { order: Order; onClick: () => void }) {
  const weight = order.actual_weight_kg ?? order.declared_weight_kg;
  const cost = order.final_cost ?? order.estimated_cost;

  return (
    <button
      onClick={() => {
        hapticImpact('light');
        onClick();
      }}
      className="w-full animate-fade-up rounded-lg border border-line bg-surface p-4 text-left shadow-1 transition-transform active:scale-[.98]"
    >
      <div className="flex items-start justify-between gap-2">
        <div>
          <div className="tnum text-body-strong">{order.order_number}</div>
          <div className="text-caption text-ink-muted">
            {dateShort(order.created_at)} · {kg(weight)} · {money(cost, order.currency)}
          </div>
        </div>
        <ChevronRight size={18} className="mt-1 shrink-0 text-ink-muted" />
      </div>
      <div className="mt-2.5 flex items-center gap-3">
        <div className="flex-1">
          <Progress status={order.status} compact />
        </div>
        <StatusBadge status={order.status} />
      </div>
    </button>
  );
}
