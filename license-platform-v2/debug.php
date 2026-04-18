<?php
/**
 * 诊断脚本 - 排查 API 路由 404 问题
 * 访问: https://auth.bestu.top/debug.php
 * 排查完毕后务必删除此文件！
 */

header('Content-Type: text/plain; charset=utf-8');

echo "=== License Platform v2 诊断 ===\n\n";

// 1. 确认这是 v2 还是 v1 (Laravel)
echo "--- 版本检测 ---\n";
echo "index.php exists: " . (file_exists(__DIR__ . '/index.php') ? 'YES' : 'NO') . "\n";
echo "artisan exists (v1): " . (file_exists(__DIR__ . '/artisan') ? 'YES - v1 Laravel!' : 'NO') . "\n";
echo "config.php exists (v2): " . (file_exists(__DIR__ . '/config.php') ? 'YES - v2!' : 'NO') . "\n";
echo "src/ApiController.php (v2): " . (file_exists(__DIR__ . '/src/ApiController.php') ? 'YES - v2!' : 'NO') . "\n";
echo "vendor/autoload.php (v1): " . (file_exists(__DIR__ . '/vendor/autoload.php') ? 'YES - v1 Laravel!' : 'NO') . "\n";

// 2. PHP 版本和函数
echo "\n--- PHP 环境 ---\n";
echo "PHP Version: " . PHP_VERSION . "\n";
echo "str_starts_with: " . (function_exists('str_starts_with') ? 'YES (PHP 8.0+)' : 'NO - NEED PHP 8.0+') . "\n";

// 3. 请求信息
echo "\n--- 请求信息 ---\n";
echo "REQUEST_URI: " . ($_SERVER['REQUEST_URI'] ?? 'N/A') . "\n";
echo "SCRIPT_NAME: " . ($_SERVER['SCRIPT_NAME'] ?? 'N/A') . "\n";
echo "DOCUMENT_ROOT: " . ($_SERVER['DOCUMENT_ROOT'] ?? 'N/A') . "\n";

// 4. index.php 路由检查
echo "\n--- 路由检查 ---\n";
if (file_exists(__DIR__ . '/index.php')) {
    $content = file_get_contents(__DIR__ . '/index.php');
    echo "index.php has 'public-key' route: " . (str_contains($content, 'public-key') ? 'YES' : 'NO - MISSING!') . "\n";
    echo "index.php has 'ApiController::publicKey': " . (str_contains($content, 'publicKey') ? 'YES' : 'NO - MISSING!') . "\n";
    echo "index.php has '/api/v1/' prefix: " . (str_contains($content, '/api/v1/') ? 'YES' : 'NO - MISSING!') . "\n";
}

// 5. ApiController 检查
echo "\n--- ApiController 检查 ---\n";
if (file_exists(__DIR__ . '/src/ApiController.php')) {
    $content = file_get_contents(__DIR__ . '/src/ApiController.php');
    echo "publicKey() method: " . (str_contains($content, 'function publicKey') ? 'YES' : 'NO - MISSING!') . "\n";
} else {
    echo "ApiController.php NOT FOUND!\n";
}

// 6. 密钥文件 - 详细检查
echo "\n--- RSA 密钥 ---\n";
$keysDir = __DIR__ . '/keys';
echo "keys dir: $keysDir\n";
echo "keys dir exists: " . (is_dir($keysDir) ? 'YES' : 'NO') . "\n";
echo "keys dir readable: " . (is_readable($keysDir) ? 'YES' : 'NO') . "\n";

$pubKey = $keysDir . '/public.pem';
$privKey = $keysDir . '/private.pem';
echo "public.pem path: $pubKey\n";
echo "public.pem file_exists: " . (file_exists($pubKey) ? 'YES' : 'NO') . "\n";
echo "public.pem is_readable: " . (is_readable($pubKey) ? 'YES' : 'NO - PERMISSION ISSUE!') . "\n";
echo "public.pem filesize: " . @filesize($pubKey) . "\n";

// 用 posix 检查运行用户
echo "\n--- 运行用户 ---\n";
if (function_exists('posix_getpwuid')) {
    $procUser = posix_getpwuid(posix_geteuid());
    echo "PHP process user: " . ($procUser['name'] ?? 'unknown') . "\n";
}
echo "Whoami: " . exec('whoami 2>/dev/null') . "\n";

// 列出 keys 目录内容
echo "\n--- keys 目录内容 ---\n";
if (is_dir($keysDir)) {
    foreach (scandir($keysDir) as $f) {
        if ($f === '.' || $f === '..') continue;
        $fp = $keysDir . '/' . $f;
        $perms = substr(sprintf('%o', @fileperms($fp)), -4);
        $owner = function_exists('posix_getpwuid') ? (posix_getpwuid(@fileowner($fp))['name'] ?? '?') : @fileowner($fp);
        echo "$f | perms: $perms | owner: $owner | size: " . @filesize($fp) . "\n";
    }
} else {
    echo "keys 目录不存在!\n";
}

// 直接尝试读取公钥
echo "\n--- 读取公钥测试 ---\n";
$content = @file_get_contents($pubKey);
if ($content === false) {
    echo "file_get_contents FAILED!\n";
    $error = error_get_last();
    echo "Error: " . ($error['message'] ?? 'unknown') . "\n";
} else {
    echo "SUCCESS! Length: " . strlen($content) . "\n";
    echo "First line: " . trim(explode("\n", $content)[0]) . "\n";
}

// 7. 模拟路由
echo "\n--- 模拟路由 ---\n";
$testUri = '/api/v1/public-key';
echo "Test: $testUri\n";
if (function_exists('str_starts_with')) {
    echo "str_starts_with('/api/v1/'): " . (str_starts_with($testUri, '/api/v1/') ? 'MATCH' : 'NO MATCH') . "\n";
    $path = substr($testUri, strlen('/api/v1/'));
    echo "Extracted path: '$path'\n";
}

echo "\n=== 诊断结束 ===\n";
echo "\n!!! 排查完毕后请删除此文件 !!!\n";
