/** Toast: подписка через модуль (без контекста). */
import { useEffect, useState } from 'react';

export interface ToastItem {
  id: number;
  text: string;
  kind: 'ok' | 'err';
}

let listeners: ((t: ToastItem | null) => void)[] = [];
let current: ToastItem | null = null;
let timer: ReturnType<typeof setTimeout> | undefined;

/** Показать тост (3 сек). */
export function toast(text: string, kind: 'ok' | 'err' = 'ok') {
  current = { id: Date.now(), text, kind };
  clearTimeout(timer);
  timer = setTimeout(() => {
    current = null;
    listeners.forEach((l) => l(null));
  }, 3000);
  listeners.forEach((l) => l(current));
}

export function ToastHost() {
  const [item, setItem] = useState<ToastItem | null>(current);
  useEffect(() => {
    listeners.push(setItem);
    return () => {
      listeners = listeners.filter((l) => l !== setItem);
    };
  }, []);
  if (!item) return null;
  return (
    <div className="pointer-events-none fixed inset-x-0 bottom-24 z-50 flex justify-center px-4">
      <div
        className={`animate-fade-up rounded-md px-4 py-2.5 text-caption shadow-2 ${
          item.kind === 'ok' ? 'bg-ink text-bg' : 'bg-danger text-white'
        }`}
      >
        {item.text}
      </div>
    </div>
  );
}
