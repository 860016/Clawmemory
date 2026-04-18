# OpenClaw License Platform v2

纯 PHP 授权管理平台，无需 Laravel/Composer/Artisan，上传到宝塔面板即可使用。

## 部署步骤

### 1. 上传文件

将以下文件/目录上传到宝塔网站根目录（如 `/www/wwwroot/auth.bestu.top`）：

```
index.php          - 入口文件
config.php         - 配置文件
.env               - 环境变量（从 .env.example 复制）
database.sql       - 数据库建表脚本
src/               - 核心代码
views/             - 页面模板
scripts/           - 工具脚本
keys/              - RSA密钥（生成后）
```

### 2. 创建数据库

1. 在宝塔面板创建 MySQL 数据库（如 `openclaw_license`）
2. 打开 phpMyAdmin，选择该数据库
3. 导入 `database.sql` 文件

### 3. 配置环境

复制 `.env.example` 为 `.env`，修改数据库连接信息：

```
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=openclaw_license
DB_USERNAME=你的数据库用户名
DB_PASSWORD=你的数据库密码
```

### 4. 生成 RSA 密钥

在服务器上执行：

```bash
cd /www/wwwroot/auth.bestu.top
php scripts/generate_keys.php
```

> **⚠️ 权限问题**：`generate_keys.php` 以 root 身份执行时，生成的 `keys/` 目录权限为 `0700`（仅 root 可读），而 PHP-FPM 通常以 `www` 用户运行，导致 API 读取公钥失败返回 404。生成密钥后**必须修复权限**：
>
> ```bash
> chown -R www:www /www/wwwroot/auth.bestu.top/keys
> chmod 755 /www/wwwroot/auth.bestu.top/keys
> chmod 644 /www/wwwroot/auth.bestu.top/keys/public.pem
> chmod 600 /www/wwwroot/auth.bestu.top/keys/private.pem
> ```
>
> 验证权限是否正确：访问 `https://你的域名/api/v1/public-key`，应返回 PEM 公钥内容。如果返回 404，说明 `www` 用户无法读取密钥文件。

### 5. 配置 Nginx（⚠️ 必须配置，否则 API 返回 404）

在宝塔的网站设置中，添加以下 Nginx 伪静态规则（也可直接使用项目中的 `nginx.conf`）：

```nginx
location / {
    try_files $uri $uri/ /index.php?$query_string;
}

location ~ \.env$ {
    deny all;
}

location ~ /keys/ {
    deny all;
}
```

> **重要**：如果不配置 `try_files`，`/api/v1/*` 等路由会返回 404，因为 Nginx 找不到对应文件直接返回 404，不会转发到 PHP。
> 配置后验证：`curl https://你的域名/api/v1/public-key` 应返回 PEM 公钥内容。

### 6. 访问

- 管理后台：`https://auth.bestu.top/login`
- 默认账号：`admin@openclaw.ai`
- 默认密码：`admin123`

### API 接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/v1/activate` | POST | 激活授权码 |
| `/api/v1/verify` | POST | 验证授权码 |
| `/api/v1/heartbeat` | POST | 心跳上报 |
| `/api/v1/ping` | POST | 安装统计 |
| `/api/v1/public-key` | GET | 获取 RSA 公钥 |

## 与 Laravel 版本的区别

- 不需要 Composer、Artisan、PHP 8.2+ 等依赖
- 不需要 `composer install`
- 不需要 `php artisan migrate`
- 支持 PHP 7.4+ / MySQL 5.7+（兼容宝塔面板默认 PHP 版本）
- API 接口完全兼容，客户端代码无需修改
- RSA 签名机制完全相同

## 安全说明

- RSA-2048 签名确保授权数据不可伪造
- 私钥保存在服务器，客户端只有公钥
- PDO 预处理语句防止 SQL 注入
- Session 管理 + 登录验证
- `.env` 和 `keys/` 目录通过 Nginx 禁止外部访问
