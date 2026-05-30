package config

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	AppVersion    = "2.30.0"
	GitHubRepoURL = "https://github.com/860016/Clawmemory"
)

type Config struct {
	DatabasePath string
	JWTSecret    string
	DataDir      string
	SkillsDir    string
	BackupsDir   string
}

var (
	globalConfig *Config
	configOnce   sync.Once
)

func Load() *Config {
	configOnce.Do(func() {
		dataDir := getDataDir()
		skillsDir := getSkillsDir(dataDir)
		backupsDir := getBackupsDir(dataDir)

		jwtSecret := getEnv("SECRET_KEY", "clawmemory-default-secret-change-me")
		if jwtSecret == "clawmemory-default-secret-change-me" {
			jwtSecret = generateSecureSecret()
			fmt.Printf("[CONFIG] WARNING: SECRET_KEY not set, auto-generated a secure secret.\n")
			fmt.Printf("[CONFIG] To persist across restarts, set SECRET_KEY in your .env file.\n")
		}

		globalConfig = &Config{
			DatabasePath: filepath.Join(dataDir, "clawmemory.db"),
			JWTSecret:    jwtSecret,
			DataDir:      dataDir,
			SkillsDir:    skillsDir,
			BackupsDir:   backupsDir,
		}
	})
	return globalConfig
}

func generateSecureSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		fmt.Printf("[CONFIG] WARNING: failed to generate secure secret: %v, using fallback\n", err)
		h := sha256.Sum256([]byte(fmt.Sprintf("%d%s", time.Now().UnixNano(), os.Getenv("SECRET_KEY"))))
		return hex.EncodeToString(h[:])
	}
	return hex.EncodeToString(b)
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
