/** Тарифы: просмотр всем, правка — только admin. */
import { useEffect, useState } from 'react';
import { api, type ApiError } from '../lib/api';
import type { Tariff } from '../lib/types';
import { LINE_TITLES } from '../lib/statuses';
import { toast } from '../components/primitives';
import { Button, Card, ErrorBox, Skeleton } from '../components/primitives';
import { getProfile } from '../lib/store';

export function Tariffs() {
  const isAdmin = getProfile()?.role === 'admin';
  const [tariffs, setTariffs] = useState<Tariff[] | null>(null);
  const [error, setError] = useState('');
  const [drafts, setDrafts] = useState<Record<number, { price: string; min: string; active: boolean }>>({});
  const [busyId, setBusyId] = useState<number | null>(null);

  const load = () => {
    setError('');
    setTariffs(null);
    api<Tariff[]>('/admin/tariffs')
      .then((list) => {
        setTariffs(list);
        setDrafts(
          Object.fromEntries(
            list.map((t) => [t.id, { price: String(t.price_per_kg), min: String(t.min_chargeable_kg), active: t.is_active }]),
          ),
        );
      })
      .catch((e) => setError(e.detail));
  };
  useEffect(load, []);

  const save = async (t: Tariff) => {
    const d = drafts[t.id];
    const price = parseFloat(d.price.replace(',', '.'));
    const min = parseFloat(d.min.replace(',', '.'));
    if (!price || price <= 0 || !min || min <= 0) {
      toast('Проверьте значения: числа > 0', 'err');
      return;
    }
    setBusyId(t.id);
    try {
      const updated = await api<Tariff>(`/admin/tariffs/${t.id}`, {
        method: 'PATCH',
        body: { price_per_kg: price, min_chargeable_kg: min, is_active: d.active },
      });
      setTariffs((list) => (list ?? []).map((x) => (x.id === t.id ? updated : x)));
      toast('Тариф сохранён');
    } catch (e) {
      toast((e as ApiError).detail, 'err');
    } finally {
      setBusyId(null);
    }
  };

  if (error) return <ErrorBox detail={error} onRetry={load} />;
  if (!tariffs) return <Skeleton className="h-64" />;

  return (
    <div className="space-y-3">
      {!isAdmin && (
        <Card className="p-3 text-caption text-ink-muted">Просмотр — правка тарифов доступна роли admin</Card>
      )}

      <Card className="overflow-hidden">
        <table className="w-full text-body">
          <thead>
            <tr className="border-b border-line text-left text-caption text-ink-muted">
              <th className="px-4 py-2.5 font-medium">Линия</th>
              <th className="px-4 py-2.5 font-medium">Ставка за кг</th>
              <th className="px-4 py-2.5 font-medium">Мин. вес, кг</th>
              <th className="px-4 py-2.5 font-medium">Активен</th>
              {isAdmin && <th className="px-4 py-2.5" />}
            </tr>
          </thead>
          <tbody>
            {tariffs.map((t) => {
              const d = drafts[t.id] ?? { price: String(t.price_per_kg), min: String(t.min_chargeable_kg), active: t.is_active };
              return (
                <tr key={t.id} className="border-b border-line last:border-0">
                  <td className="px-4 py-3 font-semibold">{LINE_TITLES[t.transport_type]}</td>
                  <td className="px-4 py-3">
                    <input
                      className="field tnum max-w-[120px]"
                      value={d.price}
                      onChange={(e) => setDrafts((s) => ({ ...s, [t.id]: { ...d, price: e.target.value.replace(/[^\d.,]/g, '') } }))}
                      disabled={!isAdmin || busyId === t.id}
                    />
                  </td>
                  <td className="px-4 py-3">
                    <input
                      className="field tnum max-w-[100px]"
                      value={d.min}
                      onChange={(e) => setDrafts((s) => ({ ...s, [t.id]: { ...d, min: e.target.value.replace(/[^\d.,]/g, '') } }))}
                      disabled={!isAdmin || busyId === t.id}
                    />
                  </td>
                  <td className="px-4 py-3">
                    <input
                      type="checkbox"
                      checked={d.active}
                      onChange={(e) => setDrafts((s) => ({ ...s, [t.id]: { ...d, active: e.target.checked } }))}
                      disabled={!isAdmin || busyId === t.id}
                      className="h-4 w-4"
                    />
                  </td>
                  {isAdmin && (
                    <td className="px-4 py-3 text-right">
                      <Button onClick={() => save(t)} disabled={busyId === t.id} className="h-9 px-3 text-caption">
                        Сохранить
                      </Button>
                    </td>
                  )}
                </tr>
              );
            })}
          </tbody>
        </table>
      </Card>
      <p className="text-caption text-ink-muted">
        Ставка применяется к новым расчётам калькулятора и при взвешивании заказов.
      </p>
    </div>
  );
}
