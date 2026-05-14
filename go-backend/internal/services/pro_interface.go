package services

import (
	proprovider "github.com/860016/clawmemory-pro/proprovider"

	"clawmemory/internal/config"

	"gorm.io/gorm"
)

type ProProvider = proprovider.ProProvider

var ErrProRequired = proprovider.ErrProRequired

var globalProProvider ProProvider

func InitProProvider(db *gorm.DB, cfg *config.Config) ProProvider {
	p := NewLocalProProvider(db)
	globalProProvider = p
	return p
}

func GetProProvider() ProProvider {
	return globalProProvider
}
