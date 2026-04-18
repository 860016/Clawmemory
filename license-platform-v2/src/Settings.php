<?php
/**
 * OpenClaw License Platform - 设置管理
 * 从数据库 settings 表读取/写入配置
 */

class Settings
{
    private static ?array $cache = null;

    /**
     * 获取设置值
     */
    public static function get(string $key, mixed $default = null): mixed
    {
        $all = self::getAll();
        return $all[$key] ?? $default;
    }

    /**
     * 设置值
     */
    public static function set(string $key, mixed $value, string $description = null): void
    {
        $jsonValue = is_array($value) ? json_encode($value, JSON_UNESCAPED_UNICODE) : (string) $value;

        $existing = Database::fetch("SELECT id FROM settings WHERE `key` = ?", [$key]);
        if ($existing) {
            Database::execute("UPDATE settings SET `value` = ?, updated_at = NOW() WHERE `key` = ?", [$jsonValue, $key]);
        } else {
            Database::insert(
                "INSERT INTO settings (`key`, `value`, `description`, updated_at) VALUES (?, ?, ?, NOW())",
                [$key, $jsonValue, $description]
            );
        }
        self::$cache = null; // 清缓存
    }

    /**
     * 获取所有设置
     */
    public static function getAll(): array
    {
        if (self::$cache === null) {
            $rows = Database::fetchAll("SELECT `key`, `value` FROM settings");
            self::$cache = [];
            foreach ($rows as $row) {
                $value = $row['value'];
                // 尝试解析 JSON
                $decoded = json_decode($value, true);
                if (json_last_error() === JSON_ERROR_NONE && is_array($decoded)) {
                    $value = $decoded;
                }
                self::$cache[$row['key']] = $value;
            }
        }
        return self::$cache;
    }

    /**
     * 获取价格方案（优先从数据库读取）
     */
    public static function getPricingTiers(): array
    {
        $tiers = self::get('pricing_tiers');
        if ($tiers && is_array($tiers)) {
            return $tiers;
        }
        // 降级到配置文件
        return getConfig()['license']['tiers'];
    }

    /**
     * 获取支付方式
     */
    public static function getPaymentMethod(): string
    {
        return self::get('payment_method', 'manual');
    }

    /**
     * 获取到期提醒天数
     */
    public static function getExpiryNotifyDays(): int
    {
        return (int) self::get('expiry_notify_days', 7);
    }

    /**
     * 清除缓存
     */
    public static function clearCache(): void
    {
        self::$cache = null;
    }
}
