package enumerator

import (
	"testing"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/writer"
)

// Test NpmEnumerator.Enumerate with MemoryWriter
func TestNpmEnumerate(t *testing.T) {
	e := NewNpmEnumerator(5)
	testWriter := writer.NewTestWriter()
	e.SetWriter(testWriter)
	err := e.Enumerate()
	if err != nil {
		t.Error(err)
	}
	// Print output content
	for i, line := range testWriter.Lines {
		t.Logf("Package %d: %s", i+1, line)
	}
	// Assert output count
	if len(testWriter.Lines) != 25 {
		t.Errorf("Expected 25 lines, got %d", len(testWriter.Lines))
	}
}
