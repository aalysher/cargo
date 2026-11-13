/** Логин оператора/админа. */
import { useState } from 'react';
import { Package } from 'lucide-react';
import { api, type ApiError } from '../lib/api';
import { saveAuth, type AdminProfile } from '../lib/store';
import { Button } from '../components/primitives';

export function Login({ onLogin }: { onLogin: (p: AdminProfile) => void }) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (busy) return;
    setBusy(true);
    setError('');
    try {
      const res = await api<{ token: string; profile: AdminProfile }>('/admin/auth/login', {
        method: 'POST',
        body: { username: username.trim(), password },
      });
      saveAuth(res.token, res.profile);
      onLogin(res.profile);
    } catch (err) {
      setError((err as ApiError).detail || 'Ошибка входа');
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="flex min-h-dvh items-center justify-center bg-bg p-6">
      <form onSubmit={submit} className="w-full max-w-sm rounded-lg border border-line bg-surface p-6 shadow-2">
        <div className="mb-5 flex items-center gap-3">
          <div className="flex h-11 w-11 items-center justify-center rounded-lg bg-brand/10 text-brand">
            <Package size={24} />
          </div>
          <div>
            <div className="text-h2">Cargo — Админка</div>
            <div className="text-caption text-ink-muted">Китай → Бишкек</div>
          </div>
        </div>

        <label className="text-caption text-ink-muted">Логин</label>
        <input
          className="field mt-1 mb-3"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          autoComplete="username"
          autoFocus
        />

        <label className="text-caption text-ink-muted">Пароль</label>
        <input
          className="field mt-1 mb-4"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          autoComplete="current-password"
        />

        {error && <div className="mb-3 rounded-md bg-danger/10 px-3 py-2 text-caption text-danger">{error}</div>}

        <Button type="submit" disabled={busy || !username || !password} className="w-full">
          {busy ? 'Входим…' : 'Войти'}
        </Button>

        <p className="mt-4 text-center text-micro text-ink-muted">
          Демо: admin/admin123 (админ), operator/operator123 — сменить в проде
        </p>
      </form>
    </div>
  );
}
