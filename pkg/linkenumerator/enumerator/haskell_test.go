package enumerator

import (
	"testing"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/writer"
)

// Test HaskellEnumerator.Enumerate with MemoryWriter
func TestHaskellEnumerate(t *testing.T) {
	e := NewHaskellEnumerator(5)
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
	if len(testWriter.Lines) != 30 {
		t.Errorf("Expected 30 lines, got %d", len(testWriter.Lines))
	}
}
