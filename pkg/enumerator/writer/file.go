package writer

import (
	"os"
	"sync"

	"github.com/HUSTSecLab/criticality_score/pkg/logger"
)

type TextWriter struct {
	fileName string
	file     *os.File
	muWrite  sync.Mutex
}

func NewTextFileWriter(fileName string) *TextWriter {
	return &TextWriter{
		fileName: fileName,
		muWrite:  sync.Mutex{},
	}
}

func (w *TextWriter) Open() error {
	var err error
	// if exists, print warning
	if _, err = os.Stat(w.fileName); err == nil {
		logger.Warnf("file %s already exists, will be overwritten", w.fileName)
	}

	// open file
	w.file, err = os.OpenFile(w.fileName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (w *TextWriter) Close() error {
	return w.file.Close()
}

func (w *TextWriter) Write(gitlink string) error {
	w.muWrite.Lock()
	defer w.muWrite.Unlock()

	_, err := w.file.WriteString(gitlink + "\n")
	return err
}
