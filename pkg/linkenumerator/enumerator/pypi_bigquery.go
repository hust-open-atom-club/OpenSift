package enumerator

import (
	"context"

	"cloud.google.com/go/bigquery"
	"github.com/HUSTSecLab/OpenSift/pkg/logger"
	"google.golang.org/api/iterator"
)

type PypiBigQueryEnumeratorConfig struct {
	ProjectID string
}

type pypiBigQueryEnumerator struct {
	enumeratorBase
	config *PypiBigQueryEnumeratorConfig
}

// Enumerate implements Enumerator.
func (p *pypiBigQueryEnumerator) Enumerate() error {
	if err := p.writer.Open(); err != nil {
		return err
	}
	defer p.writer.Close()

	ctx := context.Background()

	client, err := bigquery.NewClient(ctx, p.config.ProjectID)
	if err != nil {
		logger.Error("Failed to create BigQuery client: ", err)
		return err
	}
	defer client.Close()

	q := client.Query(`SELECT DISTINCT TRIM(SPLIT(url,',')[SAFE_OFFSET(1)]) as git_link
	FROM bigquery-public-data.pypi.distribution_metadata t, UNNEST(t.project_urls) AS url
	WHERE starts_with(url, "Source") OR starts_with(url, "Repository") OR starts_with(url, "Source Code") OR starts_with(url, "repository") OR starts_with(url, "Github") OR starts_with(url, "Code") OR starts_with(url, "source") OR starts_with(url, "repo")`)
	it, err := q.Read(ctx)

	if err != nil {
		logger.Error("Failed to run BigQuery query: ", err)
	}

	for {
		var row struct {
			GitLink string `bigquery:"git_link"`
		}
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			logger.Error("Failed to read BigQuery result: ", err)
		}

		p.writer.Write(row.GitLink)
	}
	return nil
}

var _ Enumerator = (*pypiBigQueryEnumerator)(nil)

func NewPypiBigQueryEnumerator(cfg *PypiBigQueryEnumeratorConfig) Enumerator {
	return &pypiBigQueryEnumerator{
		enumeratorBase: newEnumeratorBase(),
		config:         cfg,
	}
}
