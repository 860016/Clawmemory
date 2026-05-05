//go:build !pro

package services

import (
	"clawmemory/internal/config"

	"gorm.io/gorm"
)

type ProGuard struct{}

func InitProGuard(proxy *ProProxy, db *gorm.DB, cfg *config.Config) *ProGuard {
	return &ProGuard{}
}

func GetProGuard() *ProGuard {
	return nil
}

func (g *ProGuard) IsProFeatureEnabled(feature string) bool {
	return false
}

func (g *ProGuard) SelfCheck() bool {
	return true
}

func (g *ProGuard) InvalidateDerivedKey() {}
