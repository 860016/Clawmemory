package database

import (
	"fmt"
	"os"
	"path/filepath"

	"clawmemory/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Init(dbPath string) (*gorm.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	return db, nil
}

func Migrate(db *gorm.DB) error {
	if err := runPreMigrations(db); err != nil {
		fmt.Printf("[DB] pre-migration warning: %v\n", err)
	}

	return db.AutoMigrate(
		&models.User{},
		&models.Memory{},
		&models.Entity{},
		&models.Relation{},
		&models.WikiPage{},
		&models.Project{},
		&models.ProjectNote{},
		&models.DailyReport{},
		&models.License{},
		&models.Backup{},
		&models.Setting{},
		&models.APIKey{},
		&models.AuditLog{},
		&models.SessionMemory{},
		&models.ReasoningConfig{},
		&models.MemoryShare{},
		&models.ShareRule{},
		&models.Invitation{},
		&models.SystemLog{},
		&models.SchemaMigration{},
		&models.MemoryHistory{},
		&models.UserProfile{},
		&models.ActionTrace{},
		&models.Skill{},
		&models.SkillSuggestion{},
	)
}

func runPreMigrations(db *gorm.DB) error {
	if db.Migrator().HasTable(&models.Skill{}) {
		if db.Migrator().HasIndex(&models.Skill{}, "Name") && !db.Migrator().HasIndex(&models.Skill{}, "idx_user_skill") {
			db.Exec("DROP INDEX IF EXISTS idx_skills_name")
			db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_user_skill ON skills(user_id, name)")
		}
	}

	if db.Migrator().HasTable(&models.ActionTrace{}) {
		var staleCount int64
		db.Model(&models.ActionTrace{}).Where("created_at < datetime('now', '-30 day')").Count(&staleCount)
		if staleCount > 0 {
			result := db.Where("created_at < datetime('now', '-30 day')").Delete(&models.ActionTrace{})
			if result.Error != nil {
				fmt.Printf("[DB] action_traces cleanup error: %v\n", result.Error)
			} else {
				fmt.Printf("[DB] cleaned up %d stale action_traces\n", result.RowsAffected)
			}
		}
	}

	if db.Migrator().HasTable(&models.APIKey{}) {
		result := db.Model(&models.APIKey{}).
			Where("permissions = ?", "read,write").
			Update("permissions", "memories:read,memories:write,conversations:write,sessions:write,reason:execute")
		if result.RowsAffected > 0 {
			fmt.Printf("[DB] migrated %d API keys from legacy 'read,write' to fine-grained permissions\n", result.RowsAffected)
		}
	}

	return nil
}
