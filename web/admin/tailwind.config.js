/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
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
      },
      borderRadius: {
        sm: "var(--r-sm)",
        md: "var(--r-md)",
        lg: "var(--r-lg)",
        xl: "var(--r-xl)",
      },
      boxShadow: {
        1: "var(--shadow-1)",
        2: "var(--shadow-2)",
        3: "var(--shadow-3)",
      },
      fontSize: {
        display: ["28px", { lineHeight: "34px", fontWeight: "700" }],
        h1: ["22px", { lineHeight: "28px", fontWeight: "700" }],
        h2: ["17px", { lineHeight: "24px", fontWeight: "600" }],
        body: ["15px", { lineHeight: "22px", fontWeight: "400" }],
        "body-strong": ["15px", { lineHeight: "22px", fontWeight: "600" }],
        caption: ["13px", { lineHeight: "18px", fontWeight: "400" }],
        micro: ["11px", { lineHeight: "14px", fontWeight: "600" }],
      },
      keyframes: {
        "pulse-soft": {
          "0%, 100%": { opacity: "1" },
          "50%": { opacity: "0.55" },
        },
      },
      animation: {
        "pulse-soft": "pulse-soft 1.5s ease-in-out infinite",
      },
    },
  },
  plugins: [],
};
