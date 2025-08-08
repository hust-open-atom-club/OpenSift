package enumerator

import (
	"testing"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/writer"
)

// Test GitLabEnumerator.Enumerate with MemoryWriter
func TestGitLabEnumerate(t *testing.T) {
	e := NewGitlabEnumerator(5, 2)
	testWriter := writer.NewTestWriter()
	e.SetWriter(testWriter)
	err := e.Enumerate()
	if err != nil {
		t.Error(err)
	}
	// Print output content
	for i, line := range testWriter.Lines {
		t.Logf("Repo %d: %s", i+1, line)
	}
	// Assert output count
	if len(testWriter.Lines) != 15 {
		t.Errorf("Expected 15 lines, got %d", len(testWriter.Lines))
	}
}
