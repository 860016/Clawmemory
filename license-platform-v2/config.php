<?php
/**
 * ClawMemory License Platform - 配置文件
 * v3.0 - 匹配 Rust PyO3 核心引擎 + Pro 功能
 */

return [
    // 应用配置
    'app_name' => getenv('APP_NAME') ?: 'ClawMemory License Server',
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

    // v3.0 功能定义（功能键 => 中文标签）
    'feature_labels' => [
        // Memory Decay
        'auto_decay' => '自动记忆衰减',
        'decay_report' => '衰减报告',
        'prune_suggest' => '清理建议',
        'reinforce' => '记忆强化',
        // Conflict Resolver
        'conflict_scan' => '矛盾扫描',
        'conflict_merge' => '自动合并',
        // Token Router
        'smart_router' => '智能模型路由',
        'token_stats' => 'Token 统计',
        // AI & Graph
        'ai_extract' => 'AI 智能提取',
        'auto_graph' => '自动知识图谱',
        'unlimited_graph' => '无限图谱节点',
        // Wiki
        'wiki' => 'Wiki 知识库',
        // Enterprise only
        'api_access' => 'API 完整访问',
        'sso' => 'SSO 单点登录',
        'audit_log' => '审计日志',
        'time_travel' => '时间回溯',
        'offline_mode' => '离线模式',
    ],

    // 默认授权码配置（数据库 settings 表优先）
    'license' => [
        'tiers' => [
            'oss' => [
                'name' => 'OSS 免费版',
                'max_devices' => 1,
                'price' => 0,
                'features' => [],
            ],
            'pro' => [
                'name' => 'Pro 专业版',
                'max_devices' => 3,
                'price' => 29,
                'features' => ['ai_extract', 'auto_graph', 'unlimited_graph', 'auto_decay', 'decay_report', 'prune_suggest', 'reinforce', 'conflict_scan', 'conflict_merge', 'smart_router', 'token_stats', 'wiki'],
            ],
            'enterprise' => [
                'name' => 'Enterprise 企业版',
                'max_devices' => 10,
                'price' => 99,
                'features' => ['ai_extract', 'auto_graph', 'unlimited_graph', 'auto_decay', 'decay_report', 'prune_suggest', 'reinforce', 'conflict_scan', 'conflict_merge', 'smart_router', 'token_stats', 'wiki', 'api_access', 'sso', 'audit_log', 'time_travel', 'offline_mode'],
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
