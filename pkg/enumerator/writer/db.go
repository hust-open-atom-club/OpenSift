package writer

import (
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/HUSTSecLab/criticality_score/pkg/storage/repositories"
)

type DatabaseWriter struct {
	configPath  string
	dbCtx       storage.AppDatabaseContext
	repo        repositories.PlatformLinkRepository
	tablePrefix string

	buffer     []string
	bufferSize int
}

func NewDatabaseWriter(configPath string, tablePrefix string) *DatabaseWriter {
	return &DatabaseWriter{
		configPath:  configPath,
		tablePrefix: tablePrefix,
		buffer:      make([]string, 0),
		bufferSize:  1000,
	}
}

func (w *DatabaseWriter) Open() error {
	repo := repositories.NewPlatformLinkRepository(w.dbCtx, repositories.PlatformLinkTablePrefix(w.tablePrefix))

	return repo.ClearLinks()
}

func (w *DatabaseWriter) Close() error {
	return nil
}

func (w *DatabaseWriter) flush() error {
	return w.repo.BatchInsertLinks(w.buffer)
}

func (w *DatabaseWriter) Write(url string) error {
	w.buffer = append(w.buffer, url)

	if len(w.buffer) >= w.bufferSize {
		err := w.flush()
		if err != nil {
			logger.Error("Failed to flush buffer: %v", err)
			return err
		}
	}

	return nil
}
