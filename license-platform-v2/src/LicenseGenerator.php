<?php
/**
 * OpenClaw License Platform - 授权码生成器
 */

class LicenseGenerator
{
    private const CHARS = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';

    /**
     * 生成授权码，格式：OC-XXXX-XXXX-XXXX
     */
    public static function generateKey(): string
    {
        $segments = [];
        for ($i = 0; $i < 3; $i++) {
            $segment = '';
            for ($j = 0; $j < 4; $j++) {
                $segment .= self::CHARS[random_int(0, strlen(self::CHARS) - 1)];
            }
            $segments[] = $segment;
        }
        return 'OC-' . implode('-', $segments);
    }

    /**
     * 获取对应 tier 的功能列表
     */
    public static function getFeaturesForTier(string $tier): array
    {
        $config = getConfig();
        $tiers = $config['license']['tiers'];
        return $tiers[$tier]['features'] ?? [];
    }
}
