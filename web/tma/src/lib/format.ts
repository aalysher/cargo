/** Форматирование денег/веса/дат (ru-RU). */

export function money(v: number | null | undefined, currency = 'USD'): string {
  if (v == null) return '—';
  return new Intl.NumberFormat('ru-RU', { style: 'currency', currency, minimumFractionDigits: 2 }).format(v);
}

export function kg(v: number | null | undefined): string {
  if (v == null) return '—';
  return `${new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 2 }).format(v)} кг`;
}

export function dateShort(iso: string): string {
  return new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'short' }).format(new Date(iso));
}

export function dateTime(iso: string): string {
  return new Intl.DateTimeFormat('ru-RU', {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(iso));
}
