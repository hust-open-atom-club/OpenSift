package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

const enableSQLLog = true

type AppDatabase struct {
	Config Config
}

var DefaultAppDatabase *AppDatabase

func NewAppDatabase(configPath string) (*AppDatabase, error) {
	config, err := loadConfig(configPath)
	if err != nil {
		fmt.Errorf("Failed to load config:", err)
		return nil, err
	}
	return &AppDatabase{Config: config}, nil
}

func (app *AppDatabase) GetDatabaseConnection() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		app.Config.Host, app.Config.Port, app.Config.User, app.Config.Password, app.Config.Database)
	db, err := sql.Open("postgres", connStr)
	return db, err
}

func (app *AppDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	if enableSQLLog {
		logrus.Info("Exec SQL: ", query)
	}

	conn, err := app.GetDatabaseConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return conn.Exec(query, args...)
}

func (app *AppDatabase) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if enableSQLLog {
		logrus.Info("Query SQL: ", query)
	}

	conn, err := app.GetDatabaseConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return conn.Query(query, args...)
}

// Deprecated: Do not use global app database
func InitializeDefaultAppDatabase(configPath string) (*AppDatabase, error) {
	var err error
	DefaultAppDatabase, err = NewAppDatabase(configPath)
	if err != nil {
		return nil, err
	}
	return DefaultAppDatabase, nil
}

// Deprecated: Do not use global app database
func GetDefaultDatabaseConnection() (*sql.DB, error) {
	if DefaultAppDatabase == nil {
		return nil, fmt.Errorf("default app database is not initialized")
	}
	return DefaultAppDatabase.GetDatabaseConnection()
}
