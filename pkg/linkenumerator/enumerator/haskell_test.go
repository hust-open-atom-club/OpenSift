// Haskell API responsiveness test
package enumerator

import (
	"io"
	"net/http"
	"testing"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api"
)

// Test Haskell index API response
func TestHaskellIndexAPIURLResponds(t *testing.T) {
	// Send GET request
	resp, err := http.Get(api.HASKELL_INDEX_API_URL)
	if err != nil {
		t.Fatalf("Failed to get HASKELL_INDEX_API_URL: %v", err)
	}
	defer resp.Body.Close()
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code from HASKELL_INDEX_API_URL: %d", resp.StatusCode)
	}
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		t.Fatalf("Failed to read response body or body is empty from HASKELL_INDEX_API_URL: %v", err)
	}
	t.Logf("HASKELL_INDEX_API_URL responded successfully, content length: %d", len(body))
}

// Test Haskell enumerate API response
func TestHaskellEnumerateAPIURLResponds(t *testing.T) {
	packageName := "aeson"
	url := api.HASKELL_ENUMERATE_API_URL + "/package/" + packageName
	// Send GET request
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to get HASKELL_ENUMERATE_API_URL: %v", err)
	}
	defer resp.Body.Close()
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code from HASKELL_ENUMERATE_API_URL: %d", resp.StatusCode)
	}
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		t.Fatalf("Failed to read response body or body is empty from HASKELL_ENUMERATE_API_URL: %v", err)
	}
	t.Logf("HASKELL_ENUMERATE_API_URL responded successfully, content length: %d", len(body))
}
