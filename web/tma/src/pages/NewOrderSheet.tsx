/** Новая заявка — bottom sheet с валидацией. */
import { useState } from 'react';
import { Plus, X } from 'lucide-react';
import { useBackButton, useMainButton } from '../hooks/useTelegramUI';
import { api, type ApiError } from '../lib/api';
import type { Order } from '../lib/types';
import { LINE_TITLES } from '../lib/statuses';
import { hapticNotify, isTMA } from '../lib/telegram';
import { toast } from '../components/Toast';
import { Button } from '../components/primitives';

interface Props {
  transport: 'road' | 'air';
  weight?: number;
  onClose: () => void;
  onCreated: (order: Order) => void;
}

export function NewOrderSheet({ transport, weight, onClose, onCreated }: Props) {
  const [contents, setContents] = useState('');
  const [transportState, setTransportState] = useState<'road' | 'air'>(transport);
  const [weightRaw, setWeightRaw] = useState(weight ? String(weight) : '');
  const [tracks, setTracks] = useState<string[]>([]);
  const [trackDraft, setTrackDraft] = useState('');
  const [comment, setComment] = useState('');
  const [busy, setBusy] = useState(false);
  const [formError, setFormError] = useState('');

  useBackButton(onClose, true);

  const weightNum = parseFloat(weightRaw.replace(',', '.'));
  const valid =
    contents.trim().length > 0 &&
    contents.trim().length <= 200 &&
    !Number.isNaN(weightNum) &&
    weightNum > 0 &&
    weightNum <= 1000;

  const submit = async () => {
    if (!valid || busy) return;
    setBusy(true);
    setFormError('');
    try {
      const order = await api<Order>('/orders', {
        method: 'POST',
        body: {
          transport_type: transportState,
          contents: contents.trim(),
          declared_weight_kg: weightNum,
          tracks,
          comment: comment.trim(),
        },
      });
      hapticNotify('success');
      toast(`Заявка ${order.order_number} отправлена`);
      onCreated(order);
    } catch (e) {
      hapticNotify('error');
      setFormError((e as ApiError).detail ?? 'Не удалось отправить заявку');
      setBusy(false);
    }
  };

  useMainButton('Отправить заявку', submit, { enabled: valid, loading: busy });

  const addTrack = () => {
    const t = trackDraft.trim();
    if (!t || tracks.length >= 10) return;
    setTracks((ts) => [...ts, t]);
    setTrackDraft('');
  };

  return (
    <div className="fixed inset-0 z-50 flex items-end bg-overlay" onClick={onClose}>
      <div
        className="max-h-[92dvh] w-full overflow-y-auto rounded-t-xl bg-bg pb-[calc(16px+env(safe-area-inset-bottom))] shadow-3 animate-fade-up"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="sticky top-0 z-10 rounded-t-xl border-b border-line bg-bg px-4 pb-2 pt-2">
          <div className="mx-auto mb-2 h-1 w-10 rounded-full bg-line" />
          <div className="flex items-center justify-between">
            <h1 className="text-h1">Новая заявка</h1>
            <button onClick={onClose} className="rounded-full p-2 text-ink-muted">
              <X size={20} />
            </button>
          </div>
        </div>

        <div className="space-y-3 px-4 pt-3">
          {/* Линия */}
          <div>
            <div className="mb-1 text-caption text-ink-muted">Линия доставки</div>
            <div className="grid grid-cols-2 gap-2">
              {(['road', 'air'] as const).map((t) => (
                <button
                  key={t}
                  onClick={() => setTransportState(t)}
                  className={`h-11 rounded-md border text-body transition-colors ${
                    transportState === t ? 'border-brand bg-brand/10 text-brand' : 'border-line bg-surface'
                  }`}
                >
                  {t === 'road' ? 'Авто' : 'Авиа'}
                </button>
              ))}
            </div>
          </div>

          {/* Груз */}
          <div>
            <label className="text-caption text-ink-muted">Что везём</label>
            <input
              className="field mt-1"
              placeholder="Например: одежда, обувь"
              value={contents}
              onChange={(e) => setContents(e.target.value)}
              maxLength={200}
            />
          </div>

          {/* Вес */}
          <div>
            <label className="text-caption text-ink-muted">Вес, кг</label>
            <input
              className="field tnum mt-1"
              inputMode="decimal"
              placeholder="4.5"
              value={weightRaw}
              onChange={(e) => setWeightRaw(e.target.value.replace(/[^\d.,]/g, ''))}
            />
          </div>

          {/* Треки */}
          <div>
            <label className="text-caption text-ink-muted">
              Трек-номера покупок (необязательно, до 10)
            </label>
            <div className="mt-1 flex gap-2">
              <input
                className="field"
                placeholder="7812345678"
                value={trackDraft}
                onChange={(e) => setTrackDraft(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && addTrack()}
                maxLength={64}
              />
              <button
                onClick={addTrack}
                disabled={!trackDraft.trim() || tracks.length >= 10}
                className="h-11 w-11 shrink-0 rounded-md bg-surface-2 text-ink disabled:opacity-40"
              >
                <Plus size={18} className="mx-auto" />
              </button>
            </div>
            {tracks.length > 0 && (
              <div className="mt-2 flex flex-wrap gap-1.5">
                {tracks.map((t, i) => (
                  <button
                    key={`${t}-${i}`}
                    onClick={() => setTracks((ts) => ts.filter((_, j) => j !== i))}
                    className="flex items-center gap-1 rounded-full bg-surface-2 px-2.5 py-1 text-caption"
                  >
                    {t} <X size={12} className="text-ink-muted" />
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Комментарий */}
          <div>
            <label className="text-caption text-ink-muted">Комментарий (необязательно)</label>
            <textarea
              className="field mt-1 h-auto min-h-[72px] py-2"
              placeholder="Пожелания, особые отметки"
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              maxLength={500}
            />
          </div>

          {formError && <div className="rounded-md bg-danger/10 px-3 py-2 text-caption text-danger">{formError}</div>}

          {!isTMA() && (
            <Button onClick={submit} disabled={!valid || busy}>
              {busy ? 'Отправляем…' : 'Отправить заявку'}
            </Button>
          )}

          <p className="text-center text-micro text-ink-muted">
            {LINE_TITLES[transportState]} · оператор проверит заявку
          </p>
        </div>
      </div>
    </div>
  );
}
