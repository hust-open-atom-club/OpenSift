package writer

import (
	"github.com/HUSTSecLab/OpenSift/pkg/logger"
	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/HUSTSecLab/OpenSift/pkg/storage/repository"
)

type DatabaseWriter struct {
	dbCtx       storage.AppDatabaseContext
	repo        repository.PlatformLinkRepository
	tablePrefix string

	buffer     []string
	bufferSize int
}

func NewDatabaseWriter(ctx storage.AppDatabaseContext, tablePrefix string) *DatabaseWriter {
	return &DatabaseWriter{
		dbCtx:       ctx,
		tablePrefix: tablePrefix,
		buffer:      make([]string, 0),
		bufferSize:  1000,
	}
}

func (w *DatabaseWriter) Open() error {
	repo := repository.NewPlatformLinkRepository(w.dbCtx, repository.PlatformLinkTablePrefix(w.tablePrefix))
	w.repo = repo
	return repo.BeginTemp()
}

func (w *DatabaseWriter) Close() error {
	w.flush()
	return w.repo.CommitTemp()
}

func (w *DatabaseWriter) flush() error {
	err := w.repo.BatchInsertTemp(w.buffer)
	if err != nil {
		logger.Error("Failed to insert links: %v", err)
		return err
	}
	w.buffer = make([]string, 0)
	return nil
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
