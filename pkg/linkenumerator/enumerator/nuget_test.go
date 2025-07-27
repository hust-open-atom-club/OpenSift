// NuGet API responsiveness test
package enumerator

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api"
)

// Test NuGet index API response
func TestNugetIndexAPIURLResponds(t *testing.T) {
	// Send GET request
	resp, err := http.Get(api.NUGET_INDEX_URL)
	if err != nil {
		t.Fatalf("Failed to get NUGET_INDEX_URL: %v", err)
	}
	defer resp.Body.Close()
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code from NUGET_INDEX_URL: %d", resp.StatusCode)
	}
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		t.Fatalf("Failed to read response body or body is empty from NUGET_INDEX_URL: %v", err)
	}
	t.Logf("NUGET_INDEX_URL responded successfully, content length: %d", len(body))
}

// Test NuGet enumerate API response
func TestNugetEnumerateAPIURLResponds(t *testing.T) {
	url := fmt.Sprintf("%s?take=%d&skip=%d", api.NUGET_INDEX_URL, 20, 0)
	// Send GET request
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to get NUGET_ENUMERATE_API_URL: %v", err)
	}
	defer resp.Body.Close()
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code from NUGET_ENUMERATE_API_URL: %d", resp.StatusCode)
	}
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		t.Fatalf("Failed to read response body or body is empty from NUGET_ENUMERATE_API_URL: %v", err)
	}
	t.Logf("NUGET_ENUMERATE_API_URL responded successfully, content length: %d", len(body))
}
