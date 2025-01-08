package storage

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Database    string `json:"database"`
	User        string `json:"user"`
	Password    string `json:"password"`
	Host        string `json:"host"`
	Port        string `json:"port"`
	GitHubToken string `json:"GitHubToken"`
	GitLabToken string `json:"GitLabToken"`
	Redispass   string `json:"redispass"`
}

// Deprecated: do not use global app database context
func GetDefaultConfig() (*Config, error) {
	if DefaultAppDatabase == nil {
		return nil, fmt.Errorf("default app database is not initialized")
	}

	return &DefaultAppDatabase.(*appDatabaseContext).Config, nil
}

func loadConfig(configPath string) (Config, error) {
	var config Config
	file, err := os.ReadFile(configPath)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(file, &config)
	return config, err
}
