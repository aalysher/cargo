/** Карточка заказа: статус, таймлайн, треки, детали, кнопка чата. */
import { useEffect, useState } from 'react';
import { Check, Copy, MessageCircle, Truck, ArrowLeft } from 'lucide-react';
import { api } from '../lib/api';
import type { Order } from '../lib/types';
import { LINE_TITLES, STATUS_TITLES } from '../lib/statuses';
import { kg, money } from '../lib/format';
import { copyText } from '../lib/copy';
import { BOT_USERNAME } from '../lib/config';
import { hapticImpact, hapticNotify, isTMA, openBotChat } from '../lib/telegram';
import { toast } from '../components/Toast';
import { Button, Card, ErrorRetry, Skeleton, StatusBadge } from '../components/primitives';
import { Timeline } from '../components/Timeline';
import { Progress } from '../components/Progress';

export function OrderDetail({ id, onBack }: { id: number; onBack?: () => void }) {
  const [order, setOrder] = useState<Order | null>(null);
  const [error, setError] = useState('');
  const [copied, setCopied] = useState<number | null>(null);
  const [cancelling, setCancelling] = useState(false);

  const load = () => {
    setError('');
    setOrder(null);
    api<Order>(`/orders/${id}`)
      .then(setOrder)
      .catch((e) => setError(e.detail));
  };
  useEffect(load, [id]);

  const cancel = async () => {
    if (!order || cancelling) return;
    if (!window.confirm(`Отменить заявку ${order.order_number}?`)) return;
    setCancelling(true);
    try {
      const updated = await api<Order>(`/orders/${id}/cancel`, { method: 'POST' });
      hapticNotify('warning');
      toast('Заявка отменена');
      setOrder({ ...order, ...updated });
    } catch (e) {
      hapticNotify('error');
      toast((e as { detail?: string }).detail ?? 'Не удалось отменить', 'err');
    } finally {
      setCancelling(false);
    }
  };

  if (error) {
    return (
      <div className="px-4 pb-28 pt-4">
        <ErrorRetry detail={error === 'не найдено' ? 'Заказ не найден' : error} onRetry={load} />
      </div>
    );
  }

  if (!order) {
    return (
      <div className="space-y-2 px-4 pb-28 pt-4">
        <Skeleton className="h-24" />
        <Skeleton className="h-40" />
        <Skeleton className="h-28" />
      </div>
    );
  }

  const copyTrack = async (trackId: number, num: string) => {
    if (await copyText(num)) {
      hapticImpact('light');
      setCopied(trackId);
      setTimeout(() => setCopied(null), 1000);
    }
  };

  return (
    <div className="space-y-3 px-4 pb-28 pt-4">
      {/* Шапка */}
      <div>
        <div className="flex items-center gap-2">
          {!isTMA() && onBack && (
            <button onClick={onBack} className="rounded-full p-1.5 text-ink-muted transition-colors active:bg-surface-2" aria-label="Назад">
              <ArrowLeft size={20} />
            </button>
          )}
          <div className="text-caption text-ink-muted">Заказ {order.order_number}</div>
        </div>
        <div className="flex items-center justify-between">
          <h1 className="text-display">{STATUS_TITLES[order.status]}</h1>
        </div>
        <div className="mt-1.5 flex items-center gap-2">
          <StatusBadge status={order.status} />
          <span className="text-caption text-ink-muted">{LINE_TITLES[order.transport_type]}</span>
        </div>
      </div>

      {/* Прогресс (для не-финальных) */}
      {order.status !== 'delivered' && order.status !== 'cancelled' && (
        <Progress status={order.status} />
      )}

      {/* Таймлайн */}
      {order.events && order.events.length > 0 && (
        <Card className="p-4">
          <h2 className="mb-3 text-h2">История</h2>
          <Timeline events={order.events} />
        </Card>
      )}

      {/* Треки */}
      {order.tracks && order.tracks.length > 0 && (
        <Card className="p-4">
          <h2 className="mb-2 text-h2">Трек-номера</h2>
          <div className="space-y-1.5">
            {order.tracks.map((t) => (
              <button
                key={t.id}
                onClick={() => copyTrack(t.id, t.tracking_number)}
                className="flex w-full items-center justify-between rounded-md bg-surface-2 px-3 py-2.5 text-left transition-colors active:bg-surface-2/70"
              >
                <span className="tnum text-body">{t.tracking_number}</span>
                {copied === t.id ? (
                  <Check size={16} className="text-success" />
                ) : (
                  <Copy size={16} className="text-ink-muted" />
                )}
              </button>
            ))}
          </div>
        </Card>
      )}

      {/* Детали */}
      <Card className="p-4">
        <h2 className="mb-2 text-h2">Детали</h2>
        <dl className="space-y-1.5 text-body">
          <Row label="Груз" value={order.contents} />
          <Row label="Заявленный вес" value={kg(order.declared_weight_kg)} />
          {order.actual_weight_kg != null && <Row label="Фактический вес" value={kg(order.actual_weight_kg)} />}
          <Row label="Оценка доставки" value={money(order.estimated_cost, order.currency)} />
          {order.final_cost != null && <Row label="К оплате" value={money(order.final_cost, order.currency)} strong />}
          {order.comment && <Row label="Комментарий" value={order.comment} />}
        </dl>
      </Card>

      {/* Чат по заказу */}
      <Button variant="secondary" onClick={() => openBotChat(BOT_USERNAME)}>
        <MessageCircle size={18} className="mr-2" /> Написать по этому заказу
      </Button>

      {/* Отмена — только до подтверждения оператором */}
      {(order.status === 'new' || order.status === 'accepted') && (
        <Button variant="danger" onClick={cancel} disabled={cancelling}>
          {cancelling ? 'Отменяем…' : 'Отменить заявку'}
        </Button>
      )}
      <p className="text-center text-micro text-ink-muted">
        Сообщение уйдёт оператору — мы увидим, что оно про заказ {order.order_number}
      </p>
    </div>
  );
}

function Row({ label, value, strong = false }: { label: string; value: string; strong?: boolean }) {
  return (
    <div className="flex items-baseline justify-between gap-3">
      <dt className="shrink-0 text-ink-muted">{label}</dt>
      <dd className={`text-right ${strong ? 'text-body-strong' : ''}`}>{value}</dd>
    </div>
  );
}

export const TruckIcon = Truck; // re-export для возможных карточек
