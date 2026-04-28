-- 创建数据库
CREATE DATABASE IF NOT EXISTS recall_platform DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE recall_platform;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    recall_service_name VARCHAR(64) NOT NULL UNIQUE COMMENT '服务用户名',
    password VARCHAR(255) NOT NULL COMMENT '加密密码',
    api_token VARCHAR(64) UNIQUE COMMENT '用户API Token(长期有效)',
    notify_url VARCHAR(512) DEFAULT '' COMMENT '通知回调URL',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_recall_service_name (recall_service_name),
    INDEX idx_api_token (api_token)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Token表（JWT Token）
CREATE TABLE IF NOT EXISTS tokens (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    token VARCHAR(512) NOT NULL UNIQUE COMMENT 'Token值',
    expires_at DATETIME NOT NULL COMMENT '过期时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_token (token),
    INDEX idx_user_id (user_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- RefreshToken表
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    token VARCHAR(128) NOT NULL UNIQUE COMMENT 'RefreshToken值',
    expires_at DATETIME NOT NULL COMMENT '过期时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_token (token),
    INDEX idx_user_id (user_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 回调记录表
CREATE TABLE IF NOT EXISTS recall_records (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    recall_service_name VARCHAR(64) NOT NULL COMMENT '服务用户名',
    platform VARCHAR(64) NOT NULL COMMENT '平台来源',
    user_name VARCHAR(128) NOT NULL COMMENT '授权用户名称',
    params TEXT COMMENT '完整参数JSON',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_recall_service_name (recall_service_name),
    INDEX idx_platform (platform),
    INDEX idx_user_name (user_name),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
