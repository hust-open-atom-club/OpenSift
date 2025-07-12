package storage

import (
	"fmt"

	"github.com/HUSTSecLab/OpenSift/pkg/logger"
)

const (
	ViperDbConfigPathKey = "db_config_path"
	DbConfigEnv          = "APPDB_CONFIG_PATH"
)

var defaultAppDatabase AppDatabaseContext

func GetDefaultConfig() (*Config, error) {
	if defaultAppDatabase == nil {
		return nil, fmt.Errorf("default app database is not initialized")
	}

	config := defaultAppDatabase.GetConfig()
	return &config, nil
}

func InitDefaultDatabaseContext(cfg *Config) {
	defaultAppDatabase = NewAppDatabase(cfg)
}

func GetDefaultAppDatabaseContext() AppDatabaseContext {
	if defaultAppDatabase == nil {
		logger.Panic("Before using default app database, you should call InitDefault first")
	}

	return defaultAppDatabase
}
