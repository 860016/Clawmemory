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
	SkillsDir         string
	BackupsDir        string
}

func Load() *Config {
	dataDir := getDataDir()
	skillsDir := getSkillsDir(dataDir)
	backupsDir := getBackupsDir(dataDir)

	return &Config{
		DatabasePath:      filepath.Join(dataDir, "clawmemory.db"),
		LicenseServerURL:  getEnv("LICENSE_SERVER_URL", "https://auth.bestu.top"),
		RSAPublicKeyPath:  filepath.Join(dataDir, "keys", "public.pem"),
		JWTSecret:         getEnv("SECRET_KEY", "clawmemory-default-secret-change-me"),
		DataDir:           dataDir,
		SkillsDir:         skillsDir,
		BackupsDir:        backupsDir,
	}
}

func getDataDir() string {
	if dir := getEnv("DATA_DIR", ""); dir != "" {
		return dir
	}

	if _, err := os.Stat("/app/data"); err == nil {
		return "/app/data"
	}

	exe, _ := os.Executable()
	return filepath.Join(filepath.Dir(exe), "data")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getSkillsDir(dataDir string) string {
	if dir := getEnv("SKILLS_DIR", ""); dir != "" {
		return dir
	}
	return filepath.Join(dataDir, "skills")
}

func getBackupsDir(dataDir string) string {
	if dir := getEnv("BACKUPS_DIR", ""); dir != "" {
		return dir
	}
	return filepath.Join(dataDir, "backups")
}
