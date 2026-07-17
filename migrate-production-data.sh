#!/bin/bash
# =============================================================================
# migrate-production-data.sh
# Script untuk mengganti data dummy di BillingRevaktor Docker
# dengan data dump dari sistem existing (legacy Python)
#
# Penggunaan:
#   ./migrate-production-data.sh <path-to-sql-dump>
#
# Contoh:
#   ./migrate-production-data.sh ./DataTerbaru-Juli2026_Cleaned.sql
#
# PENTING: Jalankan dari folder BillingSystem-CICD
# =============================================================================

set -euo pipefail

# === Warna untuk output ===
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# === Konfigurasi ===
MYSQL_CONTAINER="mysql-master"
SLAVE_CONTAINER="mysql-slave"
BACKEND1_CONTAINER="billing-backend-1"
BACKEND2_CONTAINER="billing-backend-2"
DB_NAME="billing_revaktor"
DB_USER="root"
DB_PASSWORD="${DB_PASSWORD:-root}"
BACKUP_DIR="./backups"

# === Fungsi Helper ===
log_info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
log_ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# === Validasi Argumen ===
if [ $# -lt 1 ]; then
    log_error "Usage: $0 <path-to-sql-dump>"
    echo ""
    echo "  Contoh: $0 ./DataTerbaru-Juli2026_Cleaned.sql"
    echo ""
    exit 1
fi

SQL_DUMP_FILE="$1"

if [ ! -f "$SQL_DUMP_FILE" ]; then
    log_error "File SQL tidak ditemukan: $SQL_DUMP_FILE"
    exit 1
fi

FILE_SIZE=$(du -h "$SQL_DUMP_FILE" | cut -f1)
log_info "File SQL ditemukan: $SQL_DUMP_FILE ($FILE_SIZE)"

# === Cek Docker containers berjalan ===
log_info "Mengecek status container..."
if ! docker ps --format '{{.Names}}' | grep -q "$MYSQL_CONTAINER"; then
    log_error "Container $MYSQL_CONTAINER tidak berjalan!"
    echo "  Jalankan dulu: docker compose up -d mysql-master"
    exit 1
fi
log_ok "Container $MYSQL_CONTAINER berjalan."

# === Konfirmasi ===
echo ""
echo "============================================"
echo -e "${YELLOW}  ⚠️  PERINGATAN: OPERASI DESTRUKTIF${NC}"
echo "============================================"
echo ""
echo "  Script ini akan:"
echo "  1. Backup database yang ada sekarang"
echo "  2. HAPUS SEMUA DATA di database $DB_NAME"
echo "  3. Import data dari: $SQL_DUMP_FILE"
echo "  4. Restart backend agar auto-migrate berjalan"
echo ""
echo "  Server: $(hostname)"
echo "  Database: $DB_NAME"
echo "  File: $SQL_DUMP_FILE ($FILE_SIZE)"
echo ""
read -p "  Lanjutkan? (ketik 'YA' untuk konfirmasi): " CONFIRM
if [ "$CONFIRM" != "YA" ]; then
    log_warn "Dibatalkan oleh user."
    exit 0
fi

# === Step 1: Backup data yang ada ===
log_info "=== STEP 1: Backup database yang ada ==="
mkdir -p "$BACKUP_DIR"
BACKUP_FILE="$BACKUP_DIR/backup_${DB_NAME}_$(date +%Y%m%d_%H%M%S).sql"

log_info "Membuat backup ke: $BACKUP_FILE"
docker exec "$MYSQL_CONTAINER" mysqldump \
    -u"$DB_USER" -p"$DB_PASSWORD" \
    --single-transaction \
    --routines \
    --triggers \
    "$DB_NAME" > "$BACKUP_FILE" 2>/dev/null

BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
log_ok "Backup selesai: $BACKUP_FILE ($BACKUP_SIZE)"

# === Step 2: Stop backend agar tidak ada koneksi aktif ===
log_info "=== STEP 2: Stop backend services ==="
docker stop "$BACKEND1_CONTAINER" "$BACKEND2_CONTAINER" 2>/dev/null || true
log_ok "Backend containers stopped."

# === Step 3: Stop slave replication ===
log_info "=== STEP 3: Stop slave replication ==="
if docker ps --format '{{.Names}}' | grep -q "$SLAVE_CONTAINER"; then
    docker exec "$SLAVE_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASSWORD" \
        -e "STOP SLAVE;" 2>/dev/null || true
    log_ok "Slave replication stopped."
else
    log_warn "Slave container tidak berjalan, skip."
fi

# === Step 4: Drop dan recreate database ===
log_info "=== STEP 4: Reset database ==="
docker exec "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASSWORD" \
    -e "DROP DATABASE IF EXISTS $DB_NAME; CREATE DATABASE $DB_NAME;" 2>/dev/null
log_ok "Database $DB_NAME di-reset."

# === Step 5: Copy SQL file ke container dan import ===
log_info "=== STEP 5: Import data produksi ==="
log_info "Mengcopy file SQL ke container..."
docker cp "$SQL_DUMP_FILE" "$MYSQL_CONTAINER":/tmp/production-data.sql

log_info "Mengimport data (ini bisa memakan waktu beberapa menit)..."
IMPORT_START=$(date +%s)

docker exec "$MYSQL_CONTAINER" mysql \
    -u"$DB_USER" -p"$DB_PASSWORD" \
    "$DB_NAME" \
    --init-command='SET FOREIGN_KEY_CHECKS=0; SET UNIQUE_CHECKS=0; SET AUTOCOMMIT=0;' \
    -e "source /tmp/production-data.sql; COMMIT;" 2>/dev/null

IMPORT_END=$(date +%s)
IMPORT_DURATION=$((IMPORT_END - IMPORT_START))
log_ok "Import selesai dalam ${IMPORT_DURATION} detik."

# === Step 6: Cleanup temp file di container ===
docker exec "$MYSQL_CONTAINER" rm -f /tmp/production-data.sql
log_ok "Temp file dihapus dari container."

# === Step 7: Verifikasi data ===
log_info "=== STEP 6: Verifikasi data ==="
echo ""
echo "  Tabel                | Jumlah Record"
echo "  ---------------------|---------------"

docker exec "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" -N -e "
SELECT 'users', COUNT(*) FROM users
UNION ALL SELECT 'pelanggan', COUNT(*) FROM pelanggan
UNION ALL SELECT 'data_teknis', COUNT(*) FROM data_teknis
UNION ALL SELECT 'langganan', COUNT(*) FROM langganan
UNION ALL SELECT 'invoices', COUNT(*) FROM invoices
UNION ALL SELECT 'trouble_ticket', COUNT(*) FROM trouble_ticket
UNION ALL SELECT 'paket_layanan', COUNT(*) FROM paket_layanan;
" 2>/dev/null | while IFS=$'\t' read -r tabel jumlah; do
    printf "  %-21s | %s\n" "$tabel" "$jumlah"
done

echo ""

# === Step 8: Start backend (auto-migrate akan berjalan) ===
log_info "=== STEP 7: Start backend services ==="
docker start "$BACKEND1_CONTAINER" "$BACKEND2_CONTAINER"
log_ok "Backend services started."

log_info "Menunggu backend selesai auto-migrate (15 detik)..."
sleep 15

# Cek apakah backend masih running
if docker ps --format '{{.Names}}' | grep -q "$BACKEND1_CONTAINER"; then
    log_ok "Backend-1 berjalan normal."
else
    log_error "Backend-1 tidak berjalan! Cek log: docker logs $BACKEND1_CONTAINER"
fi

# === Step 9: Reset slave replication ===
log_info "=== STEP 8: Reset slave replication ==="
if docker ps --format '{{.Names}}' | grep -q "$SLAVE_CONTAINER"; then
    # Get new master position
    MASTER_STATUS=$(docker exec "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASSWORD" -N -e "SHOW MASTER STATUS;" 2>/dev/null)
    MASTER_FILE=$(echo "$MASTER_STATUS" | awk '{print $1}')
    MASTER_POS=$(echo "$MASTER_STATUS" | awk '{print $2}')

    log_info "Master position: $MASTER_FILE @ $MASTER_POS"

    # Re-import data ke slave juga
    log_info "Mengimport data ke slave..."
    docker cp "$SQL_DUMP_FILE" "$SLAVE_CONTAINER":/tmp/production-data.sql 2>/dev/null || true
    docker exec "$SLAVE_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASSWORD" \
        -e "DROP DATABASE IF EXISTS $DB_NAME; CREATE DATABASE $DB_NAME;" 2>/dev/null || true
    docker exec "$SLAVE_CONTAINER" mysql \
        -u"$DB_USER" -p"$DB_PASSWORD" \
        "$DB_NAME" \
        --init-command='SET FOREIGN_KEY_CHECKS=0;' \
        -e "source /tmp/production-data.sql;" 2>/dev/null || true
    docker exec "$SLAVE_CONTAINER" rm -f /tmp/production-data.sql 2>/dev/null || true

    # Reconfigure replication
    docker exec "$SLAVE_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASSWORD" -e "
        STOP SLAVE;
        RESET SLAVE ALL;
        CHANGE MASTER TO
            MASTER_HOST='mysql-master',
            MASTER_USER='repl_user',
            MASTER_PASSWORD='repl_password',
            MASTER_LOG_FILE='$MASTER_FILE',
            MASTER_LOG_POS=$MASTER_POS,
            GET_MASTER_PUBLIC_KEY=1;
        START SLAVE;
    " 2>/dev/null || true

    # Verify slave status
    SLAVE_IO=$(docker exec "$SLAVE_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASSWORD" -N -e "SHOW SLAVE STATUS\G" 2>/dev/null | grep "Slave_IO_Running" | awk '{print $2}')
    SLAVE_SQL=$(docker exec "$SLAVE_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASSWORD" -N -e "SHOW SLAVE STATUS\G" 2>/dev/null | grep "Slave_SQL_Running" | head -1 | awk '{print $2}')

    if [ "$SLAVE_IO" = "Yes" ] && [ "$SLAVE_SQL" = "Yes" ]; then
        log_ok "Slave replication berjalan normal."
    else
        log_warn "Slave replication mungkin perlu dicek manual."
        log_warn "  IO: $SLAVE_IO, SQL: $SLAVE_SQL"
    fi
else
    log_warn "Slave container tidak berjalan, skip replication setup."
fi

# === Selesai ===
echo ""
echo "============================================"
echo -e "${GREEN}  ✅ MIGRASI DATA SELESAI!${NC}"
echo "============================================"
echo ""
echo "  Database  : $DB_NAME"
echo "  Server    : $(hostname)"
echo "  Backup    : $BACKUP_FILE"
echo "  Durasi    : ${IMPORT_DURATION} detik"
echo ""
echo "  Akses aplikasi:"
echo "    - Frontend : http://localhost:3000 atau https://jpo.jelantik.com"
echo "    - API      : http://localhost:8000/api/v1/"
echo ""
echo "  Jika ada masalah, restore dari backup:"
echo "    docker exec -i $MYSQL_CONTAINER mysql -u$DB_USER -p$DB_PASSWORD $DB_NAME < $BACKUP_FILE"
echo ""
