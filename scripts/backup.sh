#!/bin/bash
# NexusACG Database Backup Script
# Run via cron: 0 2 * * * /home/jct/nexusacg/scripts/backup.sh

set -euo pipefail

DB_NAME="nexusacg"
DB_USER="nexusacg"
DB_HOST="localhost"
DB_PORT="5432"
export PGPASSWORD="nexusacg_dev_pass"
BACKUP_DIR="/home/jct/nexusacg/backups"
RETENTION_DAYS=30
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"

echo "[$(date)] Starting database backup..."

pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -F c -f "$BACKUP_DIR/nexusacg_${DATE}.dump"

SIZE=$(du -h "$BACKUP_DIR/nexusacg_${DATE}.dump" | cut -f1)
echo "[$(date)] Backup complete: nexusacg_${DATE}.dump ($SIZE)"

# Clean up old backups
find "$BACKUP_DIR" -name "nexusacg_*.dump" -mtime +${RETENTION_DAYS} -delete
echo "[$(date)] Cleaned up backups older than ${RETENTION_DAYS} days"
