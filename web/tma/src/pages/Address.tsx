/** Мой адрес склада: карточки по полям + личный код + копирование. */
import { useEffect, useState } from 'react';
import { ArrowLeft, Check, Copy, Warehouse } from 'lucide-react';
import { api } from '../lib/api';
import type { WarehouseAddress } from '../lib/types';
import { copyText } from '../lib/copy';
import { isTMA } from '../lib/telegram';
import { toast } from '../components/Toast';
import { Button, Card, ErrorRetry, Skeleton } from '../components/primitives';

export function Address({ onBack }: { onBack?: () => void }) {
  const [wa, setWa] = useState<WarehouseAddress | null>(null);
  const [error, setError] = useState('');
  const [copied, setCopied] = useState<string | null>(null);

  const load = () => {
    setError('');
    setWa(null);
    api<WarehouseAddress>('/warehouse-address')
      .then(setWa)
      .catch((e) => setError(e.detail));
  };
  useEffect(load, []);

  if (error) {
    return (
      <div className="px-4 pb-28 pt-4">
        <ErrorRetry detail={error} onRetry={load} />
      </div>
    );
  }

  if (!wa) {
    return (
      <div className="space-y-2 px-4 pb-28 pt-4">
        <Skeleton className="h-20" />
        <Skeleton className="h-28" />
        <Skeleton className="h-20" />
      </div>
    );
  }

  const copy = async (key: string, value: string) => {
    if (await copyText(value)) {
      setCopied(key);
      setTimeout(() => setCopied(null), 1000);
    }
  };

  const copyAll = async () => {
    const text = [
      `Получатель: ${wa.contact}`,
      `Адрес склада: ${wa.address}`,
      `Телефон: ${wa.phone}`,
      `Личный код: ${wa.client_code}`,
      'Город: ' + wa.city_cn,
    ].join('\n');
    if (await copyText(text)) toast('Адрес скопирован — отправьте продавцу');
  };

  return (
    <div className="space-y-3 px-4 pb-28 pt-4">
      <h1 className="flex items-center gap-2 text-h1">
        {!isTMA() && onBack && (
          <button onClick={onBack} className="rounded-full p-1.5 text-ink-muted transition-colors active:bg-surface-2" aria-label="Назад">
            <ArrowLeft size={20} />
          </button>
        )}
        <Warehouse size={24} className="text-brand" /> Мой адрес склада
      </h1>

      <p className="text-caption text-ink-muted">
        Укажите эти данные при покупке — продавец отправит посылку на наш склад в Китае.
      </p>

      {/* Личный код — самое важное */}
      <button
        onClick={() => copy('code', wa.client_code)}
        className="w-full rounded-lg border border-brand/30 bg-brand/10 p-4 text-left transition-transform active:scale-[.98]"
      >
        <div className="text-caption text-ink-muted">Ваш личный код (обязательно)</div>
        <div className="mt-0.5 flex items-center justify-between">
          <span className="tnum text-display tracking-[0.2em] text-brand">{wa.client_code}</span>
          {copied === 'code' ? <Check size={18} className="text-success" /> : <Copy size={18} className="text-brand" />}
        </div>
      </button>

      <Field title="Получатель" value={wa.contact} copied={copied === 'contact'} onCopy={() => copy('contact', wa.contact)} />
      <Field title="Адрес склада" value={wa.address} copied={copied === 'address'} onCopy={() => copy('address', wa.address)} />
      <Field title="Телефон" value={wa.phone} copied={copied === 'phone'} onCopy={() => copy('phone', wa.phone)} />
      <Field title="Город" value={wa.city_cn} copied={copied === 'city'} onCopy={() => copy('city', wa.city_cn)} />

      <Button onClick={copyAll}>
        <Copy size={18} className="mr-2" /> Скопировать всё
      </Button>

      <Card className="p-4">
        <h2 className="text-h2">Как заказывать у продавца</h2>
        <ol className="mt-2 list-decimal space-y-1.5 pl-4 text-caption text-ink-muted">
          <li>Укажите получателем наш склад и ваши данные (блок выше).</li>
          <li>Обязательно напишите в адресе ваш личный код — так мы опознаём посылки.</li>
          <li>После покупки добавьте трек-номер в заявку — мы будем ждать посылку.</li>
        </ol>
      </Card>
    </div>
  );
}

function Field({ title, value, copied, onCopy }: { title: string; value: string; copied: boolean; onCopy: () => void }) {
  return (
    <button onClick={onCopy} className="w-full rounded-lg border border-line bg-surface p-3.5 text-left shadow-1 transition-transform active:scale-[.98]">
      <div className="flex items-center justify-between">
        <div>
          <div className="text-caption text-ink-muted">{title}</div>
          <div className="mt-0.5 text-body">{value}</div>
        </div>
        {copied ? <Check size={16} className="shrink-0 text-success" /> : <Copy size={16} className="shrink-0 text-ink-muted" />}
      </div>
    </button>
  );
}
