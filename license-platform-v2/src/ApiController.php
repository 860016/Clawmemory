<?php
/**
 * OpenClaw License Platform - API 控制器
 * 处理所有 /api/v1/* 请求
 */

class ApiController
{
    /**
     * POST /api/v1/activate - 激活授权码
     */
    public static function activate(): void
    {
        $input = self::getInput();
        
        if (empty($input['license_key']) || empty($input['fingerprint'])) {
            self::error('Missing required fields: license_key, fingerprint', 400);
        }

        $licenseKey = trim($input['license_key']);
        $fingerprint = trim($input['fingerprint']);
        $version = $input['version'] ?? '';
        $deviceName = $input['device_name'] ?? 'Unknown';
        $os = $input['os'] ?? '';
        $customerEmail = trim($input['email'] ?? '');  // 用户绑定邮箱

        // 查找授权码
        $license = Database::fetch(
            "SELECT * FROM licenses WHERE license_key = ? AND status = 'active'",
            [$licenseKey]
        );

        if (!$license) {
            self::error('Invalid or revoked license');
        }

        // 检查过期
        if ($license['expires_at'] && strtotime($license['expires_at']) < time()) {
            Database::execute("UPDATE licenses SET status = 'expired' WHERE id = ?", [$license['id']]);
            self::error('License expired');
        }

        // 检查设备限制
        $deviceCount = (int) Database::fetchColumn(
            "SELECT COUNT(*) FROM devices WHERE license_id = ?",
            [$license['id']]
        );
        $maxDevices = (int) $license['max_devices'];

        $existingDevice = Database::fetch(
            "SELECT * FROM devices WHERE license_id = ? AND fingerprint_hash = ?",
            [$license['id'], $fingerprint]
        );

        if (!$existingDevice && $deviceCount >= $maxDevices) {
            self::error("Device limit reached ($deviceCount/$maxDevices). Please deactivate a device.");
        }

        // 注册/更新设备
        if (!$existingDevice) {
            Database::insert(
                "INSERT INTO devices (license_id, fingerprint_hash, device_name, os_info, app_version, last_active_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, NOW(), NOW(), NOW())",
                [$license['id'], $fingerprint, $deviceName, $os, $version]
            );
            $deviceCount++;
        } else {
            Database::execute(
                "UPDATE devices SET last_active_at = NOW(), updated_at = NOW() WHERE id = ?",
                [$existingDevice['id']]
            );
        }

        // 生成签名数据
        $tier = $license['tier'];
        $features = $license['features'] ? json_decode($license['features'], true) : [];
        $data = json_encode([
            'tier' => $tier,
            'features' => $features ?: [],
            'expires_at' => $license['expires_at'],
        ]);

        $signature = RsaSigner::sign($data);

        // 记录激活
        Database::insert(
            "INSERT INTO activations (license_id, action, fingerprint, ip_address, created_at, updated_at) VALUES (?, 'activate', ?, ?, NOW(), NOW())",
            [$license['id'], $fingerprint, self::getClientIp()]
        );

        // 绑定用户邮箱（激活时传入，用于到期提醒）
        if (!empty($customerEmail) && empty($license['customer_email'])) {
            Database::execute(
                "UPDATE licenses SET customer_email = ?, updated_at = NOW() WHERE id = ? AND (customer_email IS NULL OR customer_email = '')",
                [$customerEmail, $license['id']]
            );
        }

        self::success([
            'valid' => true,
            'tier' => $tier,
            'features' => $features ?: [],
            'expires_at' => $license['expires_at'],
            'device_slot' => "$deviceCount/$maxDevices",
            'signature' => base64_encode(json_encode(['data' => $data, 'signature' => $signature])),
        ]);
    }

    /**
     * POST /api/v1/verify - 验证授权码
     */
    public static function verify(): void
    {
        $input = self::getInput();

        if (empty($input['license_key'])) {
            self::error('Missing required field: license_key', 400);
        }

        $license = Database::fetch(
            "SELECT * FROM licenses WHERE license_key = ? AND status = 'active'",
            [trim($input['license_key'])]
        );

        if (!$license) {
            self::error('Invalid license');
        }

        if ($license['expires_at'] && strtotime($license['expires_at']) < time()) {
            self::error('License expired');
        }

        $features = $license['features'] ? json_decode($license['features'], true) : [];

        self::success([
            'valid' => true,
            'tier' => $license['tier'],
            'features' => $features ?: [],
            'expires_at' => $license['expires_at'],
        ]);
    }

    /**
     * POST /api/v1/heartbeat - 心跳上报
     */
    public static function heartbeat(): void
    {
        $input = self::getInput();

        if (empty($input['license_key'])) {
            self::error('Missing required field: license_key', 400);
        }

        $license = Database::fetch(
            "SELECT id FROM licenses WHERE license_key = ?",
            [trim($input['license_key'])]
        );

        if (!$license) {
            self::success(['status' => 'ignored']);
        }

        // 记录使用统计
        $statDate = date('Y-m-d');
        $existing = Database::fetch(
            "SELECT id FROM usage_stats WHERE license_id = ? AND stat_date = ?",
            [$license['id'], $statDate]
        );

        if ($existing) {
            Database::execute(
                "UPDATE usage_stats SET memory_count = ?, active_users = ?, app_version = ?, os = ?, updated_at = NOW() WHERE id = ?",
                [
                    (int) ($input['memory_count'] ?? 0),
                    (int) ($input['active_users'] ?? 0),
                    $input['version'] ?? '',
                    $input['os'] ?? '',
                    $existing['id'],
                ]
            );
        } else {
            Database::insert(
                "INSERT INTO usage_stats (license_id, stat_date, memory_count, active_users, app_version, os, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())",
                [
                    $license['id'],
                    $statDate,
                    (int) ($input['memory_count'] ?? 0),
                    (int) ($input['active_users'] ?? 0),
                    $input['version'] ?? '',
                    $input['os'] ?? '',
                ]
            );
        }

        self::success(['status' => 'ok']);
    }

    /**
     * POST /api/v1/ping - 安装统计
     */
    public static function ping(): void
    {
        $input = self::getInput();

        if (empty($input['install_id'])) {
            self::error('Missing required field: install_id', 400);
        }

        Database::insert(
            "INSERT INTO install_pings (install_id, version, os, ip_address, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())",
            [
                $input['install_id'],
                $input['version'] ?? '',
                $input['os'] ?? '',
                self::getClientIp(),
            ]
        );

        self::success(['status' => 'ok']);
    }

    /**
     * GET /api/v1/public-key - 获取 RSA 公钥（供客户端安装时自动获取）
     */
    public static function publicKey(): void
    {
        $config = getConfig();
        $keyPath = $config['public_key_path'] ?? __DIR__ . '/../keys/public.pem';
        
        if (!file_exists($keyPath)) {
            self::error('Public key not found', 404);
        }
        
        $keyContent = file_get_contents($keyPath);
        header('Content-Type: text/plain; charset=utf-8');
        echo $keyContent;
        exit;
    }

    // ========== Helper Methods ==========

    private static function getInput(): array
    {
        $contentType = $_SERVER['CONTENT_TYPE'] ?? '';
        $raw = file_get_contents('php://input');

        if (strpos($contentType, 'application/json') !== false) {
            return json_decode($raw, true) ?: [];
        }

        return $_POST;
    }

    private static function success(array $data, int $code = 200): void
    {
        http_response_code($code);
        header('Content-Type: application/json; charset=utf-8');
        echo json_encode($data, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
        exit;
    }

    private static function error(string $message, int $code = 400): void
    {
        http_response_code($code);
        header('Content-Type: application/json; charset=utf-8');
        echo json_encode(['valid' => false, 'message' => $message], JSON_UNESCAPED_UNICODE);
        exit;
    }

    private static function getClientIp(): string
    {
        return $_SERVER['HTTP_X_FORWARDED_FOR'] 
            ?? $_SERVER['HTTP_X_REAL_IP'] 
            ?? $_SERVER['REMOTE_ADDR'] 
            ?? '0.0.0.0';
    }
}
