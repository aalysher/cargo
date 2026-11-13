/** Настройки: адрес склада (правка — только admin). */
import { useEffect, useState } from 'react';
import { api, type ApiError } from '../lib/api';
import { toast } from '../components/primitives';
import { Button, Card, ErrorBox, Skeleton } from '../components/primitives';
import { getProfile } from '../lib/store';

interface WarehouseAddress {
  contact: string;
  address: string;
  phone: string;
  city_cn: string;
}

export function Settings() {
  const isAdmin = getProfile()?.role === 'admin';
  const [wa, setWa] = useState<WarehouseAddress | null>(null);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  const load = () => {
    setError('');
    api<{ settings: Record<string, WarehouseAddress> }>('/admin/settings')
      .then((d) => setWa(d.settings.warehouse_address ?? { contact: '', address: '', phone: '', city_cn: '' }))
      .catch((e) => setError(e.detail));
  };
  useEffect(load, []);

  const save = async () => {
    if (!wa || busy) return;
    setBusy(true);
    try {
      await api('/admin/settings/warehouse_address', { method: 'PUT', body: { value: wa } });
      toast('Настройка сохранена');
    } catch (e) {
      toast((e as ApiError).detail, 'err');
    } finally {
      setBusy(false);
    }
  };

  if (error) return <ErrorBox detail={error} onRetry={load} />;
  if (!wa) return <Skeleton className="h-72" />;

  const set = (k: keyof WarehouseAddress, v: string) => setWa({ ...wa, [k]: v });

  return (
    <div className="max-w-xl space-y-3">
      {!isAdmin && <Card className="p-3 text-caption text-ink-muted">Просмотр — правка доступна роли admin</Card>}

      <Card className="space-y-3 p-4">
        <div className="text-h2">Адрес склада в Китае</div>
        <p className="text-caption text-ink-muted">
          Показывается клиентам в TMA (блок «Мой адрес склада»). Личный код клиенту подставляется автоматически.
        </p>
        <div>
          <label className="text-caption text-ink-muted">Получатель / контакт</label>
          <input className="field mt-1" value={wa.contact} onChange={(e) => set('contact', e.target.value)} disabled={!isAdmin} />
        </div>
        <div>
          <label className="text-caption text-ink-muted">Адрес (можно по-китайски)</label>
          <textarea
            className="field mt-1 h-auto min-h-[72px] py-2"
            value={wa.address}
            onChange={(e) => set('address', e.target.value)}
            disabled={!isAdmin}
          />
        </div>
        <div className="grid grid-cols-2 gap-2">
          <div>
            <label className="text-caption text-ink-muted">Телефон</label>
            <input className="field mt-1" value={wa.phone} onChange={(e) => set('phone', e.target.value)} disabled={!isAdmin} />
          </div>
          <div>
            <label className="text-caption text-ink-muted">Город</label>
            <input className="field mt-1" value={wa.city_cn} onChange={(e) => set('city_cn', e.target.value)} disabled={!isAdmin} />
          </div>
        </div>
        {isAdmin && (
          <Button onClick={save} disabled={busy}>
            {busy ? 'Сохраняем…' : 'Сохранить'}
          </Button>
        )}
      </Card>
    </div>
  );
}
