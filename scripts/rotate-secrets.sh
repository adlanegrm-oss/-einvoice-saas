#!/usr/bin/env bash
set -euo pipefail

ENV_FILE="${1:-.env}"

if [ ! -f "$ENV_FILE" ]; then
    echo "[-] Fichier $ENV_FILE introuvable."
    exit 1
fi

NEW_JWT=$(openssl rand -base64 32)
NEW_ADMIN_PASS=$(openssl rand -base64 24)

sed -i.bak -E "s|^JWT_SECRET=.*|JWT_SECRET=${NEW_JWT}|" "$ENV_FILE"
sed -i.bak -E "s|^ADMIN_PASSWORD=.*|ADMIN_PASSWORD=${NEW_ADMIN_PASS}|" "$ENV_FILE"

rm -f "${ENV_FILE}.bak"
echo "[+] Secrets renouveles avec succes dans $ENV_FILE."
