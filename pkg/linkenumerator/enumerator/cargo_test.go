package enumerator

import (
	"testing"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/writer"
)

// Test CargoEnumerator.Enumerate with MemoryWriter
func TestCargoEnumerate(t *testing.T) {
	e := NewCargoEnumerator(5)
	testWriter := writer.NewTestWriter()
	e.SetWriter(testWriter)
	err := e.Enumerate()
	if err != nil {
		t.Error(err)
	}
	// Print output content
	for i, line := range testWriter.Lines {
		t.Logf("Crate %d: %s", i+1, line)
	}
	// Assert output count
	if len(testWriter.Lines) != 30 {
		t.Errorf("Expected 30 repos, got %d", len(testWriter.Lines))
	}
}
