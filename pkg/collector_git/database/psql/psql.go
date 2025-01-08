/*
 * @Date: 2024-09-07 16:30:21
 * @LastEditTime: 2025-01-07 19:07:26
 * @Description:
 */
package psql

import (
	"fmt"

	"github.com/HUSTSecLab/criticality_score/pkg/collector_git/config"
	"github.com/HUSTSecLab/criticality_score/pkg/collector_git/database"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDBFromStorageConfig() (*gorm.DB, error) {
	config := storage.DefaultAppDatabase.GetConfig()

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.Host,
		config.User,
		config.Password,
		config.Database,
		config.Port,
		"disable",
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	return db, err
}

func InitDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.PSQL_HOST,
		config.PSQL_USER,
		config.PSQL_PASSWORD,
		config.PSQL_DATABASE_NAME,
		config.PSQL_PORT,
		config.PSQL_SSL_MODE,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	return db, err
}

func CreateTable(db *gorm.DB) error {
	return db.AutoMigrate(&database.GitMetrics{})
}

func InsertTable(db *gorm.DB, metrics *database.GitMetrics) {
	db.Where(&database.GitMetrics{URL: metrics.URL}).
		Assign(database.GitMetrics{
			Name:             metrics.Name,
			Owner:            metrics.Owner,
			Source:           metrics.Source,
			URL:              metrics.URL,
			License:          metrics.License,
			Ecosystems:       metrics.Ecosystems,
			Languages:        metrics.Languages,
			CreatedSince:     metrics.CreatedSince,
			UpdatedSince:     metrics.UpdatedSince,
			ContributorCount: metrics.ContributorCount,
			OrgCount:         metrics.OrgCount,
			CommitFrequency:  metrics.CommitFrequency,
			NeedUpdate:       metrics.NeedUpdate,
		}).FirstOrCreate(metrics)
}
