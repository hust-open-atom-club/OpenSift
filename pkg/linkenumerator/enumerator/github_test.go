package enumerator

import (
	"testing"
	"time"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/writer"
)

// Test GithubEnumerator.Enumerate with MemoryWriter
func TestGithubEnumerate(t *testing.T) {
	config := &GithubEnumeratorConfig{
		MinStars:    1000, // Lower star requirement
		StarOverlap: 0,
		Query:       "",
		Workers:     1,
		StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	e := NewGithubEnumerator(config)
	testWriter := writer.NewTestWriter()
	e.SetWriter(testWriter)
	err := e.Enumerate()
	if err != nil {
		t.Error(err)
	}
	for i, line := range testWriter.Lines {
		t.Logf("Repo %d: %s", i+1, line)
	}
	if len(testWriter.Lines) != 16 {
		t.Errorf("Expected 16 repos, got %d", len(testWriter.Lines))
	}
}
