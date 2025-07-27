// Packagist API responsiveness test
package enumerator

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api"
)

// Test Packagist list API response
func TestPackagistListAPIURLResponds(t *testing.T) {
	// Send GET request
	resp, err := http.Get(api.PACKAGIST_LIST_API_URL)
	if err != nil {
		t.Fatalf("Failed to get PACKAGIST_LIST_API_URL: %v", err)
	}
	defer resp.Body.Close()
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code from PACKAGIST_LIST_API_URL: %d", resp.StatusCode)
	}
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		t.Fatalf("Failed to read response body or body is empty from PACKAGIST_LIST_API_URL: %v", err)
	}
	t.Logf("PACKAGIST_LIST_API_URL responded successfully, content length: %d", len(body))
}

// Test Packagist enumerate API response
func TestPackagistEnumerateAPIURLResponds(t *testing.T) {
	vendor := "symfony"
	name := "console"
	url := fmt.Sprintf("%s%s/%s.json", api.PACKAGIST_ENUMERATE_API_URL, vendor, name)
	// Send GET request
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to get PACKAGIST_ENUMERATE_API_URL: %v", err)
	}
	defer resp.Body.Close()
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code from PACKAGIST_ENUMERATE_API_URL: %d", resp.StatusCode)
	}
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		t.Fatalf("Failed to read response body or body is empty from PACKAGIST_ENUMERATE_API_URL: %v", err)
	}
	t.Logf("PACKAGIST_ENUMERATE_API_URL responded successfully, content length: %d", len(body))
}
