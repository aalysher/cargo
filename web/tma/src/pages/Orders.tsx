/** Заказы: сегменты-фильтры + список карточек. */
import { useEffect, useState } from 'react';
import { PackageOpen } from 'lucide-react';
import { api } from '../lib/api';
import type { Order } from '../lib/types';
import { SCOPES } from '../lib/statuses';
import { Button, EmptyState, ErrorRetry, Segments, Skeleton } from '../components/primitives';
import { OrderCard } from '../components/OrderCard';

type Scope = (typeof SCOPES)[number]['key'];

export function Orders({ onOpenOrder, onNewOrder }: { onOpenOrder: (id: number) => void; onNewOrder: () => void }) {
  const [scope, setScope] = useState<Scope>('active');
  const [orders, setOrders] = useState<Order[] | null>(null);
  const [error, setError] = useState('');

  const load = (s: Scope) => {
    setError('');
    setOrders(null);
    api<{ items: Order[]; next_cursor: number }>(`/orders?scope=${s}`)
      .then((d) => setOrders(d.items))
      .catch((e) => setError(e.detail));
  };

  useEffect(() => {
    load(scope);
  }, [scope]);

  return (
    <div className="space-y-3 px-4 pb-28 pt-4">
      <h1 className="text-h1">Заказы</h1>

      <Segments items={[...SCOPES]} value={scope} onChange={setScope} />

      {error ? (
        <ErrorRetry detail={error} onRetry={() => load(scope)} />
      ) : orders === null ? (
        <div className="space-y-2">
          <Skeleton className="h-[104px]" />
          <Skeleton className="h-[104px]" />
          <Skeleton className="h-[104px]" />
        </div>
      ) : orders.length === 0 ? (
        <EmptyState
          icon={<PackageOpen />}
          title="Пока нет заказов"
          caption="Создайте заявку — или сначала посчитайте доставку"
          action={<Button onClick={onNewOrder}>Новая заявка</Button>}
        />
      ) : (
        <div className="space-y-2">
          {orders.map((o) => (
            <OrderCard key={o.id} order={o} onClick={() => onOpenOrder(o.id)} />
          ))}
        </div>
      )}
    </div>
  );
}
