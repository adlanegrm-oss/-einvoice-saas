#!/usr/bin/env bash
set -euo pipefail

DB_PATH="${DB_PATH:-/data/einvoice.db}"
BACKUP_DIR="${BACKUP_DIR:-/backups}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
TARGET="${BACKUP_DIR}/einvoice_${TIMESTAMP}.db"

mkdir -p "$BACKUP_DIR"

if command -v sqlite3 &> /dev/null; then
    echo "[*] Execution du snapshot sqlite3 .backup..."
    sqlite3 "$DB_PATH" ".backup '${TARGET}'"
else
    echo "[-] ERREUR: sqlite3 CLI requis pour eviter la corruption."
    exit 1
fi

gzip -9 "$TARGET"
echo "[+] Sauvegarde terminee : ${TARGET}.gz"
find "$BACKUP_DIR" -name "einvoice_*.db.gz" -mtime +7 -delete
