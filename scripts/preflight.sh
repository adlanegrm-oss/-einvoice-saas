#!/usr/bin/env bash
set -euo pipefail

echo "=== [PREFLIGHT CHECK] ==="

if [ ! -f .env ]; then
    echo "[-] ERREUR: Fichier .env manquant. Copiez .env.example.hardened vers .env"
    exit 1
fi

if grep -q "admin-password-123" .env || grep -q "JWT_SECRET=9NzZFArr" .env; then
    echo "[-] ERREUR CRITIQUE: Secrets par defaut detectes dans .env !"
    exit 1
fi

for cmd in docker openssl; do
    if ! command -v "$cmd" &> /dev/null; then
        echo "[-] ERREUR: Dependance absente : $cmd"
        exit 1
    fi
done

echo "[+] Prerequis valides avec succes."
