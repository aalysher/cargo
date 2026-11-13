/** Примитивы админки. */
import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { useEffect, useState } from 'react';
import { STATUS_TITLES, type OrderStatus } from '../lib/statuses';

export function StatusBadge({ status }: { status: string }) {
  return (
    <span className="badge" data-st={status}>
      {STATUS_TITLES[status as OrderStatus] ?? status}
    </span>
  );
}

export function Card({ className = '', children }: { className?: string; children: ReactNode }) {
  return <div className={`rounded-lg border border-line bg-surface shadow-1 ${className}`}>{children}</div>;
}

interface BtnProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'danger';
}

export function Button({ variant = 'primary', className = '', ...rest }: BtnProps) {
  const v = variant === 'primary' ? 'btn-primary' : variant === 'danger' ? 'btn-danger' : 'btn-secondary';
  return <button className={`btn ${v} ${className}`} {...rest} />;
}

export function Skeleton({ className = '' }: { className?: string }) {
  return <div className={`animate-pulse-soft rounded-md bg-surface-2 ${className}`} />;
}

export function ErrorBox({ detail, onRetry }: { detail: string; onRetry?: () => void }) {
  return (
    <Card className="flex items-center justify-between gap-3 p-4">
      <div className="text-body text-danger">{detail}</div>
      {onRetry && (
        <Button variant="secondary" onClick={onRetry}>
          Повторить
        </Button>
      )}
    </Card>
  );
}

/** Toast: модульная подписка. */
export interface ToastItem {
  id: number;
  text: string;
  kind: 'ok' | 'err';
}

let listeners: ((t: ToastItem | null) => void)[] = [];
let current: ToastItem | null = null;
let timer: ReturnType<typeof setTimeout> | undefined;

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
    <div className="pointer-events-none fixed bottom-6 left-1/2 z-50 -translate-x-1/2">
      <div
        className={`rounded-md px-4 py-2.5 text-caption shadow-2 ${
          item.kind === 'ok' ? 'bg-ink text-bg' : 'bg-danger text-white'
        }`}
      >
        {item.text}
      </div>
    </div>
  );
}
