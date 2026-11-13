/** Корень админки: логин → layout с секциями. */
import { useState } from 'react';
import { Package, Settings as SettingsIcon, Users } from 'lucide-react';
import { getToken, getProfile, clearAuth, type AdminProfile } from './lib/store';
import { Login } from './pages/Login';
import { Orders } from './pages/Orders';
import { OrderDetail } from './pages/OrderDetail';
import { Tariffs } from './pages/Tariffs';
import { Clients } from './pages/Clients';
import { Settings } from './pages/Settings';
import { ToastHost } from './components/primitives';

type Section =
  | { name: 'orders' }
  | { name: 'order'; id: number }
  | { name: 'tariffs' }
  | { name: 'clients' }
  | { name: 'settings' };

export default function App() {
  const [authed, setAuthed] = useState(!!getToken() && !!getProfile());
  const [profile, setProfile] = useState<AdminProfile | null>(() => getProfile());
  const [section, setSection] = useState<Section>({ name: 'orders' });

  if (!authed || !profile) {
    return (
      <>
        <Login
          onLogin={(p) => {
            setProfile(p);
            setAuthed(true);
          }}
        />
        <ToastHost />
      </>
    );
  }

  const navItem = (label: string, icon: React.ReactNode, s: Section) => (
    <button
      onClick={() => setSection(s)}
      className={`flex items-center gap-2 rounded-md px-3 py-2 text-body transition-colors ${
        section.name === s.name ? 'bg-brand/10 text-brand' : 'text-ink-muted hover:bg-surface-2'
      }`}
    >
      {icon} {label}
    </button>
  );

  let content: React.ReactNode;
  switch (section.name) {
    case 'orders':
      content = <Orders onOpen={(id) => setSection({ name: 'order', id })} />;
      break;
    case 'order':
      content = <OrderDetail key={section.id} id={section.id} onBack={() => setSection({ name: 'orders' })} />;
      break;
    case 'tariffs':
      content = <Tariffs />;
      break;
    case 'clients':
      content = <Clients />;
      break;
    case 'settings':
      content = <Settings />;
      break;
  }

  return (
    <div className="min-h-dvh">
      {/* Верхняя панель */}
      <header className="sticky top-0 z-40 border-b border-line bg-surface">
        <div className="mx-auto flex max-w-6xl items-center justify-between gap-3 px-4 py-2.5">
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <div className="flex h-8 w-8 items-center justify-center rounded-md bg-brand/10 text-brand">
                <Package size={18} />
              </div>
              <span className="text-h2">Cargo · Админка</span>
            </div>
            <nav className="flex gap-1">
              {navItem('Заказы', null, { name: 'orders' })}
              {navItem('Клиенты', <Users size={16} />, { name: 'clients' })}
              {navItem('Тарифы', null, { name: 'tariffs' })}
              {navItem('Настройки', <SettingsIcon size={16} />, { name: 'settings' })}
            </nav>
          </div>
          <div className="flex items-center gap-3">
            <div className="text-right">
              <div className="text-body">{profile.display_name || profile.username}</div>
              <div className="text-caption text-ink-muted">
                {profile.role === 'admin' ? 'администратор' : 'оператор'}
              </div>
            </div>
            <button
              onClick={() => {
                clearAuth();
                window.location.reload();
              }}
              className="text-caption text-ink-muted hover:text-danger"
            >
              Выйти
            </button>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-4 py-4">{content}</main>
      <ToastHost />
    </div>
  );
}
