/** Прогресс из 8 сегментов (docs/04). */
import { STATUS_ORDER, statusProgress, type OrderStatus } from '../lib/statuses';

export function Progress({ status, compact = false }: { status: OrderStatus; compact?: boolean }) {
  const total = STATUS_ORDER.length; // 8
  const done = statusProgress(status);
  const isCancelled = status === 'cancelled';

  return (
    <div className={`flex gap-1 ${compact ? 'h-1' : 'h-1.5'}`}>
      {Array.from({ length: total }, (_, i) => {
        const seg = i + 1;
        let cls = 'bg-surface-2';
        if (isCancelled) {
          cls = seg <= done ? 'bg-danger' : 'bg-danger/20';
        } else if (seg < done) {
          cls = 'bg-success';
        } else if (seg === done && done > 0) {
          cls = 'animate-pulse-soft bg-success';
        }
        return <div key={seg} className={`flex-1 rounded-full transition-colors duration-300 ${cls}`} />;
      })}
    </div>
  );
}
