package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	DatabasePath      string
	LicenseServerURL  string
	RSAPublicKeyPath  string
	JWTSecret         string
	DataDir           string
}

func Load() *Config {
	dataDir := getDataDir()

	return &Config{
		DatabasePath:      filepath.Join(dataDir, "clawmemory.db"),
		LicenseServerURL:  getEnv("LICENSE_SERVER_URL", "https://license.clawmemory.com"),
		RSAPublicKeyPath:  filepath.Join(dataDir, "keys", "public.pem"),
		JWTSecret:         getEnv("JWT_SECRET", "clawmemory-default-secret-change-me"),
		DataDir:           dataDir,
	}
}

func getDataDir() string {
	// Docker 环境
	if _, err := os.Stat("/app/data"); err == nil {
		return "/app/data"
	}

	// 本地开发
	exe, _ := os.Executable()
	return filepath.Join(filepath.Dir(exe), "data")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
