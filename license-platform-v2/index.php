<?php
/**
 * OpenClaw License Platform - 入口文件
 * 纯 PHP 路由分发，无需 Laravel/Composer/Artisan
 */

// 加载配置
$_CONFIG = null;
function getConfig(): array
{
    global $_CONFIG;
    if ($_CONFIG === null) {
        $_CONFIG = require __DIR__ . '/config.php';
        $envFile = __DIR__ . '/.env';
        if (file_exists($envFile)) {
            foreach (file($envFile, FILE_IGNORE_NEW_LINES | FILE_SKIP_EMPTY_LINES) as $line) {
                $trimmed = trim($line);
                if (strpos($trimmed, '#') === 0) continue;
                if (strpos($line, '=') === false) continue;
                putenv($trimmed);
            }
        }
        $_CONFIG = require __DIR__ . '/config.php';
    }
    return $_CONFIG;
}

// 自动加载类
spl_autoload_register(function (string $class) {
    $file = __DIR__ . '/src/' . $class . '.php';
    if (file_exists($file)) {
        require $file;
    }
});

// 错误处理
$debug = filter_var(getenv('APP_DEBUG') ?: false, FILTER_VALIDATE_BOOLEAN);
if ($debug) {
    error_reporting(E_ALL);
    ini_set('display_errors', 1);
} else {
    error_reporting(0);
    ini_set('display_errors', 0);
}

// 获取请求路径
$requestUri = parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH);
$requestMethod = $_SERVER['REQUEST_METHOD'];

if ($requestUri !== '/' && substr($requestUri, -1) === '/') {
    $requestUri = rtrim($requestUri, '/');
}

try {
    matchRoute($requestUri, $requestMethod);
} catch (Throwable $e) {
    error_log('License Platform Error: ' . $e->getMessage() . "\n" . $e->getTraceAsString());
    if ($debug) {
        http_response_code(500);
        echo '<h1>Server Error</h1><pre>' . htmlspecialchars($e->getMessage() . "\n" . $e->getTraceAsString()) . '</pre>';
    } else {
        http_response_code(500);
        echo json_encode(['error' => 'Internal Server Error']);
    }
}

function matchRoute(string $uri, string $method): void
{
    // API 路由 (兼容 PHP 7.4+)
    if (strpos($uri, '/api/v1/') === 0) {
        header('Content-Type: application/json; charset=utf-8');
        header('X-Content-Type-Options: nosniff');

        $path = substr($uri, strlen('/api/v1/'));

        if ($method === 'POST') {
            switch ($path) {
                case 'activate': ApiController::activate(); break;
                case 'verify': ApiController::verify(); break;
                case 'heartbeat': ApiController::heartbeat(); break;
                case 'ping': ApiController::ping(); break;
                default: jsonError('Not Found', 404); break;
            }
        } elseif ($method === 'GET') {
            switch ($path) {
                case 'public-key': ApiController::publicKey(); break;
                default: jsonError('Not Found', 404); break;
            }
        } else {
            jsonError('Method Not Allowed', 405);
        }
        return;
    }

    // 健康检查
    if ($uri === '/up') {
        header('Content-Type: application/json');
        echo json_encode(['status' => 'ok']);
        return;
    }

    // Web 路由
    switch ($uri) {
        case '/':
        case '/login':
            if ($method === 'GET') AdminController::showLogin();
            elseif ($method === 'POST') AdminController::handleLogin();
            break;

        case '/logout':
            if ($method === 'POST') AdminController::handleLogout();
            break;

        case '/admin/dashboard':
            AdminController::dashboard();
            break;

        case '/admin/licenses':
            AdminController::licenses();
            break;

        case '/admin/licenses/create':
            if ($method === 'GET') AdminController::showCreateLicense();
            elseif ($method === 'POST') AdminController::createLicense();
            break;

        case '/admin/licenses/revoke':
            if ($method === 'POST') AdminController::revokeLicense();
            break;

        case '/admin/licenses/renew':
            if ($method === 'POST') AdminController::renewLicense();
            break;

        case '/admin/devices':
            AdminController::devices();
            break;

        case '/admin/devices/delete':
            if ($method === 'POST') AdminController::deleteDevice();
            break;

        case '/admin/stats':
            AdminController::stats();
            break;

        case '/admin/pricing':
            AdminController::pricing();
            break;

        case '/admin/settings':
            if ($method === 'GET') AdminController::showSettings();
            elseif ($method === 'POST') AdminController::saveSettings();
            break;

        case '/admin/change-password':
            if ($method === 'GET') AdminController::showChangePassword();
            elseif ($method === 'POST') AdminController::handleChangePassword();
            break;

        default:
            http_response_code(404);
            echo '<h1>404 - Page Not Found</h1><a href="/">返回首页</a>';
            break;
    }
}

function jsonError(string $message, int $code): void
{
    http_response_code($code);
    echo json_encode(['error' => $message]);
    exit;
}
