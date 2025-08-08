package enumerator

import (
	"testing"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/writer"
)

// Test BitBucketEnumerator.Enumerate with MemoryWriter
func TestBitBucketEnumerate(t *testing.T) {
	e := NewBitBucketEnumerator(5)
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
	if len(testWriter.Lines) != 5 {
		t.Errorf("Expected 5 repos, got %d", len(testWriter.Lines))
	}
}
