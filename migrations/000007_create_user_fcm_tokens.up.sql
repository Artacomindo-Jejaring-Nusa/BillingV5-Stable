-- Migration: Create user_fcm_tokens table
-- Description: Stores FCM tokens for user push notifications

CREATE TABLE IF NOT EXISTS user_fcm_tokens (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL COMMENT 'ID User pemilik perangkat',
    token VARCHAR(255) NOT NULL COMMENT 'FCM Device Token',
    device_type VARCHAR(50) NOT NULL DEFAULT 'web' COMMENT 'android, ios, web',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    UNIQUE KEY idx_user_fcm_token_hash (token),
    INDEX idx_user_fcm_tokens_user (user_id),
    INDEX idx_user_fcm_tokens_deleted_at (deleted_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
