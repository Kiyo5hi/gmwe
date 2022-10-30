#!/usr/bin/env sh

caddy run --config /etc/caddy/Caddyfile --adapter caddyfile &
/app/api/api
