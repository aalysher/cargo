/** Корень TMA: авторизация, стек-роутер, TabBar, MainButton от страниц. */
import { useEffect, useState } from 'react';
import { api } from './lib/api';
import { BOT_USERNAME } from './lib/config';
import { clearToken, devTelegramId, getToken, isOnboarded, setOnboarded, setToken } from './lib/store';
import { initTelegram, isTMA, openBotChat, startParam, webApp } from './lib/telegram';
import type { Profile } from './lib/types';
import { TabBar, type Tab } from './components/TabBar';import { ToastHost } from './components/Toast';
import { Button } from './components/primitives';
import { Onboarding } from './pages/Onboarding';
import { Home } from './pages/Home';
import { Orders } from './pages/Orders';
import { OrderDetail } from './pages/OrderDetail';
import { Calculator } from './pages/Calculator';
import { NewOrderSheet } from './pages/NewOrderSheet';
import { Address } from './pages/Address';
import { ProfilePage } from './pages/Profile';

type Route =
  | { name: 'onboarding' }
  | { name: 'home' }
  | { name: 'orders' }
  | { name: 'order'; id: number }
  | { name: 'calc' }
  | { name: 'profile' }
  | { name: 'address' }
  | { name: 'new'; transport: 'road' | 'air'; weight?: number };

const ROOT_TABS: Route['name'][] = ['home', 'orders', 'calc', 'profile'];

export default function App() {
  const [stack, setStack] = useState<Route[]>([{ name: 'home' }]);
  const [profile, setProfile] = useState<Profile | null>(null);
  const [status, setStatus] = useState<'loading' | 'ready' | 'error'>('loading');
  const [authError, setAuthError] = useState('');

  const current = stack[stack.length - 1];
  const isRoot = ROOT_TABS.includes(current.name);

  // ─── Навигация ───
  const navigate = (r: Route) => setStack((s) => [...s, r]);
  const switchTab = (r: Route) => setStack([r]);
  const back = () => setStack((s) => (s.length > 1 ? s.slice(0, -1) : s));

  // ─── Авторизация при старте ───
  useEffect(() => {
    initTelegram();
    (async () => {
      try {
        if (getToken()) {
          try {
            setProfile(await api<Profile>('/me'));
          } catch {
            clearToken();
          }
        }
        if (!getToken()) {
          const res = isTMA()
            ? await api<{ token: string; profile: Profile }>('/auth/telegram', {
                method: 'POST',
                body: { init_data: webApp()!.initData },
              })
            : await api<{ token: string; profile: Profile }>('/auth/dev', {
                method: 'POST',
                body: { telegram_id: devTelegramId(), first_name: 'Тестовый', username: 'dev_user' },
              });
          setToken(res.token);
          setProfile(res.profile);
        }

        setStatus('ready');
        const m = startParam()?.match(/^order_(\d+)$/);
        if (!isOnboarded()) setStack([{ name: 'onboarding' }]);
        else if (m) setStack([{ name: 'home' }, { name: 'order', id: Number(m[1]) }]);
        else setStack([{ name: 'home' }]);
      } catch (e) {
        setAuthError((e as { detail?: string })?.detail ?? 'Не удалось войти');
        setStatus('error');
      }
    })();
  }, []);

  if (status === 'loading') return <Splash />;
  if (status === 'error' || !profile) return <AuthError detail={authError} />;

  // ─── Страницы ───
  let page: React.ReactNode;
  switch (current.name) {
    case 'onboarding':
      page = <Onboarding onFinish={() => { setOnboarded(); setStack([{ name: 'home' }]); }} />;
      break;
    case 'home':
      page = (
        <Home
          profile={profile}
          onOpenOrder={(id) => navigate({ name: 'order', id })}
          onOpenOrders={() => switchTab({ name: 'orders' })}
          onOpenAddress={() => navigate({ name: 'address' })}
          onOpenCalc={() => switchTab({ name: 'calc' })}
          onNewOrder={() => navigate({ name: 'new', transport: 'road' })}
        />
      );
      break;
    case 'orders':
      page = <Orders onOpenOrder={(id) => navigate({ name: 'order', id })} onNewOrder={() => navigate({ name: 'new', transport: 'road' })} />;
      break;
    case 'order':
      page = <OrderDetail key={current.id} id={current.id} onBack={back} />;
      break;
    case 'calc':
      page = <Calculator onOrder={(transport, weight) => navigate({ name: 'new', transport, weight })} />;
      break;
    case 'profile':
      page = (
        <ProfilePage
          profile={profile}
          onProfileChange={setProfile}
          onOpenAddress={() => navigate({ name: 'address' })}
        />
      );
      break;
    case 'address':
      page = <Address onBack={back} />;
      break;
    case 'new':
      page = (
        <NewOrderSheet
          transport={current.transport}
          weight={current.weight}
          onClose={back}
          onCreated={(order) => setStack((s) => [...s.slice(0, -1), { name: 'order', id: order.id }])}
        />
      );
      break;
  }

  return (
    <div className="mx-auto min-h-dvh max-w-md">
      {page}
      {isRoot && <TabBar active={current.name as Tab} onSelect={(t) => switchTab({ name: t })} />}
      <ToastHost />
    </div>
  );
}

/** Экран загрузки. */
function Splash() {
  return (
    <div className="flex min-h-dvh flex-col items-center justify-center gap-3">
      <div className="animate-pulse-soft text-h1 text-brand">📦 Cargo</div>
      <div className="text-caption text-ink-muted">Китай → Бишкек</div>
    </div>
  );
}

/** Ошибка авторизации (initData не прошёл / сервер недоступен). */
function AuthError({ detail }: { detail: string }) {
  return (
    <div className="flex min-h-dvh flex-col items-center justify-center gap-4 px-6 text-center">
      <div className="text-h1">Нет доступа</div>
      <p className="max-w-[300px] text-body text-ink-muted">
        {detail}. Откройте приложение через бота @{BOT_USERNAME}.
      </p>
      <div className="flex w-full max-w-[280px] flex-col gap-2">
        <Button onClick={() => window.location.reload()}>Повторить</Button>
        <Button variant="secondary" onClick={() => openBotChat(BOT_USERNAME)}>
          Открыть бота
        </Button>
      </div>
    </div>
  );
}
