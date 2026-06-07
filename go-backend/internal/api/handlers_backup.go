package api

import (
	"clawmemory/internal/config"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleListBackups(c *gin.Context) {
	cfg := config.Load()
	backupDir := cfg.BackupsDir
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		c.JSON(http.StatusOK, gin.H{"backups": []interface{}{}})
		return
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"backups": []interface{}{}})
		return
	}

	backups := make([]map[string]interface{}, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".db") && !strings.HasSuffix(entry.Name(), ".sql") && !strings.HasSuffix(entry.Name(), ".zip") {
			continue
		}
		info, _ := entry.Info()
		backups = append(backups, map[string]interface{}{
			"filename":   entry.Name(),
			"size":       info.Size(),
			"created_at": info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"backups": backups})
}

func handleCreateBackup(c *gin.Context) {
	cfg := config.Load()
	backupDir := cfg.BackupsDir
	os.MkdirAll(backupDir, 0755)

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("clawmemory_backup_%s.db", timestamp)
	backupPath := filepath.Join(backupDir, filename)

	dbPath := cfg.DatabasePath

	src, err := os.Open(dbPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("Cannot open database file: %v", err),
		})
		return
	}
	defer src.Close()

	dst, err := os.Create(backupPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("Cannot create backup file: %v", err),
		})
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("Backup failed: %v", err),
		})
		return
	}

	fi, _ := dst.Stat()

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"filename": filename,
		"path":     backupPath,
		"size":     fi.Size(),
	})
}

func handleDownloadBackup(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := c.Param("filename")
		if strings.Contains(filename, "..") || strings.ContainsAny(filename, "/\\") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
			return
		}
		cfg := config.Load()
		backupPath := filepath.Join(cfg.BackupsDir, filename)

		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "backup not found"})
			return
		}

		c.FileAttachment(backupPath, filename)
	}
}

func handleRestoreBackup(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := c.Param("filename")
		if strings.Contains(filename, "..") || strings.ContainsAny(filename, "/\\") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
			return
		}
		cfg := config.Load()
		backupPath := filepath.Join(cfg.BackupsDir, filename)

		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "backup not found"})
			return
		}

		dbPath := cfg.DatabasePath
		preRestorePath := dbPath + ".pre-restore"
		if _, err := os.Stat(dbPath); err == nil {
			if err := copyFile(dbPath, preRestorePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create safety backup before restore"})
				return
			}
		}

		src, err := os.Open(backupPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read backup file"})
			return
		}
		defer src.Close()

		dst, err := os.Create(dbPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot write database file"})
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			os.Rename(preRestorePath, dbPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "restore failed, rolled back to previous state"})
			return
		}

		os.Remove(preRestorePath)

		c.JSON(http.StatusOK, gin.H{"message": "backup restored successfully", "filename": filename})
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func handleDeleteBackup(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := c.Param("filename")
		if strings.Contains(filename, "..") || strings.ContainsAny(filename, "/\\") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
			return
		}
		cfg := config.Load()
		backupPath := filepath.Join(cfg.BackupsDir, filename)

		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "backup not found"})
			return
		}

		if err := os.Remove(backupPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot delete backup file"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "backup deleted", "filename": filename})
	}
}
