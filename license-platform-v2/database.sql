-- ============================================================
-- OpenClaw License Platform - 数据库初始化脚本
-- 兼容 MySQL 5.7 / 8.0
-- 导入方式：宝塔面板 → 数据库 → phpMyAdmin → 导入此文件
-- ============================================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- -----------------------------------------------------------
-- 1. 管理员用户表
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(191) NOT NULL,
  `email` varchar(191) NOT NULL,
  `email_verified_at` timestamp NULL DEFAULT NULL,
  `password` varchar(191) NOT NULL,
  `remember_token` varchar(100) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `users_email_unique` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入默认管理员（密码：admin123，登录后请立即修改！）
-- 如登录失败，在服务器执行：php scripts/reset_password.php
INSERT INTO `users` (`name`, `email`, `password`, `created_at`, `updated_at`) VALUES
('Admin', 'admin@openclaw.ai', '$2y$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', NOW(), NOW());

-- -----------------------------------------------------------
-- 2. 授权码表
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `licenses`;
CREATE TABLE `licenses` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `license_key` varchar(19) NOT NULL COMMENT '授权码格式：OC-XXXX-XXXX-XXXX',
  `tier` varchar(20) NOT NULL DEFAULT 'pro' COMMENT 'oss/pro/enterprise',
  `features` json DEFAULT NULL COMMENT 'JSON features list',
  `max_devices` int NOT NULL DEFAULT 3,
  `status` varchar(20) NOT NULL DEFAULT 'active' COMMENT 'active/expired/revoked',
  `expires_at` timestamp NULL DEFAULT NULL,
  `price_paid` decimal(10,2) NOT NULL DEFAULT 0.00,
  `customer_name` varchar(191) DEFAULT NULL,
  `customer_email` varchar(191) DEFAULT NULL,
  `notify_expiry` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否发送到期提醒',
  `notes` text DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `licenses_license_key_unique` (`license_key`),
  KEY `licenses_status_index` (`status`),
  KEY `licenses_tier_index` (`tier`),
  KEY `licenses_expires_at_index` (`expires_at`),
  KEY `licenses_customer_email_index` (`customer_email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------
-- 3. 设备绑定表
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `devices`;
CREATE TABLE `devices` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `license_id` bigint unsigned NOT NULL,
  `fingerprint_hash` varchar(64) NOT NULL COMMENT '设备指纹哈希',
  `device_name` varchar(100) NOT NULL DEFAULT 'Unknown',
  `os_info` varchar(50) DEFAULT NULL,
  `app_version` varchar(20) DEFAULT NULL,
  `last_active_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `devices_license_id_fingerprint_hash_unique` (`license_id`, `fingerprint_hash`),
  KEY `devices_fingerprint_hash_index` (`fingerprint_hash`),
  CONSTRAINT `devices_license_id_foreign` FOREIGN KEY (`license_id`) REFERENCES `licenses` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------
-- 4. 激活记录表
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `activations`;
CREATE TABLE `activations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `license_id` bigint unsigned NOT NULL,
  `action` varchar(20) NOT NULL COMMENT 'activate/deactivate',
  `fingerprint` varchar(64) DEFAULT NULL,
  `ip_address` varchar(45) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `activations_license_id_index` (`license_id`),
  KEY `activations_created_at_index` (`created_at`),
  CONSTRAINT `activations_license_id_foreign` FOREIGN KEY (`license_id`) REFERENCES `licenses` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------
-- 5. 使用统计表
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `usage_stats`;
CREATE TABLE `usage_stats` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `license_id` bigint unsigned NOT NULL,
  `stat_date` date NOT NULL,
  `memory_count` int NOT NULL DEFAULT 0,
  `active_users` int NOT NULL DEFAULT 0,
  `app_version` varchar(20) DEFAULT NULL,
  `os` varchar(20) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `usage_stats_license_id_stat_date_unique` (`license_id`, `stat_date`),
  KEY `usage_stats_stat_date_index` (`stat_date`),
  CONSTRAINT `usage_stats_license_id_foreign` FOREIGN KEY (`license_id`) REFERENCES `licenses` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------
-- 6. 安装统计表（匿名）
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `install_pings`;
CREATE TABLE `install_pings` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `install_id` varchar(36) NOT NULL,
  `version` varchar(20) DEFAULT NULL,
  `os` varchar(20) DEFAULT NULL,
  `ip_address` varchar(45) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `install_pings_install_id_index` (`install_id`),
  KEY `install_pings_created_at_index` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------
-- 7. 系统设置表（价格方案、支付配置等）
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `settings`;
CREATE TABLE `settings` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `key` varchar(100) NOT NULL,
  `value` text NOT NULL,
  `description` varchar(255) DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `settings_key_unique` (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入默认价格方案
INSERT INTO `settings` (`key`, `value`, `description`, `updated_at`) VALUES
('pricing_tiers', '{"oss":{"name":"OSS 免费版","max_devices":1,"price":0,"features":[]},"pro":{"name":"Pro 专业版","max_devices":3,"price":29,"features":["ai_extract","auto_graph","unlimited_graph","auto_decay","decay_report","prune_suggest","reinforce","conflict_scan","conflict_merge","smart_router","token_stats","wiki"]},"enterprise":{"name":"Enterprise 企业版","max_devices":10,"price":99,"features":["ai_extract","auto_graph","unlimited_graph","auto_decay","decay_report","prune_suggest","reinforce","conflict_scan","conflict_merge","smart_router","token_stats","wiki","api_access","sso","audit_log","time_travel","offline_mode"]}}', '价格方案配置 v3.0', NOW()),
('payment_method', 'manual', '支付方式: manual(手动)/alipay(支付宝)/wechat(微信)', NOW()),
('alipay_app_id', '', '支付宝App ID', NOW()),
('alipay_private_key', '', '支付宝应用私钥', NOW()),
('alipay_public_key', '', '支付宝公钥', NOW()),
('wechat_mch_id', '', '微信商户号', NOW()),
('wechat_api_key', '', '微信API密钥', NOW()),
('wechat_app_id', '', '微信App ID', NOW()),
('smtp_host', '', '邮件SMTP地址', NOW()),
('smtp_port', '465', 'SMTP端口', NOW()),
('smtp_user', '', 'SMTP用户名', NOW()),
('smtp_pass', '', 'SMTP密码', NOW()),
('smtp_from', '', '发件人地址', NOW()),
('expiry_notify_days', '7', '到期提前提醒天数', NOW());

-- -----------------------------------------------------------
-- 8. 登录日志表（安全审计）
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `login_logs`;
CREATE TABLE `login_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `email` varchar(191) NOT NULL,
  `ip_address` varchar(45) NOT NULL,
  `user_agent` varchar(500) DEFAULT NULL,
  `success` tinyint(1) NOT NULL DEFAULT 0,
  `created_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `login_logs_email_index` (`email`),
  KEY `login_logs_created_at_index` (`created_at`),
  KEY `login_logs_success_index` (`success`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

SET FOREIGN_KEY_CHECKS = 1;
