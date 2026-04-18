<?php
/**
 * OpenClaw License Platform - RSA 签名服务
 * 使用 RSA-2048 + SHA256 对授权数据进行签名
 */

class RsaSigner
{
    /**
     * 用私钥对数据进行签名
     */
    public static function sign(string $data): string
    {
        $config = getConfig();
        $privateKeyPath = $config['private_key_path'];

        if (!file_exists($privateKeyPath)) {
            throw new RuntimeException('Private key not found at ' . $privateKeyPath);
        }

        $privateKey = openssl_pkey_get_private(file_get_contents($privateKeyPath));
        if (!$privateKey) {
            throw new RuntimeException('Failed to load private key: ' . openssl_error_string());
        }

        openssl_sign($data, $signature, $privateKey, OPENSSL_ALGO_SHA256);
        openssl_free_key($privateKey);

        return base64_encode($signature);
    }

    /**
     * 用公钥验证签名
     */
    public static function verify(string $data, string $signature): bool
    {
        $config = getConfig();
        $publicKeyPath = $config['public_key_path'];

        if (!file_exists($publicKeyPath)) {
            return false;
        }

        $publicKey = openssl_pkey_get_public(file_get_contents($publicKeyPath));
        if (!$publicKey) {
            return false;
        }

        $result = openssl_verify($data, base64_decode($signature), $publicKey, OPENSSL_ALGO_SHA256);
        openssl_free_key($publicKey);

        return $result === 1;
    }
}
