<?php
/**
 * OpenClaw License Platform - 数据库连接类
 * 纯 PHP PDO 封装，兼容 MySQL 5.7
 */

class Database
{
    private static ?PDO $instance = null;

    public static function getInstance(): PDO
    {
        if (self::$instance === null) {
            $config = getConfig();
            $db = $config['db'];

            $dsn = sprintf(
                'mysql:host=%s;port=%s;dbname=%s;charset=%s',
                $db['host'],
                $db['port'],
                $db['database'],
                $db['charset']
            );

            try {
                self::$instance = new PDO($dsn, $db['username'], $db['password'], [
                    PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
                    PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
                    PDO::ATTR_EMULATE_PREPARES => false,
                    PDO::MYSQL_ATTR_INIT_COMMAND => "SET NAMES utf8mb4 COLLATE utf8mb4_unicode_ci",
                ]);
            } catch (PDOException $e) {
                error_log('Database connection failed: ' . $e->getMessage());
                http_response_code(500);
                echo json_encode(['error' => 'Database connection failed']);
                exit(1);
            }
        }

        return self::$instance;
    }

    /**
     * 查询一行
     */
    public static function fetch(string $sql, array $params = []): ?array
    {
        $stmt = self::getInstance()->prepare($sql);
        $stmt->execute($params);
        $result = $stmt->fetch();
        return $result ?: null;
    }

    /**
     * 查询多行
     */
    public static function fetchAll(string $sql, array $params = []): array
    {
        $stmt = self::getInstance()->prepare($sql);
        $stmt->execute($params);
        return $stmt->fetchAll();
    }

    /**
     * 执行写操作（INSERT/UPDATE/DELETE），返回影响行数
     */
    public static function execute(string $sql, array $params = []): int
    {
        $stmt = self::getInstance()->prepare($sql);
        $stmt->execute($params);
        return $stmt->rowCount();
    }

    /**
     * 插入数据，返回 lastInsertId
     */
    public static function insert(string $sql, array $params = []): string
    {
        $stmt = self::getInstance()->prepare($sql);
        $stmt->execute($params);
        return self::getInstance()->lastInsertId();
    }

    /**
     * 查询单个值
     */
    public static function fetchColumn(string $sql, array $params = []): mixed
    {
        $stmt = self::getInstance()->prepare($sql);
        $stmt->execute($params);
        return $stmt->fetchColumn();
    }
}
