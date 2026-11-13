/** Обёртка Telegram WebApp (безопасно работает и вне Telegram — dev-браузер). */

interface WebApp {
  initData: string;
  initDataUnsafe: {
    user?: { id: number; first_name: string; last_name?: string; username?: string; photo_url?: string };
    start_param?: string;
  };
  version: string;
  colorScheme: 'light' | 'dark';
  ready: () => void;
  expand: () => void;
  MainButton: {
    setText: (t: string) => void;
    onClick: (cb: () => void) => void;
    offClick: (cb: () => void) => void;
    show: () => void;
    hide: () => void;
    enable: () => void;
    disable: () => void;
    showProgress: (leaveActive?: boolean) => void;
    hideProgress: () => void;
    setParams: (p: { color?: string; text_color?: string; is_visible?: boolean; is_active?: boolean }) => void;
  };
  BackButton: {
    onClick: (cb: () => void) => void;
    offClick: (cb: () => void) => void;
    show: () => void;
    hide: () => void;
  };
  HapticFeedback?: {
    impactOccurred: (style: 'light' | 'medium' | 'heavy') => void;
    notificationOccurred: (type: 'error' | 'success' | 'warning') => void;
    selectionChanged: () => void;
  };
  openTelegramLink: (url: string) => void;
  onEvent: (event: string, cb: () => void) => void;
  offEvent: (event: string, cb: () => void) => void;
}

declare global {
  interface Window {
    Telegram?: { WebApp?: WebApp };
  }
}

/** Инстанс WebApp или null (вне Telegram). */
export function webApp(): WebApp | null {
  return window.Telegram?.WebApp ?? null;
}

/** Открыто ли как Mini App (есть непустой initData). */
export function isTMA(): boolean {
  return !!webApp()?.initData;
}

/** Инициализация: ready/expand, тема. Возвращает colorScheme. */
export function initTelegram(): 'light' | 'dark' {
  const tg = webApp();
  if (!tg) return 'light';
  tg.ready();
  tg.expand();
  applyTheme(tg.colorScheme);
  const onTheme = () => applyTheme(tg.colorScheme);
  tg.onEvent('themeChanged', onTheme);
  return tg.colorScheme;
}

function applyTheme(scheme: 'light' | 'dark') {
  document.documentElement.dataset.theme = scheme;
  const meta = document.querySelector('meta[name="theme-color"]');
  if (meta) meta.setAttribute('content', scheme === 'dark' ? '#0B1220' : '#F8FAFC');
}

/** start_param из deep link (t.me/<bot>?startapp=order_2 → "order_2"). */
export function startParam(): string | undefined {
  return webApp()?.initDataUnsafe.start_param;
}

/** Цвет MainButton из токена бренда (тема уже применена). */
/**
 * Хук-примитивы не используем напрямую в компонентах — они в hooks/useTelegramUI.ts.
 * Здесь — низкоуровневые действия.
 */

export function hapticImpact(style: 'light' | 'medium' | 'heavy' = 'light') {
  webApp()?.HapticFeedback?.impactOccurred(style);
}

export function hapticSelection() {
  webApp()?.HapticFeedback?.selectionChanged();
}

export function hapticNotify(type: 'error' | 'success' | 'warning' = 'success') {
  webApp()?.HapticFeedback?.notificationOccurred(type);
}

export function openBotChat(username: string) {
  const tg = webApp();
  if (tg) tg.openTelegramLink(`https://t.me/${username}`);
  else window.open(`https://t.me/${username}`, '_blank');
}
