// Bitbucket API responsiveness test
package enumerator

import (
	"io"
	"net/http"
	"testing"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api"
)

// Test Bitbucket enumerate API URL response
func TestBitbucketIndexAPIURLResponds(t *testing.T) {
	// Send GET request
	resp, err := http.Get(api.BITBUCKET_ENUMERATE_API_URL)
	if err != nil {
		t.Fatalf("Failed to get BITBUCKET_ENUMERATE_API_URL: %v", err)
	}
	defer resp.Body.Close()
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code from BITBUCKET_ENUMERATE_API_URL: %d", resp.StatusCode)
	}
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		t.Fatalf("Failed to read response body or body is empty from BITBUCKET_ENUMERATE_API_URL: %v", err)
	}
	t.Logf("BITBUCKET_ENUMERATE_API_URL responded successfully, content length: %d", len(body))
}
