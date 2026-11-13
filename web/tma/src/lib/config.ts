/** Конфигурация фронта (значения можно переопределить через VITE_*). */
export const API_BASE: string =
  (import.meta.env.VITE_API_URL as string | undefined) ?? 'http://localhost:8085/api/v1';

/** Username бота — для ссылок «написать в поддержку» и deep link. */
export const BOT_USERNAME: string =
  (import.meta.env.VITE_BOT_USERNAME as string | undefined) ?? 'cargo_demo_bot';
