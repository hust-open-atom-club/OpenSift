package task

import (
	"strings"
	"sync"

	"github.com/samber/lo"
)

// a goroutine safe bytes.TermBuffer
type TermBuffer struct {
	lines [][]byte
	pos   int
	mutex sync.Mutex
}

// Write appends the contents of p to the buffer, growing the buffer as needed. It returns
// the number of bytes written.
func (s *TermBuffer) Write(p []byte) (n int, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.lines == nil {
		s.lines = make([][]byte, 0)
		s.lines = append(s.lines, make([]byte, 0))
	}
	for _, c := range p {
		if c == '\n' {
			s.lines = append(s.lines, make([]byte, 0))
			s.pos = 0
		} else if c == '\r' {
			s.pos = 0
		} else {
			lastLine := &s.lines[len(s.lines)-1]
			if s.pos >= len(*lastLine) {
				s.lines[len(s.lines)-1] = append(s.lines[len(s.lines)-1], c)
				s.pos++
			} else {
				(*lastLine)[s.pos] = c
				s.pos++
			}
		}

	}
	return len(p), nil
}

// String returns the contents of the unread portion of the buffer
// as a string.  If the Buffer is a nil pointer, it returns "<nil>".
func (s *TermBuffer) String() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.lines == nil {
		return ""
	}
	return strings.Join(lo.Map(s.lines, func(i []byte, _ int) string {
		return string(i)
	}), "\n")
}
