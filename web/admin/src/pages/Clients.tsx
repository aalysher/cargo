/** Клиенты: поиск + карточка с последними заказами. */
import { useEffect, useRef, useState } from 'react';
import { api, type ApiError } from '../lib/api';
import type { AdminClientFull } from '../lib/types';
import { dateShort, kg, money } from '../lib/format';
import { Card, ErrorBox, Skeleton, StatusBadge } from '../components/primitives';

export function Clients({ initialQuery = '' }: { initialQuery?: string }) {
  const [search, setSearch] = useState(initialQuery);
  const [query, setQuery] = useState(initialQuery);
  const [clients, setClients] = useState<AdminClientFull[] | null>(null);
  const [selected, setSelected] = useState<AdminClientFull | null>(null);
  const [error, setError] = useState('');
  const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  useEffect(() => {
    clearTimeout(timer.current);
    timer.current = setTimeout(() => setQuery(search.trim()), 400);
    return () => clearTimeout(timer.current);
  }, [search]);

  const load = () => {
    setError('');
    setClients(null);
    api<AdminClientFull[]>(`/admin/clients?q=${encodeURIComponent(query)}`)
      .then(setClients)
      .catch((e) => setError(e.detail));
  };
  useEffect(load, [query]);

  const open = async (id: number) => {
    setSelected(null);
    try {
      setSelected(await api<AdminClientFull>(`/admin/clients/${id}`));
    } catch (e) {
      setError((e as ApiError).detail);
    }
  };

  return (
    <div className="space-y-3">
      <input
        className="field max-w-[320px]"
        placeholder="Поиск: телефон, код, @username, имя…"
        value={search}
        onChange={(e) => setSearch(e.target.value)}
      />

      {error ? (
        <ErrorBox detail={error} onRetry={load} />
      ) : clients === null ? (
        <Skeleton className="h-56" />
      ) : (
        <Card className="overflow-hidden">
          <table className="w-full text-body">
            <thead>
              <tr className="border-b border-line text-left text-caption text-ink-muted">
                <th className="px-4 py-2.5 font-medium">Клиент</th>
                <th className="px-4 py-2.5 font-medium">Telegram</th>
                <th className="px-4 py-2.5 font-medium">Телефон</th>
                <th className="px-4 py-2.5 font-medium">Код</th>
                <th className="px-4 py-2.5" />
              </tr>
            </thead>
            <tbody>
              {clients.map((c) => (
                <tr key={c.id} className="border-b border-line last:border-0">
                  <td className="px-4 py-2.5">
                    {[c.first_name, c.last_name].filter(Boolean).join(' ') || 'Без имени'}
                  </td>
                  <td className="px-4 py-2.5 text-ink-muted">{c.username ? `@${c.username}` : '—'}</td>
                  <td className="tnum px-4 py-2.5">{c.phone || '—'}</td>
                  <td className="tnum px-4 py-2.5 font-semibold">{c.code}</td>
                  <td className="px-4 py-2.5 text-right">
                    <button
                      onClick={() => (selected?.id === c.id ? setSelected(null) : open(c.id))}
                      className="text-caption text-brand"
                    >
                      {selected?.id === c.id ? 'Скрыть' : 'История'}
                    </button>
                  </td>
                </tr>
              ))}
              {clients.length === 0 && (
                <tr>
                  <td colSpan={5} className="px-4 py-10 text-center text-ink-muted">
                    Клиенты не найдены
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </Card>
      )}

      {selected && (
        <Card className="p-4">
          <div className="mb-3 flex items-baseline justify-between">
            <div className="text-h2">
              {[selected.first_name, selected.last_name].filter(Boolean).join(' ') || 'Без имени'} — заказы
            </div>
            <div className="text-caption text-ink-muted">
              telegram_id {selected.telegram_id} · всего {selected.orders_count}
            </div>
          </div>
          {selected.orders && selected.orders.length > 0 ? (
            <table className="w-full text-body">
              <thead>
                <tr className="border-b border-line text-left text-caption text-ink-muted">
                  <th className="py-2 pr-3 font-medium">Номер</th>
                  <th className="py-2 pr-3 font-medium">Статус</th>
                  <th className="py-2 pr-3 font-medium text-right">Вес</th>
                  <th className="py-2 pr-3 font-medium text-right">Стоимость</th>
                  <th className="py-2 font-medium">Дата</th>
                </tr>
              </thead>
              <tbody>
                {selected.orders.map((o) => (
                  <tr key={o.id} className="border-b border-line last:border-0">
                    <td className="tnum py-2 pr-3 font-semibold">{o.order_number}</td>
                    <td className="py-2 pr-3">
                      <StatusBadge status={o.status} />
                    </td>
                    <td className="tnum py-2 pr-3 text-right">{kg(o.declared_weight_kg)}</td>
                    <td className="tnum py-2 pr-3 text-right">{money(o.final_cost ?? o.estimated_cost, o.currency)}</td>
                    <td className="py-2 text-caption text-ink-muted">{dateShort(o.created_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            <div className="text-caption text-ink-muted">Заказов пока нет</div>
          )}
        </Card>
      )}
    </div>
  );
}
