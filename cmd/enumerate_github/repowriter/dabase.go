package repowriter

import (
	"database/sql"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
)

type DatabaseWriter struct {
	db *sql.DB
}

// Text creates a new Writer instance that is used to write a simple text file
// of repositories, where each line has a single repository url.
func Database(config string) (*DatabaseWriter, error) {
	storage.InitializeDefaultAppDatabase(config)
	db, err := storage.GetDefaultDatabaseConnection()
	if err != nil {
		return nil, err
	}
	return &DatabaseWriter{db}, nil

}

// Write implements the Writer interface.
func (w *DatabaseWriter) Write(repo string) error {

	var exists bool

	err := w.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM github_links WHERE git_link = $1)`, repo).Scan(&exists)

	if err != nil {
		return err
	}

	if !exists {
		_, err := w.db.Exec(`INSERT INTO github_links (git_link) VALUES ($1)`, repo)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *DatabaseWriter) Begin() error {
	_, err := w.db.Exec(`DELETE FROM github_links`)
	return err
}
