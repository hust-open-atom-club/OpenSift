// PyPI API responsiveness test
package enumerator

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api"
)

// Test PyPI index API response
func TestPypiIndexAPIURLResponds(t *testing.T) {
	// Send GET request
	resp, err := http.Get(api.PYPI_INDEX_API_URL)
	if err != nil {
		t.Fatalf("Failed to get PYPI_INDEX_API_URL: %v", err)
	}
	defer resp.Body.Close()
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code from PYPI_INDEX_API_URL: %d", resp.StatusCode)
	}
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		t.Fatalf("Failed to read response body or body is empty from PYPI_INDEX_API_URL: %v", err)
	}
	t.Logf("PYPI_INDEX_API_URL responded successfully, content length: %d", len(body))
}

// Test PyPI enumerate API response
func TestPypiEnumerateAPIURLResponds(t *testing.T) {
	packageName := "requests"
	url := fmt.Sprintf("%s/%s/json", api.PYPI_ENUMERAE_API_URL, packageName)
	// Send GET request
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to get PYPI_ENUMERATE_API_URL: %v", err)
	}
	defer resp.Body.Close()
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code from PYPI_ENUMERATE_API_URL: %d", resp.StatusCode)
	}
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		t.Fatalf("Failed to read response body or body is empty from PYPI_ENUMERATE_API_URL: %v", err)
	}
	t.Logf("PYPI_ENUMERATE_API_URL responded successfully, content length: %d", len(body))
}
