// NOTE:
// These tests require Google Cloud credentials to access BigQuery.
// Please set the environment variable GOOGLE_APPLICATION_CREDENTIALS to your service account JSON key file before running:
// export GOOGLE_APPLICATION_CREDENTIALS=/path/to/your/key.json
// For more information, see: https://cloud.google.com/docs/authentication/external/set-up-adc

package enumerator

import (
	"context"
	"testing"

	"cloud.google.com/go/bigquery"
)

func TestBigQueryClientConnect(t *testing.T) {
	projectID := "bigquery-public-data"
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		t.Fatalf("Failed to create BigQuery client: %v", err)
	}
	defer client.Close()
	t.Logf("BigQuery client created successfully for project: %s", projectID)
}

func TestBigQueryQueryResponds(t *testing.T) {
	projectID := "bigquery-public-data"
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		t.Fatalf("Failed to create BigQuery client: %v", err)
	}
	defer client.Close()

	q := client.Query(`SELECT DISTINCT TRIM(SPLIT(url,',')[SAFE_OFFSET(1)]) as git_link
        FROM bigquery-public-data.pypi.distribution_metadata t, UNNEST(t.project_urls) AS url
        WHERE starts_with(url, "Source") OR starts_with(url, "Repository") OR starts_with(url, "Source Code") OR starts_with(url, "repository") OR starts_with(url, "Github") OR starts_with(url, "Code") OR starts_with(url, "source") OR starts_with(url, "repo")`)
	it, err := q.Read(ctx)
	if err != nil {
		t.Fatalf("Failed to run BigQuery query: %v", err)
	}

	// Only fetch the first 5 results for a simple responsiveness test
	count := 0
	for {
		var row struct {
			GitLink string `bigquery:"git_link"`
		}
		err := it.Next(&row)
		if err != nil {
			break
		}
		count++
		if count >= 5 {
			break
		}
	}
	if count == 0 {
		t.Fatalf("No results returned from BigQuery query")
	}
	t.Logf("BigQuery query responded successfully, rows fetched: %d", count)
}
