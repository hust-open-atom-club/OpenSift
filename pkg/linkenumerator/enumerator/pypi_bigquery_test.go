package enumerator

import (
	"os"
	"testing"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/writer"
)

// Test PypiBigQueryEnumerator.Enumerate with MemoryWriter
func TestPypiBigQueryEnumerate(t *testing.T) {
	projectID := os.Getenv("PYPI_BIGQUERY_PROJECT_ID")
	if projectID == "" {
		t.Skip("PYPI_BIGQUERY_PROJECT_ID not set, skipping test")
	}
	cfg := &PypiBigQueryEnumeratorConfig{
		ProjectID: projectID,
	}
	e := NewPypiBigQueryEnumerator(cfg)
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
	if len(testWriter.Lines) == 0 {
		t.Errorf("No repos collected")
	}
}
