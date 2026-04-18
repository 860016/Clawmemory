<?php
/**
 * OpenClaw License Platform - 邮件服务
 * 支持 SMTP 发送到期提醒等通知邮件
 */

class Mailer
{
    /**
     * 发送邮件（通过 PHP mail 或 SMTP）
     */
    public static function send(string $to, string $subject, string $body): bool
    {
        $smtpHost = Settings::get('smtp_host');
        
        if (!empty($smtpHost)) {
            return self::sendSmtp($to, $subject, $body);
        }

        // 降级到 PHP mail()
        $headers = "From: " . (Settings::get('smtp_from') ?: 'noreply@openclaw.ai') . "\r\n";
        $headers .= "Content-Type: text/html; charset=UTF-8\r\n";
        return mail($to, $subject, $body, $headers);
    }

    /**
     * SMTP 发送
     */
    private static function sendSmtp(string $to, string $subject, string $body): bool
    {
        $host = Settings::get('smtp_host');
        $port = (int) Settings::get('smtp_port', 465);
        $user = Settings::get('smtp_user', '');
        $pass = Settings::get('smtp_pass', '');
        $from = Settings::get('smtp_from', $user);

        if (empty($host) || empty($user)) {
            error_log("SMTP not configured, skipping email to $to");
            return false;
        }

        // 使用 fsockopen 实现 SMTP（无依赖）
        $errno = 0;
        $errstr = '';
        $prefix = ($port === 465) ? 'ssl://' : '';
        $socket = @fsockopen($prefix . $host, $port, $errno, $errstr, 10);

        if (!$socket) {
            error_log("SMTP connect failed: $errstr ($errno)");
            return false;
        }

        self::smtpRead($socket); // 220

        self::smtpSend($socket, "EHLO " . gethostname());
        self::smtpRead($socket); // 250

        self::smtpSend($socket, "STARTTLS");
        self::smtpRead($socket); // 220

        // 升级到 TLS
        if (!stream_socket_enable_crypto($socket, true, STREAM_CRYPTO_METHOD_TLS_CLIENT)) {
            error_log("SMTP TLS handshake failed");
            fclose($socket);
            return false;
        }

        self::smtpSend($socket, "EHLO " . gethostname());
        self::smtpRead($socket); // 250

        self::smtpSend($socket, "AUTH LOGIN");
        self::smtpRead($socket); // 334

        self::smtpSend($socket, base64_encode($user));
        self::smtpRead($socket); // 334

        self::smtpSend($socket, base64_encode($pass));
        $authResp = self::smtpRead($socket); // 235

        if (strpos($authResp, '235') !== 0) {
            error_log("SMTP auth failed: $authResp");
            fclose($socket);
            return false;
        }

        self::smtpSend($socket, "MAIL FROM: <$from>");
        self::smtpRead($socket); // 250

        self::smtpSend($socket, "RCPT TO: <$to>");
        self::smtpRead($socket); // 250

        self::smtpSend($socket, "DATA");
        self::smtpRead($socket); // 354

        $msg = "From: $from\r\n";
        $msg .= "To: $to\r\n";
        $msg .= "Subject: =?UTF-8?B?" . base64_encode($subject) . "?=\r\n";
        $msg .= "Content-Type: text/html; charset=UTF-8\r\n";
        $msg .= "\r\n";
        $msg .= $body;
        $msg .= "\r\n.";

        self::smtpSend($socket, $msg);
        self::smtpRead($socket); // 250

        self::smtpSend($socket, "QUIT");
        fclose($socket);

        return true;
    }

    private static function smtpSend($socket, string $data): void
    {
        fwrite($socket, $data . "\r\n");
    }

    private static function smtpRead($socket): string
    {
        $response = '';
        while ($line = fgets($socket, 515)) {
            $response .= $line;
            if (substr($line, 3, 1) === ' ') break;
        }
        return $response;
    }

    /**
     * 发送到期提醒邮件
     */
    public static function sendExpiryReminder(array $license, int $daysLeft): bool
    {
        $email = $license['customer_email'] ?? '';
        if (empty($email)) return false;

        $appUrl = getConfig()['app_url'];
        $subject = "授权码即将到期 - 还剩 {$daysLeft} 天";
        $body = "
        <html><body style='font-family: sans-serif; color: #333;'>
        <div style='max-width:600px;margin:0 auto;padding:20px;'>
            <h2 style='color:#4f46e5;'>授权到期提醒</h2>
            <p>您好 {$license['customer_name']},</p>
            <p>您的授权码 <strong>{$license['license_key']}</strong> 将在 <strong>{$daysLeft} 天后到期</strong>。</p>
            <p>到期时间：" . ($license['expires_at'] ?? '未知') . "</p>
            <p>套餐：" . strtoupper($license['tier']) . "</p>
            <hr style='border:none;border-top:1px solid #eee;margin:20px 0;'>
            <p style='color:#666;font-size:14px;'>如需续期，请联系管理员或访问 <a href='{$appUrl}'>{$appUrl}</a></p>
        </div>
        </body></html>";

        return self::send($email, $subject, $body);
    }
}
