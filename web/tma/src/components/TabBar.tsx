/** Нижнее меню: 🏠 Главная · 📦 Заказы · 🧮 Калькулятор · 👤 Профиль. */
import { Calculator, Home, Package, User } from 'lucide-react';
import { hapticSelection } from '../lib/telegram';

export type Tab = 'home' | 'orders' | 'calc' | 'profile';

const TABS: { key: Tab; label: string; icon: typeof Home }[] = [
  { key: 'home', label: 'Главная', icon: Home },
  { key: 'orders', label: 'Заказы', icon: Package },
  { key: 'calc', label: 'Калькулятор', icon: Calculator },
  { key: 'profile', label: 'Профиль', icon: User },
];

export function TabBar({ active, onSelect }: { active: Tab; onSelect: (t: Tab) => void }) {
  return (
    <nav className="fixed inset-x-0 bottom-0 z-40 border-t border-line bg-surface pb-[calc(8px+env(safe-area-inset-bottom))] pt-1.5">
      <div className="mx-auto grid max-w-md grid-cols-4">
        {TABS.map(({ key, label, icon: Icon }) => {
          const isActive = key === active;
          return (
            <button
              key={key}
              onClick={() => {
                if (!isActive) hapticSelection();
                onSelect(key);
              }}
              className={`flex h-14 flex-col items-center justify-center gap-0.5 ${isActive ? 'text-brand' : 'text-ink-muted'}`}
            >
              <Icon size={24} strokeWidth={1.75} />
              <span className={`text-micro ${isActive ? '' : 'font-normal'}`}>{label}</span>
            </button>
          );
        })}
      </div>
    </nav>
  );
}
