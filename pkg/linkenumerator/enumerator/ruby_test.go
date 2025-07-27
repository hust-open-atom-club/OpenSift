// Ruby API responsiveness test
package enumerator

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api"
)

// Test Ruby index API response
func TestRubyIndexAPIURLResponds(t *testing.T) {
	// Send GET request
	resp, err := http.Get(api.RUBY_INDEX_API_URL)
	if err != nil {
		t.Fatalf("Failed to get RUBY_INDEX_API_URL: %v", err)
	}
	defer resp.Body.Close()
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code from RUBY_INDEX_API_URL: %d", resp.StatusCode)
	}
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		t.Fatalf("Failed to read response body or body is empty from RUBY_INDEX_API_URL: %v", err)
	}
	t.Logf("RUBY_INDEX_API_URL responded successfully, content length: %d", len(body))
}

// Test Ruby enumerate API response
func TestRubyEnumerateAPIURLResponds(t *testing.T) {
	gemName := "rails"
	url := fmt.Sprintf("%s/%s.json", api.RUBY_ENUMERATE_API_URL, gemName)
	// Send GET request
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to get RUBY_ENUMERATE_API_URL: %v", err)
	}
	defer resp.Body.Close()
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code from RUBY_ENUMERATE_API_URL: %d", resp.StatusCode)
	}
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		t.Fatalf("Failed to read response body or body is empty from RUBY_ENUMERATE_API_URL: %v", err)
	}
	t.Logf("RUBY_ENUMERATE_API_URL responded successfully, content length: %d", len(body))
}
