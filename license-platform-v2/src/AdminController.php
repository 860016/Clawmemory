<?php
/**
 * OpenClaw License Platform - 管理后台控制器
 * 处理所有 /admin/* 请求
 */

class AdminController
{
    // ========== 登录相关 ==========

    public static function showLogin(): void
    {
        if (Session::isLoggedIn()) {
            header('Location: /admin/dashboard');
            exit;
        }
        self::render('login', ['error' => '', 'locked_until' => null]);
    }

    public static function handleLogin(): void
    {
        $email = trim($_POST['email'] ?? '');
        $password = $_POST['password'] ?? '';
        $config = getConfig();

        // 检查锁定
        $lockedUntil = self::checkLockout($email);
        if ($lockedUntil) {
            self::render('login', ['error' => "登录尝试过多，请 {$lockedUntil} 分钟后再试", 'locked_until' => $lockedUntil]);
            return;
        }

        if (empty($email) || empty($password)) {
            self::render('login', ['error' => '请输入邮箱和密码', 'locked_until' => null]);
            return;
        }

        $user = Database::fetch("SELECT * FROM users WHERE email = ?", [$email]);

        // 记录登录日志
        self::logLoginAttempt($email, $user && password_verify($password, $user['password']));

        if (!$user || !password_verify($password, $user['password'])) {
            self::render('login', ['error' => '邮箱或密码错误', 'locked_until' => null]);
            return;
        }

        Session::login((int) $user['id'], $user['email']);
        header('Location: /admin/dashboard');
        exit;
    }

    public static function handleLogout(): void
    {
        Session::destroy();
        header('Location: /login');
        exit;
    }

    // ========== 修改密码 ==========

    public static function showChangePassword(): void
    {
        Session::requireAuth();
        self::render('change_password', ['error' => '', 'success' => '']);
    }

    public static function handleChangePassword(): void
    {
        Session::requireAuth();

        $oldPassword = $_POST['old_password'] ?? '';
        $newPassword = $_POST['new_password'] ?? '';
        $confirmPassword = $_POST['confirm_password'] ?? '';

        if (empty($oldPassword) || empty($newPassword)) {
            self::render('change_password', ['error' => '请填写所有字段', 'success' => '']);
            return;
        }

        $user = Database::fetch("SELECT * FROM users WHERE id = ?", [Session::get('user_id')]);

        if (!password_verify($oldPassword, $user['password'])) {
            self::render('change_password', ['error' => '旧密码错误', 'success' => '']);
            return;
        }

        if (strlen($newPassword) < 6) {
            self::render('change_password', ['error' => '新密码至少6位', 'success' => '']);
            return;
        }

        if ($newPassword !== $confirmPassword) {
            self::render('change_password', ['error' => '两次密码不一致', 'success' => '']);
            return;
        }

        $hash = password_hash($newPassword, PASSWORD_DEFAULT);
        Database::execute("UPDATE users SET password = ?, updated_at = NOW() WHERE id = ?", [$hash, $user['id']]);

        self::render('change_password', ['error' => '', 'success' => '密码修改成功']);
    }

    // ========== 仪表盘 ==========

    public static function dashboard(): void
    {
        Session::requireAuth();

        $totalLicenses = (int) Database::fetchColumn("SELECT COUNT(*) FROM licenses");
        $activeLicenses = (int) Database::fetchColumn("SELECT COUNT(*) FROM licenses WHERE status = 'active'");
        $expiredLicenses = (int) Database::fetchColumn("SELECT COUNT(*) FROM licenses WHERE status = 'expired'");
        $revokedLicenses = (int) Database::fetchColumn("SELECT COUNT(*) FROM licenses WHERE status = 'revoked'");
        $totalDevices = (int) Database::fetchColumn("SELECT COUNT(*) FROM devices");
        $todayActive = (int) Database::fetchColumn(
            "SELECT COUNT(DISTINCT license_id) FROM usage_stats WHERE stat_date = CURDATE()"
        );
        $totalInstalls = (int) Database::fetchColumn("SELECT COUNT(*) FROM install_pings");
        $monthlyInstalls = (int) Database::fetchColumn(
            "SELECT COUNT(*) FROM install_pings WHERE created_at >= DATE_SUB(NOW(), INTERVAL 1 MONTH)"
        );
        $monthlyRevenue = Database::fetchColumn(
            "SELECT COALESCE(SUM(price_paid), 0) FROM licenses WHERE created_at >= DATE_SUB(NOW(), INTERVAL 1 MONTH)"
        ) ?: 0;

        // 最近登录警报
        $recentFailedLogins = Database::fetchAll(
            "SELECT * FROM login_logs WHERE success = 0 AND created_at >= DATE_SUB(NOW(), INTERVAL 24 HOUR) ORDER BY created_at DESC LIMIT 10"
        );

        // 即将到期的授权（7天内）
        $expiringSoon = Database::fetchAll(
            "SELECT * FROM licenses WHERE status = 'active' AND expires_at IS NOT NULL AND expires_at <= DATE_ADD(NOW(), INTERVAL 7 DAY) AND expires_at > NOW() ORDER BY expires_at ASC LIMIT 5"
        );

        self::render('dashboard', [
            'totalLicenses' => $totalLicenses,
            'activeLicenses' => $activeLicenses,
            'expiredLicenses' => $expiredLicenses,
            'revokedLicenses' => $revokedLicenses,
            'totalDevices' => $totalDevices,
            'todayActive' => $todayActive,
            'totalInstalls' => $totalInstalls,
            'monthlyInstalls' => $monthlyInstalls,
            'monthlyRevenue' => $monthlyRevenue,
            'recentFailedLogins' => $recentFailedLogins,
            'expiringSoon' => $expiringSoon,
        ]);
    }

    // ========== 授权码管理 ==========

    public static function licenses(): void
    {
        Session::requireAuth();

        $page = max(1, (int) ($_GET['page'] ?? 1));
        $perPage = getConfig()['per_page'];
        $offset = ($page - 1) * $perPage;

        $total = (int) Database::fetchColumn("SELECT COUNT(*) FROM licenses");
        $licenses = Database::fetchAll(
            "SELECT l.*, (SELECT COUNT(*) FROM devices WHERE license_id = l.id) AS device_count 
             FROM licenses l ORDER BY l.created_at DESC LIMIT ? OFFSET ?",
            [$perPage, $offset]
        );

        $totalPages = (int) ceil($total / $perPage);

        self::render('licenses', [
            'licenses' => $licenses,
            'page' => $page,
            'totalPages' => $totalPages,
            'total' => $total,
        ]);
    }

    public static function showCreateLicense(): void
    {
        Session::requireAuth();
        self::render('license_create', [
            'tiers' => Settings::getPricingTiers(),
            'error' => '',
        ]);
    }

    public static function createLicense(): void
    {
        Session::requireAuth();

        $tier = $_POST['tier'] ?? 'pro';
        $tiers = Settings::getPricingTiers();

        if (!isset($tiers[$tier])) {
            self::render('license_create', ['tiers' => $tiers, 'error' => '无效的套餐类型']);
            return;
        }

        $licenseKey = LicenseGenerator::generateKey();
        $features = json_encode($tiers[$tier]['features']);
        $maxDevices = (int) ($_POST['max_devices'] ?? $tiers[$tier]['max_devices']);
        $expiresAt = !empty($_POST['expires_at']) ? $_POST['expires_at'] : null;
        $customerName = trim($_POST['customer_name'] ?? '');
        $customerEmail = trim($_POST['customer_email'] ?? '');
        $notes = trim($_POST['notes'] ?? '');

        try {
            Database::insert(
                "INSERT INTO licenses (license_key, tier, features, max_devices, status, expires_at, customer_name, customer_email, notes, created_at, updated_at) 
                 VALUES (?, ?, ?, ?, 'active', ?, ?, ?, ?, NOW(), NOW())",
                [$licenseKey, $tier, $features, $maxDevices, $expiresAt, $customerName, $customerEmail, $notes]
            );

            header('Location: /admin/licenses?created=' . urlencode($licenseKey));
            exit;
        } catch (Exception $e) {
            self::render('license_create', ['tiers' => $tiers, 'error' => '创建失败：' . $e->getMessage()]);
        }
    }

    public static function revokeLicense(): void
    {
        Session::requireAuth();
        $id = (int) ($_POST['id'] ?? 0);
        if ($id > 0) {
            Database::execute("UPDATE licenses SET status = 'revoked', updated_at = NOW() WHERE id = ?", [$id]);
        }
        header('Location: /admin/licenses');
        exit;
    }

    public static function renewLicense(): void
    {
        Session::requireAuth();
        $id = (int) ($_POST['id'] ?? 0);
        if ($id > 0) {
            Database::execute(
                "UPDATE licenses SET status = 'active', expires_at = DATE_ADD(NOW(), INTERVAL 1 YEAR), updated_at = NOW() WHERE id = ?",
                [$id]
            );
        }
        header('Location: /admin/licenses');
        exit;
    }

    // ========== 设备管理 ==========

    public static function devices(): void
    {
        Session::requireAuth();

        $page = max(1, (int) ($_GET['page'] ?? 1));
        $perPage = getConfig()['per_page'];
        $offset = ($page - 1) * $perPage;
        $search = trim($_GET['search'] ?? '');

        $where = '';
        $params = [];

        if ($search) {
            $where = "WHERE d.fingerprint_hash LIKE ? OR d.device_name LIKE ?";
            $params = ["%$search%", "%$search%"];
        }

        $total = (int) Database::fetchColumn("SELECT COUNT(*) FROM devices d $where", $params);

        $allParams = array_merge($params, [$perPage, $offset]);
        $devices = Database::fetchAll(
            "SELECT d.*, l.license_key, l.tier FROM devices d 
             JOIN licenses l ON d.license_id = l.id 
             $where ORDER BY d.last_active_at DESC LIMIT ? OFFSET ?",
            $allParams
        );

        $totalPages = (int) ceil($total / $perPage);

        self::render('devices', [
            'devices' => $devices,
            'page' => $page,
            'totalPages' => $totalPages,
            'total' => $total,
            'search' => $search,
        ]);
    }

    public static function deleteDevice(): void
    {
        Session::requireAuth();
        $id = (int) ($_POST['id'] ?? 0);
        if ($id > 0) {
            Database::execute("DELETE FROM devices WHERE id = ?", [$id]);
        }
        header('Location: /admin/devices');
        exit;
    }

    // ========== 统计 ==========

    public static function stats(): void
    {
        Session::requireAuth();

        $dauTrend = Database::fetchAll(
            "SELECT stat_date, COUNT(DISTINCT license_id) AS dau 
             FROM usage_stats WHERE stat_date >= DATE_SUB(CURDATE(), INTERVAL 30 DAY) 
             GROUP BY stat_date ORDER BY stat_date"
        );

        $versionDist = Database::fetchAll(
            "SELECT app_version, COUNT(*) AS count FROM devices WHERE app_version IS NOT NULL GROUP BY app_version ORDER BY count DESC"
        );

        $tierDist = Database::fetchAll(
            "SELECT tier, COUNT(*) AS count FROM licenses WHERE status = 'active' GROUP BY tier"
        );

        self::render('stats', [
            'dauTrend' => $dauTrend,
            'versionDist' => $versionDist,
            'tierDist' => $tierDist,
        ]);
    }

    // ========== 价格方案（可编辑） ==========

    public static function pricing(): void
    {
        Session::requireAuth();
        $tiers = Settings::getPricingTiers();
        self::render('pricing', ['tiers' => $tiers]);
    }

    public static function showSettings(): void
    {
        Session::requireAuth();

        $tiers = Settings::getPricingTiers();
        $paymentMethod = Settings::getPaymentMethod();
        $allSettings = Settings::getAll();

        // 登录警报
        $loginLogs = Database::fetchAll(
            "SELECT * FROM login_logs ORDER BY created_at DESC LIMIT 50"
        );

        self::render('settings', [
            'tiers' => $tiers,
            'paymentMethod' => $paymentMethod,
            'allSettings' => $allSettings,
            'loginLogs' => $loginLogs,
            'success' => '',
            'error' => '',
        ]);
    }

    public static function saveSettings(): void
    {
        Session::requireAuth();

        $section = $_POST['section'] ?? '';

        try {
            switch ($section) {
                case 'pricing':
                    $tiers = [];
                    $tierKeys = $_POST['tier_key'] ?? [];
                    $tierNames = $_POST['tier_name'] ?? [];
                    $tierPrices = $_POST['tier_price'] ?? [];
                    $tierMaxDevices = $_POST['tier_max_devices'] ?? [];
                    $tierFeatures = $_POST['tier_features'] ?? [];

                    for ($i = 0; $i < count($tierKeys); $i++) {
                        $key = trim($tierKeys[$i]);
                        if (empty($key)) continue;
                        $tiers[$key] = [
                            'name' => trim($tierNames[$i] ?? ucfirst($key)),
                            'price' => (float) ($tierPrices[$i] ?? 0),
                            'max_devices' => (int) ($tierMaxDevices[$i] ?? 1),
                            'features' => array_filter(array_map('trim', explode(',', $tierFeatures[$i] ?? ''))),
                        ];
                    }
                    Settings::set('pricing_tiers', $tiers, '价格方案配置');
                    break;

                case 'payment':
                    Settings::set('payment_method', $_POST['payment_method'] ?? 'manual', '支付方式');
                    Settings::set('alipay_app_id', $_POST['alipay_app_id'] ?? '', '支付宝App ID');
                    Settings::set('alipay_private_key', $_POST['alipay_private_key'] ?? '', '支付宝应用私钥');
                    Settings::set('alipay_public_key', $_POST['alipay_public_key'] ?? '', '支付宝公钥');
                    Settings::set('wechat_mch_id', $_POST['wechat_mch_id'] ?? '', '微信商户号');
                    Settings::set('wechat_api_key', $_POST['wechat_api_key'] ?? '', '微信API密钥');
                    Settings::set('wechat_app_id', $_POST['wechat_app_id'] ?? '', '微信App ID');
                    break;

                case 'email':
                    Settings::set('smtp_host', $_POST['smtp_host'] ?? '', 'SMTP地址');
                    Settings::set('smtp_port', $_POST['smtp_port'] ?? '465', 'SMTP端口');
                    Settings::set('smtp_user', $_POST['smtp_user'] ?? '', 'SMTP用户名');
                    Settings::set('smtp_pass', $_POST['smtp_pass'] ?? '', 'SMTP密码');
                    Settings::set('smtp_from', $_POST['smtp_from'] ?? '', '发件人地址');
                    Settings::set('expiry_notify_days', $_POST['expiry_notify_days'] ?? '7', '到期提醒天数');
                    break;
            }

            Settings::clearCache();

            // 重新加载数据
            $tiers = Settings::getPricingTiers();
            $paymentMethod = Settings::getPaymentMethod();
            $allSettings = Settings::getAll();
            $loginLogs = Database::fetchAll("SELECT * FROM login_logs ORDER BY created_at DESC LIMIT 50");

            self::render('settings', [
                'tiers' => $tiers,
                'paymentMethod' => $paymentMethod,
                'allSettings' => $allSettings,
                'loginLogs' => $loginLogs,
                'success' => '保存成功',
                'error' => '',
            ]);
        } catch (Exception $e) {
            $tiers = Settings::getPricingTiers();
            $paymentMethod = Settings::getPaymentMethod();
            $allSettings = Settings::getAll();
            $loginLogs = Database::fetchAll("SELECT * FROM login_logs ORDER BY created_at DESC LIMIT 50");

            self::render('settings', [
                'tiers' => $tiers,
                'paymentMethod' => $paymentMethod,
                'allSettings' => $allSettings,
                'loginLogs' => $loginLogs,
                'success' => '',
                'error' => '保存失败：' . $e->getMessage(),
            ]);
        }
    }

    // ========== Helper ==========

    private static function checkLockout(string $email): ?int
    {
        $config = getConfig();
        $maxAttempts = $config['login_max_attempts'] ?? 5;
        $lockoutMinutes = $config['login_lockout_minutes'] ?? 15;

        $recent = Database::fetch(
            "SELECT COUNT(*) AS cnt FROM login_logs WHERE email = ? AND success = 0 AND created_at >= DATE_SUB(NOW(), INTERVAL ? MINUTE)",
            [$email, $lockoutMinutes]
        );

        if ((int) ($recent['cnt'] ?? 0) >= $maxAttempts) {
            $lastAttempt = Database::fetch(
                "SELECT created_at FROM login_logs WHERE email = ? AND success = 0 ORDER BY created_at DESC LIMIT 1",
                [$email]
            );
            if ($lastAttempt) {
                $remaining = $lockoutMinutes - (time() - strtotime($lastAttempt['created_at'])) / 60;
                return max(1, (int) ceil($remaining));
            }
        }

        return null;
    }

    private static function logLoginAttempt(string $email, bool $success): void
    {
        Database::insert(
            "INSERT INTO login_logs (email, ip_address, user_agent, success, created_at) VALUES (?, ?, ?, ?, NOW())",
            [
                $email,
                $_SERVER['HTTP_X_FORWARDED_FOR'] ?? $_SERVER['HTTP_X_REAL_IP'] ?? $_SERVER['REMOTE_ADDR'] ?? '0.0.0.0',
                substr($_SERVER['HTTP_USER_AGENT'] ?? '', 0, 500),
                $success ? 1 : 0,
            ]
        );
    }

    private static function render(string $view, array $data = []): void
    {
        extract($data);
        $config = getConfig();
        $isLoggedIn = Session::isLoggedIn();
        $currentUserEmail = Session::get('user_email', '');

        ob_start();
        include __DIR__ . '/../views/' . $view . '.php';
        $content = ob_get_clean();

        include __DIR__ . '/../views/layout.php';
    }
}
