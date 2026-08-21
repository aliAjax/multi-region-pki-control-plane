#!/usr/bin/env sh
set -eu
base="${BASE_URL:-http://127.0.0.1:8080}"
curl -fsS "$base/healthz" | grep -q '"status":"ok"'
curl -fsS "$base/readyz" | grep -q '"status":"ready"'
curl -fsS "$base/acme/directory" | grep -q 'newNonce'
curl -fsS -i "$base/acme/new-nonce" | grep -q 'Replay-Nonce'
echo "pki smoke passed"
