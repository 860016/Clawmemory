package database

import (
	"os"
	"path/filepath"

	"clawmemory/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Init(dbPath string) (*gorm.DB, error) {
	// 确保目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Memory{},
		&models.Entity{},
		&models.Relation{},
		&models.WikiPage{},
		&models.DailyReport{},
		&models.License{},
		&models.Backup{},
	)
}
