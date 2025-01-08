package writer

import (
	"database/sql"
	"fmt"

	"github.com/HUSTSecLab/criticality_score/internal/logger"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
)

type DatabaseWriter struct {
	configPath   string
	tableToWrite string
	db           *sql.DB
}

func NewDatabaseWriter(configPath string, tableToWrite string) *DatabaseWriter {
	return &DatabaseWriter{
		configPath:   configPath,
		tableToWrite: tableToWrite,
	}
}

func (w *DatabaseWriter) Open() error {
	_, err := storage.InitializeDefaultAppDatabase(w.configPath)
	if err != nil {
		return err
	}
	conn, err := storage.GetDefaultDatabaseConnection()

	if err != nil {
		return err
	}
	w.db = conn

	conn.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (git_link VARCHAR(255) NOT NULL PRIMARY KEY)", w.tableToWrite))
	conn.Exec(fmt.Sprintf("DELETE FROM %s", w.tableToWrite))

	return nil
}

func (w *DatabaseWriter) Close() error {
	return w.db.Close()
}

func (w *DatabaseWriter) Write(url string) error {
	_, err := w.db.Exec(fmt.Sprintf("INSERT INTO %s (git_link) VALUES ($1)", w.tableToWrite), url)
	if err != nil {
		logger.Warnf("failed to insert repository %s: %v", url, err)
	}
	return err
}
