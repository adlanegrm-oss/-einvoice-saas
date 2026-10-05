#!/usr/bin/env bash
set -euo pipefail

# Script PRA : Sauvegarde à chaud chiffrée de la base SQLite et des archives
# RPO : 1h | RTO : 15 minutes

BACKUP_DIR="${BACKUP_DIR:-/var/backups/einvoice}"
DB_PATH="${DB_PATH:-./data/invoices.db}"
ENCRYPTION_KEY="${BACKUP_ENCRYPTION_KEY:-secret_pra_key_2026}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
TARGET_FILE="${BACKUP_DIR}/backup_${TIMESTAMP}.tar.gz.enc"

mkdir -p "${BACKUP_DIR}"

echo "[INFO] Démarrage du backup SQLite à chaud..."
TMP_DB=$(mktemp)
sqlite3 "${DB_PATH}" ".backup '${TMP_DB}'"

echo "[INFO] Compression et chiffrement AES-256..."
tar -czf - -C "$(dirname "${TMP_DB}")" "$(basename "${TMP_DB}")" | \
  openssl enc -aes-256-cbc -salt -pbkdf2 -pass "pass:${ENCRYPTION_KEY}" -out "${TARGET_FILE}"

rm -f "${TMP_DB}"
echo "[OK] Backup sécurisé généré : ${TARGET_FILE}"
