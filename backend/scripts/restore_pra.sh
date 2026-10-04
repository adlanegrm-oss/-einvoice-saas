#!/usr/bin/env bash
set -euo pipefail

# Script PRA : Restauration d'un backup chiffré
BACKUP_FILE="${1:-}"
DB_TARGET="${DB_TARGET:-./data/invoices.db}"
ENCRYPTION_KEY="${BACKUP_ENCRYPTION_KEY:-secret_pra_key_2026}"

if [ -z "${BACKUP_FILE}" ] || [ ! -f "${BACKUP_FILE}" ]; then
  echo "[ERREUR] Usage: $0 <chemin_vers_backup.tar.gz.enc>"
  exit 1
fi

echo "[INFO] Déchiffrement et restauration..."
RESTORE_DIR=$(mktemp -d)
openssl enc -d -aes-256-cbc -pbkdf2 -pass "pass:${ENCRYPTION_KEY}" -in "${BACKUP_FILE}" | \
  tar -xzf - -C "${RESTORE_DIR}"

RESTORED_DB=$(find "${RESTORE_DIR}" -type f | head -n 1)
cp "${RESTORED_DB}" "${DB_TARGET}"
rm -rf "${RESTORE_DIR}"

echo "[OK] Base SQLite restaurée avec succès dans ${DB_TARGET}"
