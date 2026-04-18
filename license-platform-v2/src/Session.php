<?php
/**
 * OpenClaw License Platform - Session 管理
 * 纯 PHP Session 封装
 */

class Session
{
    private static bool $started = false;

    public static function start(): void
    {
        if (self::$started) {
            return;
        }

        $config = getConfig();
        ini_set('session.cookie_httponly', 1);
        ini_set('session.cookie_samesite', 'Lax');
        ini_set('session.gc_maxlifetime', $config['session_lifetime']);
        session_name('oc_license');
        session_start();
        self::$started = true;
    }

    public static function set(string $key, mixed $value): void
    {
        self::start();
        $_SESSION[$key] = $value;
    }

    public static function get(string $key, mixed $default = null): mixed
    {
        self::start();
        return $_SESSION[$key] ?? $default;
    }

    public static function has(string $key): bool
    {
        self::start();
        return isset($_SESSION[$key]);
    }

    public static function remove(string $key): void
    {
        self::start();
        unset($_SESSION[$key]);
    }

    public static function destroy(): void
    {
        self::start();
        $_SESSION = [];
        if (ini_get('session.use_cookies')) {
            $params = session_get_cookie_params();
            setcookie(session_name(), '', time() - 42000,
                $params['path'], $params['domain'],
                $params['secure'], $params['httponly']
            );
        }
        session_destroy();
        self::$started = false;
    }

    public static function isLoggedIn(): bool
    {
        return self::has('user_id');
    }

    public static function requireAuth(): void
    {
        if (!self::isLoggedIn()) {
            header('Location: /login');
            exit;
        }
    }

    public static function login(int $userId, string $email): void
    {
        session_regenerate_id(true);
        self::set('user_id', $userId);
        self::set('user_email', $email);
        self::set('login_time', time());
    }
}
