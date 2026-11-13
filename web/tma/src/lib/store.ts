/** Локальное состояние: токен, dev-личность, флаг онбординга. */

const TOKEN_KEY = 'cargo_token';
const DEV_TG_KEY = 'cargo_dev_tg';
const ONBOARDED_KEY = 'cargo_onboarded';

export function getToken(): string {
  return localStorage.getItem(TOKEN_KEY) ?? '';
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}

/** Стабильный dev-telegram_id (вне Telegram). */
export function devTelegramId(): number {
  let id = Number(localStorage.getItem(DEV_TG_KEY));
  if (!id) {
    id = 1_000_000_000 + Math.floor(Math.random() * 8_999_999_999);
    localStorage.setItem(DEV_TG_KEY, String(id));
  }
  return id;
}

export function isOnboarded(): boolean {
  return localStorage.getItem(ONBOARDED_KEY) === '1';
}

export function setOnboarded() {
  localStorage.setItem(ONBOARDED_KEY, '1');
}
