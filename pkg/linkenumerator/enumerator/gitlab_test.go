// GitLab API responsiveness test
package enumerator

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api"
)

// Test GitLab enumerate API URL response
func TestGitlabEnumerateAPIURLResponds(t *testing.T) {
	// Build request URL
	url := fmt.Sprintf("%s?order_by=star_count&sort=desc&per_page=%d&page=1", api.GITLAB_ENUMERATE_API_URL, api.PER_PAGE)
	// Send GET request
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to get GITLAB_ENUMERATE_API_URL: %v", err)
	}
	defer resp.Body.Close()
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code from GITLAB_ENUMERATE_API_URL: %d", resp.StatusCode)
	}
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		t.Fatalf("Failed to read response body or body is empty from GITLAB_ENUMERATE_API_URL: %v", err)
	}
	t.Logf("GITLAB_ENUMERATE_API_URL responded successfully, content length: %d", len(body))
}
