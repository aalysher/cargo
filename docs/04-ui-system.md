# 04. UI-система (дизайн-система TMA)

## Принципы
1. Доверие через спокойствие: нейтральная база, один фирменный цвет, без кислотных градиентов.
2. Данные важнее декора: цифры (вес, цена, ETA) — главные герои, tabular-nums.
3. Нативность TMA: тема Telegram (light/dark) подхватывается; наши токены — фолбэк.
4. Зона пальца: действия ≥44px, ключевые — снизу (MainButton).

## Палитра

### Бренд и семантика
| Токен | Light | Dark |
|---|---|---|
| --brand-500 | #2563EB | #3B82F6 |
| --brand-600 | #1D4ED8 | #60A5FA |
| --accent-500 | #F97316 | #FB923C |
| --success-500 | #16A34A | #22C55E |
| --warning-500 | #D97706 | #FBBF24 |
| --danger-500 | #DC2626 | #EF4444 |

### Поверхности и текст
| Токен | Light | Dark |
|---|---|---|
| --bg | #F8FAFC | #0B1220 |
| --surface | #FFFFFF | #111A2B |
| --surface-2 | #F1F5F9 | #18233A |
| --border | #E2E8F0 | #1E293B |
| --text | #0F172A | #F1F5F9 |
| --text-muted | #64748B | #94A3B8 |
| --overlay | rgba(15,23,42,.45) | rgba(0,0,0,.6) |

### Статусы заказа
| Статус | Hex (light / dark) |
|---|---|
| Новая | #64748B / #94A3B8 |
| Принята | #2563EB / #60A5FA |
| На складе (Китай) | #7C3AED / #A78BFA |
| Отгружено | #D97706 / #FBBF24 |
| В пути | #F97316 / #FB923C |
| Прибыло в Бишкек | #0D9488 / #2DD4BF |
| Готово к выдаче | #16A34A / #22C55E |
| Выдано | #15803D / #4ADE80 |
| Отменено | #DC2626 / #EF4444 |

## Типографика
Системный стек (нативно, как Telegram): `-apple-system, "SF Pro Text", Roboto, "Segoe UI", sans-serif`. Inter — опция для админки.

| Токен | Размер/высота | Вес | Где |
|---|---|---|---|
| display | 28/34 | 700 | цена, статус крупно |
| h1 | 22/28 | 700 | заголовки экранов |
| h2 | 17/24 | 600 | заголовки карточек |
| body | 15/22 | 400 | основной текст |
| body-strong | 15/22 | 600 | акцентные строки |
| caption | 13/18 | 400 | вторичное |
| micro | 11/14 | 600 | бейджи (uppercase, ls 0.04em) |

Цены/вес — `font-variant-numeric: tabular-nums`.

## Радиусы, тени, отступы
- Радиусы: sm 8, md 12 (кнопки), lg 16 (карточки), xl 24 (sheet сверху), full (бейджи).
- Тени (light; dark — границы): `--shadow-1: 0 1px 2px rgba(15,23,42,.06)`, `--shadow-2: 0 4px 16px rgba(15,23,42,.08)`, `--shadow-3: 0 12px 32px rgba(15,23,42,.12)`.
- Отступы: база 4; экран 16 по бокам; между карточками 12; внутри карточки 16.

## Иконки
Lucide (MIT, outline): сетка 24px, stroke 1.75. Нужные: package, package-check, truck, plane, warehouse, map-pin, copy, calculator, message-circle, bell, user, chevron-right, plus, clock.

## Компоненты
- Кнопки: primary/secondary/ghost/danger; высоты 48/44/36; радиус 12; прессинг opacity + scale .98.
- Карточка посылки: бейдж, номер (tabular), дата, вес, прогресс 8 сегментов, chevron; тень shadow-1.
- Таймлайн: вертикальный; точка 12px (пройдено — залита, текущее — пульс, будущее — border), линия 2px, подписи + даты.
- Бейдж статуса: full-radius, фон статуса @12%, текст — цвет статуса, 11px/600/uppercase.
- Формы: лейбл сверху (caption), инпут 44px, радиус 12, фокус — border brand + ring; ошибка inline под полем.
- Chips: 32px, full-radius, selected = brand @10%.
- Bottom sheet: радиус 24 сверху, drag-handle, shadow-3.
- Таббар: 4 пункта, иконка 24 + micro-подпись.
- Скелетоны: surface-2, pulse 1.5s. Toast: снизу над таббаром, 3s.
- Empty state: иконка 48 (muted), заголовок h2, подпись caption, CTA.

## Анимации
- Длительности: 150/200/300ms; easing `cubic-bezier(0.2,0.8,0.2,1)`.
- Смена статуса: сегмент заливается 300ms + `hapticFeedback.impactLight`; «Готово к выдаче» — `notificationSuccess`.
- Копирование: иконка → check 1s + `selectionChanged`.
- Списки: fade + 8px вверх, каскад 40ms (только первый экран).
- `prefers-reduced-motion` — отключить всё, кроме opacity.

## CSS-переменные (styles/tokens.css)
```css
:root {
  --brand-500: #2563EB; --brand-600: #1D4ED8;
  --accent-500: #F97316;
  --success-500: #16A34A; --warning-500: #D97706; --danger-500: #DC2626;
  --bg: #F8FAFC; --surface: #FFFFFF; --surface-2: #F1F5F9;
  --border: #E2E8F0;
  --text: #0F172A; --text-muted: #64748B;
  --overlay: rgba(15, 23, 42, 0.45);
  --st-new: #64748B;      --st-accepted: #2563EB; --st-warehouse: #7C3AED;
  --st-shipped: #D97706;  --st-transit: #F97316;  --st-arrived: #0D9488;
  --st-ready: #16A34A;    --st-delivered: #15803D; --st-cancelled: #DC2626;
  --r-sm: 8px; --r-md: 12px; --r-lg: 16px; --r-xl: 24px;
  --shadow-1: 0 1px 2px rgba(15,23,42,.06);
  --shadow-2: 0 4px 16px rgba(15,23,42,.08);
  --shadow-3: 0 12px 32px rgba(15,23,42,.12);
  --ease: cubic-bezier(0.2, 0.8, 0.2, 1);
}
[data-theme="dark"] {
  --brand-500: #3B82F6; --brand-600: #60A5FA;
  --accent-500: #FB923C;
  --success-500: #22C55E; --warning-500: #FBBF24; --danger-500: #EF4444;
  --bg: #0B1220; --surface: #111A2B; --surface-2: #18233A;
  --border: #1E293B;
  --text: #F1F5F9; --text-muted: #94A3B8;
  --overlay: rgba(0, 0, 0, 0.6);
  --st-new: #94A3B8; --st-accepted: #60A5FA; --st-warehouse: #A78BFA;
  --st-shipped: #FBBF24; --st-transit: #FB923C; --st-arrived: #2DD4BF;
  --st-ready: #22C55E; --st-delivered: #4ADE80; --st-cancelled: #EF4444;
  --shadow-1: none; --shadow-2: none; --shadow-3: 0 12px 32px rgba(0,0,0,.5);
}
```

## tailwind.config.js
```js
/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  darkMode: ["selector", '[data-theme="dark"]'],
  theme: {
    extend: {
      colors: {
        brand: { DEFAULT: "var(--brand-500)", strong: "var(--brand-600)" },
        accent: "var(--accent-500)",
        success: "var(--success-500)",
        warning: "var(--warning-500)",
        danger: "var(--danger-500)",
        bg: "var(--bg)",
        surface: { DEFAULT: "var(--surface)", 2: "var(--surface-2)" },
        line: "var(--border)",
        ink: { DEFAULT: "var(--text)", muted: "var(--text-muted)" },
        status: {
          new: "var(--st-new)", accepted: "var(--st-accepted)",
          warehouse: "var(--st-warehouse)", shipped: "var(--st-shipped)",
          transit: "var(--st-transit)", arrived: "var(--st-arrived)",
          ready: "var(--st-ready)", delivered: "var(--st-delivered)",
          cancelled: "var(--st-cancelled)",
        },
      },
      borderRadius: { sm: "var(--r-sm)", md: "var(--r-md)", lg: "var(--r-lg)", xl: "var(--r-xl)" },
      boxShadow: { 1: "var(--shadow-1)", 2: "var(--shadow-2)", 3: "var(--shadow-3)" },
      fontSize: {
        display: ["28px", { lineHeight: "34px", fontWeight: "700" }],
        h1: ["22px", { lineHeight: "28px", fontWeight: "700" }],
        h2: ["17px", { lineHeight: "24px", fontWeight: "600" }],
        body: ["15px", { lineHeight: "22px", fontWeight: "400" }],
        "body-strong": ["15px", { lineHeight: "22px", fontWeight: "600" }],
        caption: ["13px", { lineHeight: "18px", fontWeight: "400" }],
        micro: ["11px", { lineHeight: "14px", fontWeight: "600" }],
      },
      transitionTimingFunction: { out: "cubic-bezier(0.2,0.8,0.2,1)" },
      keyframes: {
        "fade-up": { from: { opacity: "0", transform: "translateY(8px)" }, to: { opacity: "1", transform: "none" } },
        "pulse-soft": { "0%, 100%": { opacity: "1" }, "50%": { opacity: "0.55" } },
      },
      animation: {
        "fade-up": "fade-up 200ms cubic-bezier(0.2,0.8,0.2,1) both",
        "pulse-soft": "pulse-soft 1.5s ease-in-out infinite",
      },
    },
  },
  plugins: [],
};
```

> Tailwind v3-совместимый формат (в v4 подключается через `@config`); при реализации можно перейти на v4 CSS-first (`@theme`), токены те же.
