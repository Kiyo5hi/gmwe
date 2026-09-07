#!/usr/bin/env sh

caddy run --config /etc/caddy/Caddyfile --adapter caddyfile &
exec /app/api/api
