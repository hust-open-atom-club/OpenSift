// NPM API responsiveness test
package enumerator

import (
	"io"
	"net/http"
	"testing"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api"
)

// Test NPM index API response
func TestNpmIndexAPIURLResponds(t *testing.T) {
	// Send GET request
	resp, err := http.Get(api.NPM_INDEX_API_URL)
	if err != nil {
		t.Fatalf("Failed to get NPM_INDEX_API_URL: %v", err)
	}
	defer resp.Body.Close()
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code from NPM_INDEX_API_URL: %d", resp.StatusCode)
	}
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		t.Fatalf("Failed to read response body or body is empty from NPM_INDEX_API_URL: %v", err)
	}
	t.Logf("NPM_INDEX_API_URL responded successfully, content length: %d", len(body))
}

// Test NPM enumerate API response
func TestNpmEnumerateAPIURLResponds(t *testing.T) {
	packageName := "express"
	url := api.NPM_ENUMERATE_API_URL + packageName
	// Send GET request
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to get NPM_ENUMERATE_API_URL: %v", err)
	}
	defer resp.Body.Close()
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code from NPM_ENUMERATE_API_URL: %d", resp.StatusCode)
	}
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		t.Fatalf("Failed to read response body or body is empty from NPM_ENUMERATE_API_URL: %v", err)
	}
	t.Logf("NPM_ENUMERATE_API_URL responded successfully, content length: %d", len(body))
}
