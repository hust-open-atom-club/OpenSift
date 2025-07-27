// GitHub API responsiveness test
package enumerator

import (
	"io"
	"net/http"
	"testing"
)

// Test GitHub REST API repository search
func TestGithubRestAPIResponds(t *testing.T) {
	// Send GET request
	url := "https://api.github.com/search/repositories?q=stars:%3E100000&sort=stars&order=desc&per_page=1"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to get GitHub REST API: %v", err)
	}
	defer resp.Body.Close()
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code from GitHub REST API: %d", resp.StatusCode)
	}
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		t.Fatalf("Failed to read response body or body is empty from GitHub REST API: %v", err)
	}
	t.Logf("GitHub REST API responded successfully, content length: %d", len(body))
}

// Test GitHub GraphQL API accessibility
func TestGithubGraphQLAPIResponds(t *testing.T) {
	url := "https://api.github.com/graphql"
	// Send GET request
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to get GitHub GraphQL API: %v", err)
	}
	defer resp.Body.Close()
	// Check HTTP status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
		t.Fatalf("Unexpected status code from GitHub GraphQL API: %d", resp.StatusCode)
	}
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body from GitHub GraphQL API: %v", err)
	}
	t.Logf("GitHub GraphQL API responded, status code: %d, content length: %d", resp.StatusCode, len(body))
}
