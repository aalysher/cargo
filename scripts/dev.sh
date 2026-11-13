#!/bin/sh
# Полный локальный стек для разработки: db + api + worker + admin (docker compose),
# Cloudflare Tunnel → API, фронты (vite).
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "==> Поднимаем db + api + worker + admin через docker compose"
docker compose up -d db api worker admin

echo "==> Запускаем Cloudflare Tunnel → http://localhost:8085"
if pgrep -f "cloudflared tunnel" > /dev/null; then
  echo "cloudflared уже запущен — пропускаем"
else
  nohup /tmp/opencode/cloudflared tunnel --url http://localhost:8085 > /tmp/opencode/cloudflared-dev.log 2>&1 &
fi

echo "==> Запускаем фронты: TMA :5175, админка :5176"
nohup npm --prefix web/tma run dev > /tmp/opencode/vite-tma.log 2>&1 &
nohup npm --prefix web/admin run dev > /tmp/opencode/vite-admin.log 2>&1 &

echo ""
echo "Локальный стек поднят:"
echo "  - API:      http://localhost:8085"
echo "  - TMA:      http://localhost:5175"
echo "  - Admin:    http://localhost:5176"
echo "  - Tunnel:   https://<id>.trycloudflare.com (см. /tmp/opencode/cloudflared-dev.log)"
echo "  - API через tunnel: https://<id>.trycloudflare.com/api/v1"
