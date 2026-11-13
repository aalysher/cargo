/** Копирование текста в буфер + haptic. */
import { hapticSelection } from './telegram';

export async function copyText(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text);
    hapticSelection();
    return true;
  } catch {
    // Фолбэк для старых webview
    const ta = document.createElement('textarea');
    ta.value = text;
    ta.style.position = 'fixed';
    ta.style.opacity = '0';
    document.body.appendChild(ta);
    ta.select();
    const ok = document.execCommand('copy');
    document.body.removeChild(ta);
    if (ok) hapticSelection();
    return ok;
  }
}
