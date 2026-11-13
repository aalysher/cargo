/** HTTP-клиент админки: Bearer JWT, единая ошибка. */
import { API_BASE } from './config';
import { getToken } from './store';

export class ApiError extends Error {
  code: string;
  detail: string;
  status: number;

  constructor(status: number, code: string, detail: string) {
    super(detail || code);
    this.status = status;
    this.code = code;
    this.detail = detail || code;
  }
}

interface Options {
  method?: 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE';
  body?: unknown;
}

export async function api<T>(path: string, opts: Options = {}): Promise<T> {
  const headers: Record<string, string> = { Accept: 'application/json' };
  const token = getToken();
  if (token) headers.Authorization = `Bearer ${token}`;
  if (opts.body !== undefined) headers['Content-Type'] = 'application/json';

  const res = await fetch(`${API_BASE}${path}`, {
    method: opts.method ?? 'GET',
    headers,
    body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
  });

  if (!res.ok) {
    // Сессия истекла — на логин (кроме самой авторизации).
    if (res.status === 401 && !path.startsWith('/admin/auth')) {
      clearAuth();
      window.location.reload();
    }
    let code = 'error';
    let detail = `Ошибка ${res.status}`;
    try {
      const data = (await res.json()) as { error?: string; detail?: string };
      code = data.error ?? code;
      detail = data.detail ?? detail;
    } catch {
      /* не JSON */
    }
    throw new ApiError(res.status, code, detail);
  }
  return (await res.json()) as T;
}

function clearAuth() {
  localStorage.removeItem('admin_token');
  localStorage.removeItem('admin_profile');
}
