package services

import (
	"clawmemory/internal/models"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

type GracefulShutdown struct {
	db        *gorm.DB
	mu        sync.Mutex
	cancelFns []context.CancelFunc
	shutdown  bool
}

func NewGracefulShutdown(db *gorm.DB) *GracefulShutdown {
	return &GracefulShutdown{db: db}
}

func (gs *GracefulShutdown) RegisterCancel(fn context.CancelFunc) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.cancelFns = append(gs.cancelFns, fn)
}

func (gs *GracefulShutdown) Shutdown(timeout time.Duration) {
	gs.mu.Lock()
	if gs.shutdown {
		gs.mu.Unlock()
		return
	}
	gs.shutdown = true
	gs.mu.Unlock()

	logger := GetLogger()
	logger.Info("graceful shutdown initiated", map[string]interface{}{
		"timeout": timeout.String(),
	})

	done := make(chan struct{})
	go func() {
		sqlDB, err := gs.db.DB()
		if err == nil {
			sqlDB.SetMaxOpenConns(0)
			time.Sleep(100 * time.Millisecond)
			sqlDB.Close()
			logger.Info("database connections closed")
		}

		gs.mu.Lock()
		for _, fn := range gs.cancelFns {
			fn()
		}
		gs.mu.Unlock()

		close(done)
	}()

	select {
	case <-done:
		logger.Info("graceful shutdown completed")
	case <-time.After(timeout):
		logger.Warn("graceful shutdown timed out, forcing exit")
	}
}

type BackupService struct {
	db        *gorm.DB
	backupDir string
}

func NewBackupService(db *gorm.DB, backupDir string) *BackupService {
	os.MkdirAll(backupDir, 0755)
	return &BackupService{db: db, backupDir: backupDir}
}

func (bs *BackupService) CreateBackup(dbPath string) (string, error) {
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("clawmemory-backup-%s.db", timestamp)
	backupPath := filepath.Join(bs.backupDir, backupName)

	sqlDB, err := bs.db.DB()
	if err != nil {
		return "", fmt.Errorf("failed to get underlying db: %w", err)
	}

	safePath := strings.ReplaceAll(backupPath, "'", "''")
	_, err = sqlDB.Exec(fmt.Sprintf("VACUUM INTO '%s'", safePath))
	if err != nil {
		return "", fmt.Errorf("VACUUM INTO failed: %w", err)
	}

	dstHash, err := bs.computeFileHash(backupPath)
	if err != nil {
		os.Remove(backupPath)
		return "", fmt.Errorf("failed to compute backup hash: %w", err)
	}

	checksumFile := backupPath + ".sha256"
	if err := os.WriteFile(checksumFile, []byte(dstHash), 0644); err != nil {
		log.Printf("Warning: failed to write checksum file: %v", err)
	}

	bs.cleanupOldBackups(10)

	logger := GetLogger()
	logger.Info("backup created", map[string]interface{}{
		"path": backupPath,
		"hash": dstHash[:16],
	})

	return backupPath, nil
}

func (bs *BackupService) VerifyBackup(backupPath string) error {
	checksumFile := backupPath + ".sha256"
	expectedHash, err := os.ReadFile(checksumFile)
	if err != nil {
		return fmt.Errorf("checksum file not found: %w", err)
	}

	actualHash, err := bs.computeFileHash(backupPath)
	if err != nil {
		return fmt.Errorf("failed to compute hash: %w", err)
	}

	if actualHash != string(expectedHash) {
		return fmt.Errorf("integrity check failed: expected=%s actual=%s", string(expectedHash)[:16], actualHash[:16])
	}

	return nil
}

func (bs *BackupService) RestoreBackup(backupPath, dbPath string) error {
	if err := bs.VerifyBackup(backupPath); err != nil {
		return fmt.Errorf("backup verification failed: %w", err)
	}

	preBackup, err := bs.CreateBackup(dbPath)
	if err != nil {
		log.Printf("Warning: failed to create pre-restore backup: %v", err)
	} else {
		log.Printf("Pre-restore backup saved: %s", preBackup)
	}

	src, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create db file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to restore: %w", err)
	}

	logger := GetLogger()
	logger.Info("backup restored", map[string]interface{}{
		"backup": backupPath,
		"target": dbPath,
	})

	return nil
}

func (bs *BackupService) ListBackups() ([]map[string]interface{}, error) {
	entries, err := os.ReadDir(bs.backupDir)
	if err != nil {
		return nil, err
	}

	var backups []map[string]interface{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".db" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		backups = append(backups, map[string]interface{}{
			"name": entry.Name(),
			"size": info.Size(),
			"date": info.ModTime().Format(time.RFC3339),
		})
	}

	return backups, nil
}

func (bs *BackupService) computeFileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func (bs *BackupService) cleanupOldBackups(keep int) {
	entries, err := os.ReadDir(bs.backupDir)
	if err != nil {
		return
	}

	var dbFiles []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".db" {
			dbFiles = append(dbFiles, entry)
		}
	}

	if len(dbFiles) <= keep {
		return
	}

	for i := 0; i < len(dbFiles)-keep; i++ {
		path := filepath.Join(bs.backupDir, dbFiles[i].Name())
		os.Remove(path)
		os.Remove(path + ".sha256")
	}
}

type MigrationService struct {
	db *gorm.DB
}

func NewMigrationService(db *gorm.DB) *MigrationService {
	return &MigrationService{db: db}
}

func (ms *MigrationService) GetCurrentVersion() int {
	var migration models.SchemaMigration
	if err := ms.db.Where("key = ?", "schema_version").FirstOrCreate(&migration).Error; err != nil {
		return 0
	}
	return migration.Version
}

func (ms *MigrationService) SetVersion(version int) {
	var migration models.SchemaMigration
	ms.db.Where("key = ?", "schema_version").FirstOrCreate(&migration)
	ms.db.Model(&migration).Updates(map[string]interface{}{
		"version": version,
		"dirty":   false,
	})
}
