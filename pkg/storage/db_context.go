package storage

import (
	"database/sql"
	"fmt"

	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	_ "github.com/lib/pq"
)

type AppDatabaseContext interface {
	GetConfig() Config
	NewBatchExecContext(config *BatchExecContextConfig) BatchExecContext
	GetDatabaseConnection() (*sql.DB, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type appDatabaseContext struct {
	Config       Config
	EnableSQLLog bool
}

var DefaultAppDatabase AppDatabaseContext

func NewAppDatabase(configPath string) (AppDatabaseContext, error) {
	config, err := loadConfig(configPath)
	if err != nil {
		fmt.Errorf("Failed to load config:", err)
		return nil, err
	}
	return &appDatabaseContext{Config: config}, nil
}

func (appDb *appDatabaseContext) GetConfig() Config {
	return appDb.Config
}

func (appDb *appDatabaseContext) NewBatchExecContext(config *BatchExecContextConfig) BatchExecContext {
	return &batchExecContext{
		appDb:          appDb,
		config:         config,
		sentences:      "",
		sentencesCount: 0,
		args:           make([]interface{}, 0),
	}
}

func (app *appDatabaseContext) GetDatabaseConnection() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		app.Config.Host, app.Config.Port, app.Config.User, app.Config.Password, app.Config.Database)
	db, err := sql.Open("postgres", connStr)
	return db, err
}

func (app *appDatabaseContext) Exec(query string, args ...interface{}) (sql.Result, error) {
	if app.EnableSQLLog {
		logger.Info("Exec SQL: ", query)
	}

	conn, err := app.GetDatabaseConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return conn.Exec(query, args...)
}

func (app *appDatabaseContext) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if app.EnableSQLLog {
		logger.Info("Query SQL: ", query)
	}

	conn, err := app.GetDatabaseConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return conn.Query(query, args...)
}

func (app *appDatabaseContext) QueryRow(query string, args ...interface{}) *sql.Row {
	if app.EnableSQLLog {
		logger.Info("QueryRow SQL: ", query)
	}

	conn, err := app.GetDatabaseConnection()
	if err != nil {
		return nil
	}
	defer conn.Close()
	return conn.QueryRow(query, args...)
}

// Deprecated: Do not use global app database
func InitializeDefaultAppDatabase(configPath string) (AppDatabaseContext, error) {
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
