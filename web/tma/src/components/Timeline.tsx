/** Вертикальный таймлайн событий заказа. */
import type { OrderEvent } from '../lib/types';
import { dateTime } from '../lib/format';

export function Timeline({ events }: { events: OrderEvent[] }) {
  if (events.length === 0) return null;
  const last = events.length - 1;

  return (
    <ol className="relative ml-1.5 space-y-4 border-l-2 border-line pl-5">
      {events.map((e, i) => {
        const isLast = i === last;
        return (
          <li key={`${e.type}-${i}`} className="relative animate-fade-up" style={{ animationDelay: `${i * 40}ms` }}>
            <span
              className={`absolute -left-[27px] top-1 h-3 w-3 rounded-full border-2 border-bg ${
                isLast ? 'animate-pulse-soft bg-brand' : 'bg-line'
              }`}
            />
            <div className={`text-body ${isLast ? 'font-semibold' : ''}`}>{e.title}</div>
            <div className="text-caption text-ink-muted">{dateTime(e.created_at)}</div>
          </li>
        );
      })}
    </ol>
  );
}
