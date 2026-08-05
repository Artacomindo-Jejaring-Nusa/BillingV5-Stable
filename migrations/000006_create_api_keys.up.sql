-- Migration: Create api_keys table
-- Description: Stores API Integration keys associated with system roles

CREATE TABLE IF NOT EXISTS api_keys (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT 'Deskripsi/Nama Integrasi',
    prefix VARCHAR(16) NOT NULL COMMENT 'Prefix API Key (e.g., jk_live_abc)',
    token_hash VARCHAR(64) NOT NULL COMMENT 'SHA-256 Hash dari token mentah',
    role_id BIGINT UNSIGNED NOT NULL COMMENT 'Role ID untuk hak akses API Key',
    is_active TINYINT(1) NOT NULL DEFAULT 1 COMMENT '1 = Aktif, 0 = Nonaktif',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    UNIQUE KEY idx_api_keys_token_hash (token_hash),
    INDEX idx_api_keys_deleted_at (deleted_at),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
