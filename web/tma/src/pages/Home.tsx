/** Главная: приветствие, активные заказы, плитки «Адрес склада» / «Посчитать доставку». */
import { useEffect, useState } from 'react';
import { ArrowUpRight, MapPin, Package } from 'lucide-react';
import { api } from '../lib/api';
import type { Order, Profile } from '../lib/types';
import { isTMA } from '../lib/telegram';
import { Card, Button, ErrorRetry, Skeleton } from '../components/primitives';
import { OrderCard } from '../components/OrderCard';

interface Props {
  profile: Profile;
  onOpenOrder: (id: number) => void;
  onOpenOrders: () => void;
  onOpenAddress: () => void;
  onOpenCalc: () => void;
  onNewOrder: () => void;
}

export function Home({ profile, onOpenOrder, onOpenOrders, onOpenAddress, onOpenCalc, onNewOrder }: Props) {
  const [orders, setOrders] = useState<Order[] | null>(null);
  const [error, setError] = useState('');

  const load = () => {
    setError('');
    setOrders(null);
    api<{ items: Order[] }>('/orders?scope=delivering')
      .then((d) => setOrders(d.items))
      .catch((e) => setError(e.detail));
  };
  useEffect(load, []);

  return (
    <div className="space-y-3 px-4 pb-28 pt-4">
      <h1 className="text-h1">Привет, {profile.first_name || 'друг'} 👋</h1>

      {!isTMA() && (
        <Button onClick={onNewOrder}>+ Новая заявка</Button>
      )}

      {/* Активные заказы */}
      <div className="flex items-baseline justify-between">
        <h2 className="text-h2">Посылки в пути</h2>
        <button onClick={onOpenOrders} className="flex items-center gap-0.5 text-caption text-brand">
          Все заказы <ArrowUpRight size={13} />
        </button>
      </div>

      {error ? (
        <ErrorRetry detail={error} onRetry={load} />
      ) : orders === null ? (
        <div className="space-y-2">
          <Skeleton className="h-[104px]" />
          <Skeleton className="h-[104px]" />
        </div>
      ) : orders.length === 0 ? (
        <Card className="p-4">
          <div className="text-body text-ink-muted">Здесь появятся ваши посылки</div>
        </Card>
      ) : (
        <div className="space-y-2">
          {orders.slice(0, 3).map((o) => (
            <OrderCard key={o.id} order={o} onClick={() => onOpenOrder(o.id)} />
          ))}
        </div>
      )}

      {/* Плитки */}
      <div className="grid grid-cols-2 gap-3 pt-1">
        <Tile icon={<MapPin size={22} strokeWidth={1.75} />} title="Мой адрес склада" subtitle="Копировать в 1 тап" onClick={onOpenAddress} />
        <Tile icon={<Package size={22} strokeWidth={1.75} />} title="Посчитать доставку" subtitle="Ставки из тарифов" onClick={onOpenCalc} />
      </div>
    </div>
  );
}

function Tile({ icon, title, subtitle, onClick }: { icon: React.ReactNode; title: string; subtitle: string; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className="rounded-lg border border-line bg-surface p-4 text-left shadow-1 transition-transform active:scale-[.98]"
    >
      <div className="text-brand">{icon}</div>
      <div className="mt-2 text-body-strong">{title}</div>
      <div className="text-caption text-ink-muted">{subtitle}</div>
    </button>
  );
}
