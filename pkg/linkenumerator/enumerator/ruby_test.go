package enumerator

import (
	"testing"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/writer"
)

// Test RubyEnumerator.Enumerate with MemoryWriter
func TestRubyEnumerate(t *testing.T) {
	e := NewRubyEnumerator(5)
	testWriter := writer.NewTestWriter()
	e.SetWriter(testWriter)
	err := e.Enumerate()
	if err != nil {
		t.Error(err)
	}
	// Print output content
	for i, line := range testWriter.Lines {
		t.Logf("Gem %d: %s", i+1, line)
	}
	// Assert output count
	if len(testWriter.Lines) != 28 {
		t.Errorf("Expected 28 lines, got %d", len(testWriter.Lines))
	}
}
