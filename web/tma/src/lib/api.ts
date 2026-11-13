/** HTTP-клиент: Bearer-токен, единая ошибка ApiError{code,detail,status}. */
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
  method?: 'GET' | 'POST' | 'PATCH' | 'DELETE';
  body?: unknown;
}

/** Запрос к API; при ошибке бросает ApiError. */
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
    let code = 'error';
    let detail = `Ошибка ${res.status}`;
    try {
      const data = (await res.json()) as { error?: string; detail?: string };
      code = data.error ?? code;
      detail = data.detail ?? detail;
    } catch {
      /* тело не JSON */
    }
    throw new ApiError(res.status, code, detail);
  }
  return (await res.json()) as T;
}
