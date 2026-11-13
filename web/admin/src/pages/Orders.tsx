/** Заказы: таблица + фильтры + поиск + курсорная пагинация. */
import { useEffect, useRef, useState } from 'react';
import { RefreshCw } from 'lucide-react';
import { api, type ApiError } from '../lib/api';
import type { AdminOrder, AdminOrderList } from '../lib/types';
import { ALL_STATUSES, LINE_TITLES, STATUS_TITLES, type OrderStatus } from '../lib/statuses';
import { dateTime, kg, money } from '../lib/format';
import { Button, Card, ErrorBox, Skeleton, StatusBadge } from '../components/primitives';

export function Orders({ onOpen }: { onOpen: (id: number) => void }) {
  const [status, setStatus] = useState('');
  const [search, setSearch] = useState('');
  const [query, setQuery] = useState('');
  const [orders, setOrders] = useState<AdminOrder[] | null>(null);
  const [nextCursor, setNextCursor] = useState(0);
  const [error, setError] = useState('');
  const [loadingMore, setLoadingMore] = useState(false);
  const searchTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  // Дебаунс поиска
  useEffect(() => {
    clearTimeout(searchTimer.current);
    searchTimer.current = setTimeout(() => setQuery(search.trim()), 400);
    return () => clearTimeout(searchTimer.current);
  }, [search]);

  const load = () => {
    setError('');
    setOrders(null);
    setNextCursor(0);
    const params = new URLSearchParams();
    if (status) params.set('status', status);
    if (query) params.set('q', query);
    api<AdminOrderList>(`/admin/orders?${params}`)
      .then((d) => {
        setOrders(d.items);
        setNextCursor(d.next_cursor);
      })
      .catch((e) => setError(e.detail));
  };

  useEffect(load, [status, query]);

  const loadMore = async () => {
    if (!nextCursor || loadingMore) return;
    setLoadingMore(true);
    try {
      const params = new URLSearchParams();
      if (status) params.set('status', status);
      if (query) params.set('q', query);
      params.set('cursor', String(nextCursor));
      const d = await api<AdminOrderList>(`/admin/orders?${params}`);
      setOrders((prev) => [...(prev ?? []), ...d.items]);
      setNextCursor(d.next_cursor);
    } catch (e) {
      setError((e as ApiError).detail);
    } finally {
      setLoadingMore(false);
    }
  };

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-center gap-2">
        <input
          className="field max-w-[280px]"
          placeholder="Поиск: номер, код, телефон, груз…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <select className="field max-w-[220px]" value={status} onChange={(e) => setStatus(e.target.value)}>
          <option value="">Все статусы</option>
          {ALL_STATUSES.map((s) => (
            <option key={s} value={s}>
              {STATUS_TITLES[s as OrderStatus]}
            </option>
          ))}
        </select>
        <Button variant="secondary" onClick={load} title="Обновить">
          <RefreshCw size={16} />
        </Button>
      </div>

      {error ? (
        <ErrorBox detail={error} onRetry={load} />
      ) : orders === null ? (
        <Skeleton className="h-72" />
      ) : (
        <Card className="overflow-hidden">
          <table className="w-full text-body">
            <thead>
              <tr className="border-b border-line text-left text-caption text-ink-muted">
                <th className="px-4 py-2.5 font-medium">Номер</th>
                <th className="px-4 py-2.5 font-medium">Статус</th>
                <th className="px-4 py-2.5 font-medium">Клиент</th>
                <th className="px-4 py-2.5 font-medium">Линия</th>
                <th className="px-4 py-2.5 font-medium text-right">Вес</th>
                <th className="px-4 py-2.5 font-medium text-right">Стоимость</th>
                <th className="px-4 py-2.5 font-medium">Обновлено</th>
              </tr>
            </thead>
            <tbody>
              {orders.map((o) => (
                <tr
                  key={o.id}
                  onClick={() => onOpen(o.id)}
                  className="cursor-pointer border-b border-line last:border-0 transition-colors hover:bg-surface-2"
                >
                  <td className="tnum px-4 py-2.5 font-semibold">{o.order_number}</td>
                  <td className="px-4 py-2.5">
                    <StatusBadge status={o.status} />
                  </td>
                  <td className="px-4 py-2.5">
                    <div>{[o.client.first_name, o.client.last_name].filter(Boolean).join(' ') || '—'}</div>
                    <div className="text-caption text-ink-muted">
                      {o.client.phone || o.client.code}
                    </div>
                  </td>
                  <td className="px-4 py-2.5 text-ink-muted">{LINE_TITLES[o.transport_type]}</td>
                  <td className="tnum px-4 py-2.5 text-right">
                    {kg(o.actual_weight_kg ?? o.declared_weight_kg)}
                  </td>
                  <td className="tnum px-4 py-2.5 text-right">
                    {money(o.final_cost ?? o.estimated_cost, o.currency)}
                  </td>
                  <td className="px-4 py-2.5 text-caption text-ink-muted">{dateTime(o.updated_at)}</td>
                </tr>
              ))}
              {orders.length === 0 && (
                <tr>
                  <td colSpan={7} className="px-4 py-10 text-center text-ink-muted">
                    Ничего не найдено
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </Card>
      )}

      {nextCursor > 0 && orders !== null && (
        <div className="flex justify-center">
          <Button variant="secondary" onClick={loadMore} disabled={loadingMore}>
            {loadingMore ? 'Загружаем…' : 'Показать ещё'}
          </Button>
        </div>
      )}
    </div>
  );
}
