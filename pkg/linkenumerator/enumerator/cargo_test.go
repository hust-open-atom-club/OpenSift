// Cargo API responsiveness test
package enumerator

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api"
)

// Test crates.io index API response
func TestCargoIndexAPIURLResponds(t *testing.T) {
	url := api.CRATES_IO_ENUMERATE_API_URL + "?sort=downloads&per_page=100&page=1"
	// Send GET request
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to get CRATES_IO_ENUMERATE_API_URL: %v", err)
	}
	defer resp.Body.Close()
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code from CRATES_IO_ENUMERATE_API_URL: %d", resp.StatusCode)
	}
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		t.Fatalf("Failed to read response body or body is empty from CRATES_IO_ENUMERATE_API_URL: %v", err)
	}
	t.Logf("CRATES_IO_ENUMERATE_API_URL responded successfully, content length: %d", len(body))
}

// Test crates.io single crate API response
func TestCargoEnumerateAPIURLResponds(t *testing.T) {
	crateName := "serde"
	url := fmt.Sprintf("%s/%s", api.CRATES_IO_ENUMERATE_API_URL, crateName)
	// Send GET request
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to get CRATES_IO_ENUMERATE_API_URL for crate: %v", err)
	}
	defer resp.Body.Close()
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code from CRATES_IO_ENUMERATE_API_URL for crate: %d", resp.StatusCode)
	}
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		t.Fatalf("Failed to read response body or body is empty from CRATES_IO_ENUMERATE_API_URL for crate: %v", err)
	}
	t.Logf("CRATES_IO_ENUMERATE_API_URL for crate responded successfully, content length: %d", len(body))
}
