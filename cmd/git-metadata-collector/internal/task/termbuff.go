package task

import (
	"strings"
	"sync"
)

// a goroutine safe bytes.TermBuffer
type TermBuffer struct {
	lines []string
	mutex sync.Mutex
}

// Write appends the contents of p to the buffer, growing the buffer as needed. It returns
// the number of bytes written.
func (s *TermBuffer) Write(p []byte) (n int, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	str := string(p)
	if s.lines == nil {
		s.lines = make([]string, 1)
	}
	for _, c := range str {
		if c == '\n' {
			s.lines = append(s.lines, "")
		} else if c == '\r' {
			s.lines[len(s.lines)-1] = ""
		} else {
			s.lines[len(s.lines)-1] += string(c)
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
	return strings.Join(s.lines, "\n")
}
