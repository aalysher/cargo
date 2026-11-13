/** Хуки Telegram UI: MainButton и BackButton (no-op вне Telegram). */
import { useEffect, useRef } from 'react';
import { webApp } from '../lib/telegram';

/**
 * MainButton с фиксированным текстом. Обработчик можно менять на лету —
 * актуальный колбэк берётся из ref при каждом клике.
 */
export function useMainButton(
  text: string | null,
  onClick: () => void,
  opts: { enabled?: boolean; loading?: boolean } = {},
) {
  const cbRef = useRef(onClick);
  cbRef.current = onClick;
  const optsRef = useRef(opts);
  optsRef.current = opts;

  useEffect(() => {
    const tg = webApp();
    if (!tg) return;

    const handler = () => {
      if (optsRef.current.enabled !== false && !optsRef.current.loading) cbRef.current();
    };

    if (text === null) {
      tg.MainButton.offClick(handler);
      tg.MainButton.hide();
      return;
    }
    tg.MainButton.setText(text);
    tg.MainButton.setParams({
      color: getComputedStyle(document.documentElement).getPropertyValue('--brand-500').trim(),
      text_color: '#ffffff',
    });
    tg.MainButton.offClick(handler);
    tg.MainButton.onClick(handler);
    if (optsRef.current.loading) tg.MainButton.showProgress(false);
    else tg.MainButton.hideProgress();
    if (optsRef.current.enabled === false) tg.MainButton.disable();
    else tg.MainButton.enable();
    tg.MainButton.show();

    return () => {
      tg.MainButton.offClick(handler);
      tg.MainButton.hide();
    };
  }, [text, opts.enabled, opts.loading]);
}

/** BackButton: показать и повесить обработчик. */
export function useBackButton(onClick: () => void, visible = true) {
  const cbRef = useRef(onClick);
  cbRef.current = onClick;

  useEffect(() => {
    const tg = webApp();
    if (!tg) return;
    const handler = () => cbRef.current();

    if (!visible) {
      tg.BackButton.offClick(handler);
      tg.BackButton.hide();
      return;
    }
    tg.BackButton.offClick(handler);
    tg.BackButton.onClick(handler);
    tg.BackButton.show();

    return () => {
      tg.BackButton.offClick(handler);
      tg.BackButton.hide();
    };
  }, [visible]);
}
