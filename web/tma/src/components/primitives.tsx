/** Базовые примитивы: Button, Badge, Card, Skeleton, EmptyState, Segments. */
import type { ButtonHTMLAttributes, ReactNode } from 'react';
import type { OrderStatus } from '../lib/statuses';
import { STATUS_TITLES } from '../lib/statuses';

type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
}

const buttonStyles: Record<ButtonVariant, string> = {
  primary: 'bg-brand text-white active:opacity-90',
  secondary: 'bg-surface-2 text-ink active:opacity-80',
  ghost: 'bg-transparent text-brand',
  danger: 'bg-transparent text-danger border border-danger/30',
};

/** Кнопка 48px, радиус 12. */
export function Button({ variant = 'primary', className = '', ...rest }: ButtonProps) {
  return (
    <button
      className={`inline-flex h-12 w-full items-center justify-center rounded-md px-4 text-body-strong transition-transform duration-150 disabled:opacity-50 ${buttonStyles[variant]} ${className}`}
      {...rest}
    />
  );
}

/** Бейдж статуса заказа. */
export function StatusBadge({ status }: { status: OrderStatus }) {
  return (
    <span className="badge" data-st={status}>
      {STATUS_TITLES[status]}
    </span>
  );
}

/** Карточка-контейнер. */
export function Card({ className = '', children }: { className?: string; children: ReactNode }) {
  return <div className={`rounded-lg border border-line bg-surface shadow-1 ${className}`}>{children}</div>;
}

/** Скелетон загрузки. */
export function Skeleton({ className = '' }: { className?: string }) {
  return <div className={`animate-pulse-soft rounded-md bg-surface-2 ${className}`} />;
}

/** Экран ошибки с повтором. */
export function ErrorRetry({ detail, onRetry }: { detail: string; onRetry: () => void }) {
  return (
    <Card className="flex flex-col items-center gap-3 p-6 text-center">
      <div className="text-h2">Не получилось загрузить</div>
      <div className="text-caption text-ink-muted">{detail}</div>
      <Button variant="secondary" onClick={onRetry} className="h-11">
        Повторить
      </Button>
    </Card>
  );
}

/** Пустое состояние. */
export function EmptyState({ icon, title, caption, action }: { icon: ReactNode; title: string; caption?: string; action?: ReactNode }) {
  return (
    <div className="flex flex-col items-center gap-2 py-10 text-center">
      <div className="text-ink-muted [&>svg]:h-12 [&>svg]:w-12">{icon}</div>
      <div className="text-h2">{title}</div>
      {caption && <div className="max-w-[280px] text-caption text-ink-muted">{caption}</div>}
      {action && <div className="mt-2 w-full max-w-[240px]">{action}</div>}
    </div>
  );
}

/** Сегменты-фильтры (scrollable). */
export function Segments<T extends string>({ items, value, onChange }: { items: { key: T; label: string }[]; value: T; onChange: (v: T) => void }) {
  return (
    <div className="flex gap-1.5 overflow-x-auto pb-1 [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
      {items.map((it) => (
        <button
          key={it.key}
          onClick={() => onChange(it.key)}
          className={`h-8 shrink-0 rounded-full px-3 text-caption transition-colors ${
            value === it.key ? 'bg-brand/10 font-semibold text-brand' : 'bg-surface-2 text-ink-muted'
          }`}
        >
          {it.label}
        </button>
      ))}
    </div>
  );
}
