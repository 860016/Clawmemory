<?php
/**
 * OpenClaw License Platform - 配置文件
 * 纯 PHP 版本，无需 Laravel/Composer
 */

return [
    // 应用配置
    'app_name' => getenv('APP_NAME') ?: 'OpenClaw License Server',
    'app_debug' => filter_var(getenv('APP_DEBUG') ?: false, FILTER_VALIDATE_BOOLEAN),
    'app_url' => getenv('APP_URL') ?: 'http://localhost',

    // 数据库配置
    'db' => [
        'host' => getenv('DB_HOST') ?: '127.0.0.1',
        'port' => getenv('DB_PORT') ?: '3306',
        'database' => getenv('DB_DATABASE') ?: 'openclaw_license',
        'username' => getenv('DB_USERNAME') ?: 'root',
        'password' => getenv('DB_PASSWORD') ?: '',
        'charset' => 'utf8mb4',
        'collation' => 'utf8mb4_unicode_ci',
    ],

    // RSA 密钥路径
    'private_key_path' => __DIR__ . '/keys/private.pem',
    'public_key_path' => __DIR__ . '/keys/public.pem',

    // 默认授权码配置（数据库 settings 表优先）
    'license' => [
        'tiers' => [
            'oss' => [
                'name' => 'OSS',
                'max_devices' => 1,
                'price' => 0,
                'features' => [],
            ],
            'pro' => [
                'name' => 'Pro',
                'max_devices' => 3,
                'price' => 29,
                'features' => ['graph', 'backup', 'decay', 'routing', 'conflict'],
            ],
            'enterprise' => [
                'name' => 'Enterprise',
                'max_devices' => 10,
                'price' => 99,
                'features' => ['graph', 'backup', 'decay', 'routing', 'conflict', 'api', 'sso', 'audit', 'timetravel', 'offline'],
            ],
        ],
    ],

    // 分页
    'per_page' => 20,

    // Session
    'session_lifetime' => 7200, // 2 hours

    // 登录安全
    'login_max_attempts' => 5,       // 最大尝试次数
    'login_lockout_minutes' => 15,   // 锁定分钟数
];
