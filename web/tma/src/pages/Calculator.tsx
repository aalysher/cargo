/** Калькулятор: линия, вес, расшифровка, итог. */
import { useEffect, useState } from 'react';
import { Calculator as CalcIcon } from 'lucide-react';
import { api } from '../lib/api';
import type { Estimate, Tariff } from '../lib/types';
import { LINE_TITLES } from '../lib/statuses';
import { hapticSelection } from '../lib/telegram';
import { money } from '../lib/format';
import { Card, ErrorRetry, Skeleton } from '../components/primitives';

interface Props {
  onOrder: (transport: 'road' | 'air', weight: number) => void;
}

const LIMITS = { min: 0.1, max: 1000 };

export function Calculator({ onOrder }: Props) {
  const [tariffs, setTariffs] = useState<Tariff[] | null>(null);
  const [error, setError] = useState('');
  const [transport, setTransport] = useState<'road' | 'air'>('road');
  const [weightRaw, setWeightRaw] = useState('');
  const [estimate, setEstimate] = useState<Estimate | null>(null);
  const [estimateError, setEstimateError] = useState('');

  const load = () => {
    setError('');
    setTariffs(null);
    api<Tariff[]>('/tariffs')
      .then(setTariffs)
      .catch((e) => setError(e.detail));
  };
  useEffect(load, []);

  // Пересчёт с debounce 300 мс — формула только на бэке.
  const weight = parseFloat(weightRaw.replace(',', '.'));
  useEffect(() => {
    setEstimate(null);
    setEstimateError('');
    if (!weightRaw || Number.isNaN(weight) || weight <= 0) return;
    const t = setTimeout(() => {
      api<Estimate>(`/tariffs/estimate?transport=${transport}&weight=${weight}`)
        .then((d) => {
          if (d.transport_type !== transport) return; // ответ устаревшей линии
          setEstimate(d);
        })
        .catch((e) => setEstimateError(e.detail));
    }, 300);
    return () => clearTimeout(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [transport, weightRaw]);

  const canOrder = estimate !== null && estimate.transport_type === transport;

  const pickLine = (t: 'road' | 'air') => {
    hapticSelection();
    setTransport(t);
  };

  return (
    <div className="space-y-3 px-4 pb-28 pt-4">
      <h1 className="flex items-center gap-2 text-h1">
        <CalcIcon size={24} className="text-brand" /> Калькулятор
      </h1>

      {error ? (
        <ErrorRetry detail={error} onRetry={load} />
      ) : tariffs === null ? (
        <Skeleton className="h-64" />
      ) : (
        <>
          {/* Линия */}
          <div className="grid grid-cols-2 gap-2">
            {(['road', 'air'] as const).map((t) => {
              const tf = tariffs.find((x) => x.transport_type === t);
              return (
                <button
                  key={t}
                  onClick={() => pickLine(t)}
                  className={`rounded-lg border p-3 text-left transition-colors ${
                    transport === t ? 'border-brand bg-brand/10' : 'border-line bg-surface'
                  }`}
                >
                  <div className="text-body-strong">{t === 'road' ? 'Авто' : 'Авиа'}</div>
                  <div className="tnum text-caption text-ink-muted">
                    {tf ? `${money(tf.price_per_kg, tf.currency)} / кг` : '—'}
                  </div>
                </button>
              );
            })}
          </div>

          {/* Вес */}
          <Card className="p-4">
            <label className="text-caption text-ink-muted">Вес посылки, кг</label>
            <input
              inputMode="decimal"
              placeholder="Например, 4.5"
              value={weightRaw}
              onChange={(e) => setWeightRaw(e.target.value.replace(/[^\d.,]/g, ''))}
              className="field tnum mt-1.5 text-h1"
            />
            <div className="mt-1 text-micro text-ink-muted">
              от {LIMITS.min} до {LIMITS.max} кг
            </div>
          </Card>

          {/* Расшифровка */}
          {estimate && !estimateError && (
            <Card className="animate-fade-up p-4">
              <div className="space-y-1.5 text-body">
                <Row label="Ставка" value={`${money(estimate.price_per_kg, estimate.currency)} / кг`} />
                <Row label="Расчётный вес" value={`${estimate.chargeable_kg} кг`} />
                {estimate.chargeable_kg > estimate.declared_kg && (
                  <div className="text-micro text-warning">
                    Минимальный расчётный вес — {estimate.chargeable_kg} кг
                  </div>
                )}
                <div className="my-2 h-px bg-line" />
                <div className="flex items-baseline justify-between">
                  <span className="text-body-strong">Итого</span>
                  <span className="tnum text-display text-brand">{money(estimate.estimated_cost, estimate.currency)}</span>
                </div>
              </div>
            </Card>
          )}
          {estimateError && <Card className="p-4 text-body text-danger">{estimateError}</Card>}

          <button
            disabled={!canOrder}
            onClick={() => canOrder && onOrder(transport, estimate!.declared_kg)}
            className="mt-2 h-12 w-full rounded-md bg-brand text-body-strong text-white transition-transform active:scale-[.98] disabled:opacity-50"
          >
            Оформить заявку с этими данными
          </button>
          <p className="text-center text-micro text-ink-muted">Линия: {LINE_TITLES[transport]}</p>
        </>
      )}
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-baseline justify-between">
      <span className="text-ink-muted">{label}</span>
      <span className="tnum">{value}</span>
    </div>
  );
}
