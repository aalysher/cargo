/** Карточка заказа: статусы, вес/цена, треки, клиент, таймлайн. */
import { useEffect, useState } from 'react';
import { ArrowLeft, Plus, Trash2 } from 'lucide-react';
import { api, type ApiError } from '../lib/api';
import type { AdminOrder } from '../lib/types';
import { LINE_TITLES, NEXT_STATUSES, STATUS_TITLES, type OrderStatus } from '../lib/statuses';
import { dateTime, kg, money } from '../lib/format';
import { toast } from '../components/primitives';
import { Button, Card, ErrorBox, Skeleton, StatusBadge } from '../components/primitives';

export function OrderDetail({ id, onBack }: { id: number; onBack: () => void }) {
  const [order, setOrder] = useState<AdminOrder | null>(null);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  // Формы
  const [note, setNote] = useState('');
  const [weight, setWeight] = useState('');
  const [finalCost, setFinalCost] = useState('');
  const [comment, setComment] = useState('');
  const [track, setTrack] = useState('');

  const load = () => {
    setError('');
    setOrder(null);
    api<AdminOrder>(`/admin/orders/${id}`)
      .then((o) => {
        setOrder(o);
        setWeight(o.actual_weight_kg != null ? String(o.actual_weight_kg) : '');
        setFinalCost(o.final_cost != null ? String(o.final_cost) : '');
        setComment(o.comment ?? '');
      })
      .catch((e) => setError(e.detail));
  };
  useEffect(load, [id]);

  const act = async (fn: () => Promise<AdminOrder>, okMsg: string) => {
    if (busy) return;
    setBusy(true);
    try {
      const updated = await fn();
      setOrder(updated);
      toast(okMsg);
      setNote('');
    } catch (e) {
      toast((e as ApiError).detail, 'err');
    } finally {
      setBusy(false);
    }
  };

  if (error) return <ErrorBox detail={error} onRetry={load} />;
  if (!order) return <Skeleton className="h-96" />;

  const changeStatus = (to: OrderStatus) =>
    act(
      () =>
        api<AdminOrder>(`/admin/orders/${id}/status`, {
          method: 'POST',
          body: { status: to, note: note.trim() },
        }),
      `Статус → ${STATUS_TITLES[to]}`,
    );

  const saveWeight = () => {
    const w = parseFloat(weight.replace(',', '.'));
    if (!w || w <= 0) {
      toast('Введите корректный вес', 'err');
      return;
    }
    act(
      () => api<AdminOrder>(`/admin/orders/${id}`, { method: 'PATCH', body: { actual_weight_kg: w } }),
      'Вес сохранён, цена пересчитана',
    );
  };

  const saveMeta = () => {
    const body: Record<string, unknown> = {};
    const fc = finalCost.trim();
    if (fc) body.final_cost = parseFloat(fc.replace(',', '.'));
    if (comment.trim() !== (order.comment ?? '')) body.comment = comment.trim();
    if (Object.keys(body).length === 0) return;
    act(() => api<AdminOrder>(`/admin/orders/${id}`, { method: 'PATCH', body }), 'Сохранено');
  };

  const addTrack = () => {
    const t = track.trim();
    if (!t) return;
    act(
      () => api<AdminOrder>(`/admin/orders/${id}/tracks`, { method: 'POST', body: { tracking_number: t } }),
      'Трек добавлен',
    ).then(() => setTrack(''));
  };

  const removeTrack = (trackId: number) =>
    act(() => api<AdminOrder>(`/admin/orders/${id}/tracks/${trackId}`, { method: 'DELETE' }), 'Трек удалён');

  const nexts = NEXT_STATUSES[order.status as OrderStatus] ?? [];

  return (
    <div className="space-y-3">
      <button onClick={onBack} className="flex items-center gap-1.5 text-caption text-brand">
        <ArrowLeft size={16} /> К списку заказов
      </button>

      {/* Шапка */}
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <div className="text-h1 tnum">{order.order_number}</div>
          <div className="text-caption text-ink-muted">
            {LINE_TITLES[order.transport_type]} · создан {dateTime(order.created_at)}
          </div>
        </div>
        <StatusBadge status={order.status} />
      </div>

      <div className="grid gap-3 lg:grid-cols-2">
        {/* Левая колонка */}
        <div className="space-y-3">
          {/* Статусы */}
          <Card className="p-4">
            <div className="mb-2 text-h2">Смена статуса</div>
            {nexts.length === 0 ? (
              <div className="text-caption text-ink-muted">Финальный статус — переходов нет</div>
            ) : (
              <>
                <input
                  className="field mb-2"
                  placeholder="Комментарий к смене статуса (клиенту в бот)"
                  value={note}
                  onChange={(e) => setNote(e.target.value)}
                  maxLength={200}
                />
                <div className="flex flex-wrap gap-2">
                  {nexts.map((s) =>
                    s === 'cancelled' ? (
                      <Button
                        key={s}
                        variant="danger"
                        disabled={busy}
                        onClick={() => {
                          if (window.confirm(`Отменить заказ ${order.order_number}?`)) changeStatus(s);
                        }}
                      >
                        {STATUS_TITLES[s]}
                      </Button>
                    ) : (
                      <Button key={s} disabled={busy} onClick={() => changeStatus(s)}>
                        {STATUS_TITLES[s]}
                      </Button>
                    ),
                  )}
                </div>
              </>
            )}
          </Card>

          {/* Вес и деньги */}
          <Card className="p-4">
            <div className="mb-2 text-h2">Вес и стоимость</div>
            <dl className="mb-3 space-y-1 text-body">
              <Row label="Заявленный" value={kg(order.declared_weight_kg)} />
              <Row label="Расчётный" value={kg(order.chargeable_weight_kg)} />
              <Row label="Оценка" value={money(order.estimated_cost, order.currency)} />
            </dl>
            <div className="grid grid-cols-2 gap-2">
              <div>
                <label className="text-caption text-ink-muted">Фактический вес, кг</label>
                <input
                  className="field tnum mt-1"
                  inputMode="decimal"
                  value={weight}
                  onChange={(e) => setWeight(e.target.value.replace(/[^\d.,]/g, ''))}
                  disabled={busy}
                />
              </div>
              <div>
                <label className="text-caption text-ink-muted">Финальная стоимость</label>
                <input
                  className="field tnum mt-1"
                  inputMode="decimal"
                  value={finalCost}
                  onChange={(e) => setFinalCost(e.target.value.replace(/[^\d.,]/g, ''))}
                  disabled={busy}
                />
              </div>
            </div>
            <textarea
              className="field mt-2 h-auto min-h-[64px] py-2"
              placeholder="Комментарий оператора (виден клиенту в карточке)"
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              maxLength={500}
              disabled={busy}
            />
            <Button className="mt-2" onClick={saveWeight} disabled={busy}>
              Сохранить вес
            </Button>{' '}
            <Button variant="secondary" onClick={saveMeta} disabled={busy}>
              Сохранить цену/комментарий
            </Button>
          </Card>

          {/* Треки */}
          <Card className="p-4">
            <div className="mb-2 text-h2">Трек-номера</div>
            <div className="space-y-1.5">
              {(order.tracks ?? []).map((t) => (
                <div key={t.id} className="flex items-center justify-between rounded-md bg-surface-2 px-3 py-2">
                  <span className="tnum text-body">{t.tracking_number}</span>
                  <button onClick={() => removeTrack(t.id)} disabled={busy} className="text-ink-muted hover:text-danger">
                    <Trash2 size={16} />
                  </button>
                </div>
              ))}
              {(order.tracks ?? []).length === 0 && <div className="text-caption text-ink-muted">Треков нет</div>}
            </div>
            <div className="mt-2 flex gap-2">
              <input
                className="field"
                placeholder="Трек-номер покупки"
                value={track}
                onChange={(e) => setTrack(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && addTrack()}
                maxLength={64}
                disabled={busy}
              />
              <Button onClick={addTrack} disabled={busy || !track.trim()} className="w-11 px-0">
                <Plus size={18} />
              </Button>
            </div>
          </Card>
        </div>

        {/* Правая колонка */}
        <div className="space-y-3">
          {/* Клиент */}
          <Card className="p-4">
            <div className="mb-2 text-h2">Клиент</div>
            <div className="text-body-strong">
              {[order.client.first_name, order.client.last_name].filter(Boolean).join(' ') || 'Без имени'}
            </div>
            <div className="mt-1 space-y-0.5 text-caption text-ink-muted">
              {order.client.username && <div>@{order.client.username}</div>}
              <div>Телефон: {order.client.phone || '—'}</div>
              <div>
                Личный код: <span className="tnum text-ink">{order.client.code}</span>
              </div>
            </div>
          </Card>

          {/* Груз */}
          <Card className="p-4">
            <div className="mb-2 text-h2">Груз</div>
            <div className="text-body">{order.contents}</div>
            {order.comment && <div className="mt-2 text-caption text-ink-muted">Комментарий: {order.comment}</div>}
          </Card>

          {/* Таймлайн */}
          <Card className="p-4">
            <div className="mb-3 text-h2">История</div>
            <ol className="relative ml-1.5 space-y-4 border-l-2 border-line pl-5">
              {(order.events ?? []).map((e, i) => {
                const isLast = i === (order.events?.length ?? 0) - 1;
                return (
                  <li key={i} className="relative">
                    <span
                      className={`absolute -left-[27px] top-1 h-3 w-3 rounded-full border-2 border-surface ${
                        isLast ? 'bg-brand' : 'bg-line'
                      }`}
                    />
                    <div className={`text-body ${isLast ? 'font-semibold' : ''}`}>{e.title}</div>
                    <div className="text-caption text-ink-muted">{dateTime(e.created_at)}</div>
                  </li>
                );
              })}
            </ol>
          </Card>
        </div>
      </div>
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-baseline justify-between gap-3">
      <dt className="text-ink-muted">{label}</dt>
      <dd className="tnum">{value}</dd>
    </div>
  );
}
