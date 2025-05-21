package tool

import (
	"io"
)

type ToolReader struct {
	buffer chan []byte
	// logFile io.Writer
}

// const MaxBufferSize = 1024 * 1024 // 1MB

var _ io.Reader = (*ToolReader)(nil)

func (r *ToolReader) WriteBuffer(p []byte) (int, error) {
	r.buffer <- p
	return len(p), nil
}

func (r *ToolReader) Read(p []byte) (n int, err error) {
	newRead := <-r.buffer
	n = copy(p, newRead)

	// r.logFile.Write(p[:n])
	return
}

type ToolWriter struct {
	writer  []io.Writer
	logFile io.Writer
}

var _ io.Writer = (*ToolWriter)(nil)

func (w *ToolWriter) AddWriter(writer io.Writer) {
	w.writer = append(w.writer, writer)
}

func (w *ToolWriter) RemoveWriter(writer io.Writer) {
	for i, wr := range w.writer {
		if wr == writer {
			w.writer = append(w.writer[:i], w.writer[i+1:]...)
			break
		}
	}
}

func (w *ToolWriter) Write(p []byte) (n int, err error) {
	if len(w.writer) != 0 {
		for _, wr := range w.writer {
			pn, e := wr.Write(p)
			if e == nil {
				if pn > n {
					n = pn
				}
			}
		}
	} else {
		n = len(p)
	}
	if w.logFile != nil {
		w.logFile.Write(p[:n])
	}
	return
}
