-- Migration: Add phone_no column to users table and create whatsapp_outbox table
-- Date: 2026-06-23
-- Description: Implements Transactional Outbox Pattern for reliable WhatsApp notifications

-- 1. Tambah kolom phone_no ke tabel users (opsional/nullable)
SET @col_exists = (SELECT COUNT(*) FROM information_schema.columns 
    WHERE table_schema = DATABASE() AND table_name = 'users' AND column_name = 'phone_no');

SET @sql = IF(@col_exists = 0, 
    'ALTER TABLE users ADD COLUMN phone_no VARCHAR(20) NULL DEFAULT NULL AFTER email',
    'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 2. Buat tabel whatsapp_outbox untuk antrian pesan
CREATE TABLE IF NOT EXISTS whatsapp_outbox (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    phone_no VARCHAR(20) NOT NULL COMMENT 'Nomor telepon tujuan',
    message TEXT NOT NULL COMMENT 'Isi pesan WhatsApp',
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' COMMENT 'PENDING, SENDING, SUCCESS, FAILED, ABANDONED',
    retry_count INT NOT NULL DEFAULT 0 COMMENT 'Jumlah percobaan pengiriman',
    max_retries INT NOT NULL DEFAULT 5 COMMENT 'Batas maksimal percobaan',
    last_error TEXT NULL COMMENT 'Pesan error terakhir',
    ref_type VARCHAR(50) NULL COMMENT 'Tipe referensi (e.g. trouble_ticket)',
    ref_id BIGINT UNSIGNED NULL COMMENT 'ID referensi',
    created_by BIGINT UNSIGNED NULL COMMENT 'User ID yang membuat notifikasi',
    sent_at DATETIME NULL COMMENT 'Waktu berhasil terkirim',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_wa_outbox_status (status),
    INDEX idx_wa_outbox_phone (phone_no),
    INDEX idx_wa_outbox_ref (ref_type, ref_id),
    INDEX idx_wa_outbox_created_at (created_at),
    INDEX idx_wa_outbox_retry (status, retry_count)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
