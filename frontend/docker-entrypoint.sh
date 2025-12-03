#!/bin/sh

# Map simple env vars to Nuxt public env vars
export NUXT_PUBLIC_BASE_URL="${BASE_URL:-${NUXT_PUBLIC_BASE_URL:-http://localhost:3000}}"
export NUXT_PUBLIC_API_URL="${API_URL:-${NUXT_PUBLIC_API_URL:-http://localhost:8080}}"

exec "$@"
