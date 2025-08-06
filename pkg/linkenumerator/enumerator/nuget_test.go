package enumerator

import (
	"testing"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/writer"
)

// Test NugetEnumerator.Enumerate with MemoryWriter
func TestNugetEnumerate(t *testing.T) {
	e := NewNugetEnumerator(100)
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
	if len(testWriter.Lines) != 500 {
		t.Errorf("Expected 500 lines, got %d", len(testWriter.Lines))
	}
}
