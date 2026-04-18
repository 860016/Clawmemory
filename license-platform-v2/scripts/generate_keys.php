#!/usr/bin/env php
<?php
/**
 * RSA 密钥对生成脚本
 * 用法：php generate_keys.php
 */

$keysDir = __DIR__ . '/../keys';

if (!is_dir($keysDir)) {
    mkdir($keysDir, 0755, true);
}

$privateKeyPath = $keysDir . '/private.pem';
$publicKeyPath = $keysDir . '/public.pem';

if (file_exists($privateKeyPath)) {
    echo "密钥已存在，是否覆盖？(y/N): ";
    $handle = fopen('php://stdin', 'r');
    $line = trim(fgets($handle));
    fclose($handle);
    if (strtolower($line) !== 'y') {
        echo "已取消\n";
        exit(0);
    }
}

echo "正在生成 RSA-2048 密钥对...\n";

$config = [
    'digest_alg' => 'sha256',
    'private_key_bits' => 2048,
    'private_key_type' => OPENSSL_KEYTYPE_RSA,
];

$keypair = openssl_pkey_new($config);
if (!$keypair) {
    die("生成失败: " . openssl_error_string() . "\n");
}

// 导出私钥
openssl_pkey_export($keypair, $privateKeyPem);
file_put_contents($privateKeyPath, $privateKeyPem);
chmod($privateKeyPath, 0600);

// 导出公钥
$details = openssl_pkey_get_details($keypair);
file_put_contents($publicKeyPath, $details['key']);
chmod($publicKeyPath, 0644);

echo "密钥生成成功:\n";
echo "  私钥: {$privateKeyPath}\n";
echo "  公钥: {$publicKeyPath}\n";
echo "\n重要提示：请勿将 private.pem 提交到版本控制或泄露给他人！\n";
