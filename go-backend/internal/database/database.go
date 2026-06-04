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
		&models.FileSyncIndex{},
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

	if db.Migrator().HasTable(&models.User{}) {
		if db.Migrator().HasColumn(&models.User{}, "is_founder") {
			result := db.Model(&models.User{}).
				Where("is_founder = ? AND role = ?", false, "admin").
				Update("is_founder", true)
			if result.RowsAffected > 0 {
				fmt.Printf("[DB] set %d existing admin users as founder\n", result.RowsAffected)
			}
		}
	}

	if db.Migrator().HasTable(&models.AuditLog{}) {
		result := db.Where("created_at < datetime('now', '-90 day')").Delete(&models.AuditLog{})
		if result.RowsAffected > 0 {
			fmt.Printf("[DB] cleaned up %d audit logs older than 90 days\n", result.RowsAffected)
		}
	}

	if db.Migrator().HasTable(&models.Memory{}) {
		layerMigrations := []struct {
			oldLayer      string
			newLayer      string
			newMemoryType string
			extraUpdates  string
		}{
			{"preference", "core", "preference", ""},
			{"knowledge", "context", "knowledge", ""},
			{"episodic", "detail", "fact", ""},
			{"semantic", "context", "knowledge", ""},
			{"procedural", "context", "knowledge", ""},
			{"short_term", "detail", "fact", ""},
			{"private", "core", "knowledge", ", visibility = 'private'"},
		}
		for _, lm := range layerMigrations {
			var count int64
			db.Model(&models.Memory{}).Where("layer = ?", lm.oldLayer).Count(&count)
			if count > 0 {
				sql := fmt.Sprintf("UPDATE memories SET layer = '%s', memory_type = '%s'%s WHERE layer = '%s'", lm.newLayer, lm.newMemoryType, lm.extraUpdates, lm.oldLayer)
				result := db.Exec(sql)
				if result.Error != nil {
					fmt.Printf("[DB] layer migration '%s' → '%s' error: %v\n", lm.oldLayer, lm.newLayer, result.Error)
				} else {
					fmt.Printf("[DB] migrated %d memories: layer '%s' → '%s' (memory_type='%s')\n", result.RowsAffected, lm.oldLayer, lm.newLayer, lm.newMemoryType)
				}
			}
		}

		var noLayerCount int64
		db.Model(&models.Memory{}).Where("layer = '' OR layer IS NULL").Count(&noLayerCount)
		if noLayerCount > 0 {
			result := db.Model(&models.Memory{}).Where("layer = '' OR layer IS NULL").Updates(map[string]interface{}{"layer": "context", "memory_type": "knowledge"})
			if result.Error == nil {
				fmt.Printf("[DB] assigned layer='context' to %d memories with missing layer\n", result.RowsAffected)
			}
		}
	}

	if db.Migrator().HasTable(&models.Setting{}) && db.Migrator().HasTable(&models.User{}) {
		var orphanCount int64
		db.Model(&models.Setting{}).
			Where("user_id NOT IN (?)", db.Model(&models.User{}).Select("id")).
			Count(&orphanCount)
		if orphanCount > 0 {
			var firstUser models.User
			if err := db.Order("id ASC").First(&firstUser).Error; err == nil {
				result := db.Model(&models.Setting{}).
					Where("user_id NOT IN (?)", db.Model(&models.User{}).Select("id")).
					Update("user_id", firstUser.ID)
				if result.Error != nil {
					fmt.Printf("[DB] orphan settings fix error: %v\n", result.Error)
				} else {
					fmt.Printf("[DB] fixed %d orphan settings to user_id=%d\n", result.RowsAffected, firstUser.ID)
				}
			}
		}
	}

	return nil
}
