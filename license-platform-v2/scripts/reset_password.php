#!/usr/bin/env php
<?php
/**
 * 重置管理员密码
 * 用法：php scripts/reset_password.php [新密码]
 * 默认重置为 admin123
 */

$newPassword = $argv[1] ?? 'admin123';

$envFile = __DIR__ . '/../.env';
if (file_exists($envFile)) {
    foreach (file($envFile, FILE_IGNORE_NEW_LINES | FILE_SKIP_EMPTY_LINES) as $line) {
        if (strpos(trim($line), '#') === 0 || strpos($line, '=') === false) continue;
        putenv(trim($line));
    }
}

$host = getenv('DB_HOST') ?: '127.0.0.1';
$port = getenv('DB_PORT') ?: '3306';
$dbname = getenv('DB_DATABASE') ?: 'openclaw_license';
$user = getenv('DB_USERNAME') ?: 'root';
$pass = getenv('DB_PASSWORD') ?: '';

try {
    $pdo = new PDO("mysql:host=$host;port=$port;dbname=$dbname;charset=utf8mb4", $user, $pass, [
        PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
    ]);

    $hash = password_hash($newPassword, PASSWORD_DEFAULT);
    $stmt = $pdo->prepare("UPDATE users SET password = ? WHERE email = 'admin@openclaw.ai'");
    $stmt->execute([$hash]);

    if ($stmt->rowCount() > 0) {
        echo "密码已重置为: $newPassword\n";
    } else {
        echo "未找到 admin@openclaw.ai 用户，尝试更新所有用户...\n";
        $stmt = $pdo->prepare("UPDATE users SET password = ? LIMIT 1");
        $stmt->execute([$hash]);
        echo "已重置第一个用户的密码为: $newPassword\n";
    }

    // 验证
    $row = $pdo->query("SELECT password FROM users WHERE email = 'admin@openclaw.ai'")->fetch();
    if ($row && password_verify($newPassword, $row['password'])) {
        echo "验证通过 ✓\n";
    } else {
        echo "验证失败 ✗\n";
    }
} catch (Exception $e) {
    echo "错误: " . $e->getMessage() . "\n";
    exit(1);
}
