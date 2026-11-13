/** Конфиг админки. */
export const API_BASE: string =
  (import.meta.env.VITE_API_URL as string | undefined) ?? 'http://localhost:8085/api/v1';
