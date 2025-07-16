package enumerator

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/hasura/go-graphql-client"
	"github.com/ossf/scorecard/v4/clients/githubrepo/roundtripper"
	"github.com/ossf/scorecard/v4/log"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/githubapi"
	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/githubsearch"
	"github.com/HUSTSecLab/OpenSift/pkg/logger"
)

const (
	reposPerPage     = 100
	oneDay           = time.Hour * 24
	runIDToken       = "[[runid]]"
	runIDDateFormat  = "20060102-1504"
	githubDateFormat = "2006-01-02"
)

var (
	// epochDate is the earliest date for which GitHub has data.
	GithubEpochDate = time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC)
	runID           = time.Now().UTC().Format(runIDDateFormat)
)

type GithubEnumeratorConfig struct {
	MinStars        int
	StarOverlap     int
	RequireMinStars bool
	Query           string
	Workers         int
	StartDate       time.Time
	EndDate         time.Time
}

type githubEnumerator struct {
	enumeratorBase
	config *GithubEnumeratorConfig
}

func NewGithubEnumerator(config *GithubEnumeratorConfig) Enumerator {
	return &githubEnumerator{
		enumeratorBase: newEnumeratorBase(),
		config:         config,
	}
}

// searchWorker waits for a query on the queries channel, starts a search with that query using s
// and returns each repository on the results channel.
func (c *githubEnumerator) searchWorker(s *githubsearch.Searcher, logger logger.AppLogger, queries, results chan string) {
	for q := range queries {
		total := 0
		err := s.ReposByStars(q, c.config.MinStars, c.config.StarOverlap, func(repo string) {
			results <- repo
			total++
		})
		if err != nil {
			// TODO: this error handling is not at all graceful, and hard to recover from.
			logger.WithFields(map[string]any{
				"query": q,
				"error": err,
			}).Error("Enumeration failed for query")
			if errors.Is(err, githubsearch.ErrorUnableToListAllResult) {
				if *&c.config.RequireMinStars {
					os.Exit(1)
				}
			} else {
				os.Exit(1)
			}
		}
		logger.WithFields(map[string]interface{}{
			"query":      q,
			"repo_count": total,
		}).Info("Enumeration for query done")
	}
}

func (c *githubEnumerator) Enumerate() error {
	// Warn if the -start date is before the epoch.
	if c.config.StartDate.Before(GithubEpochDate) {
		logger.Warn("-start date is before epoch", map[string]any{
			"start": c.config.StartDate.Format(githubDateFormat),
			"epoch": GithubEpochDate.Format(githubDateFormat),
		})
	}

	// Ensure -start is before -end
	if c.config.EndDate.Before(c.config.StartDate) {
		logger.Error("-start date must be before -end date", map[string]any{
			"start": c.config.StartDate.Format(githubDateFormat),
			"end":   c.config.EndDate.Format(githubDateFormat),
		})
		os.Exit(2)
	}

	// We need a context to support a bunch of operations.
	ctx := context.Background()

	rtLogger := log.NewLogger(log.InfoLevel)

	// Prepare a client for communicating with GitHub's GraphQL API.
	// Do this before opening the output file to avoid creating an empty file
	// if we fail to authenticate, or connect to the authentication server.
	rt := githubapi.NewRetryRoundTripper(roundtripper.NewTransport(ctx, rtLogger), logger.GetDefaultLogger())
	httpClient := &http.Client{
		Transport: rt,
	}
	client := graphql.NewClient(githubapi.DefaultGraphQLEndpoint, httpClient).WithDebug(true)

	w := c.writer
	w.Open()
	defer w.Close()

	logger.WithFields(map[string]any{
		"start":        c.config.StartDate.String(),
		"end":          c.config.EndDate.String(),
		"min_stars":    c.config.MinStars,
		"star_overlap": c.config.StarOverlap,
		"workers":      c.config.Workers,
	}).Info("Starting enumeration")

	// Track how long it takes to enumerate the repositories
	startTime := time.Now()

	baseQuery := c.config.Query
	queries := make(chan string)
	results := make(chan string, c.config.Workers*reposPerPage)

	// Start the worker goroutines to execute the search queries
	pool := gopool.NewPool("github-enumerator", int32(c.config.Workers), gopool.NewConfig())

	var wg sync.WaitGroup

	for i := 0; i < c.config.Workers; i++ {
		wg.Add(1)

		pool.Go(func() {
			defer wg.Done()

			workerLogger := logger.WithFields(map[string]interface{}{
				"worker": i,
			})

			s := githubsearch.NewSearcher(ctx, client, workerLogger, githubsearch.PerPage(reposPerPage))
			c.searchWorker(s, workerLogger, queries, results)
		})
	}

	// Start a separate goroutine to collect results so worker output is always consumed.
	done := make(chan bool)
	totalRepos := 0
	go func() {
		for repo := range results {
			w.Write(repo)
			totalRepos++
		}
		done <- true
	}()

	// Work happens here. Iterate through the dates from today, until the start date.
	for created := c.config.EndDate; !c.config.StartDate.After(created); created = created.Add(-oneDay) {
		logger.WithFields(map[string]any{
			"created": created.Format(githubDateFormat),
		}).Info("Scheduling day for enumeration")
		queries <- baseQuery + fmt.Sprintf(" created:%s", created.Format(githubDateFormat))
	}
	logger.Debug("Waiting for workers to finish")
	// Indicate to the workers that we're finished.
	close(queries)
	// Wait for the workers to be finished.
	wg.Wait()

	logger.Debug("Waiting for writer to finish")
	// Close the results channel now the workers are done.
	close(results)
	// Wait for the writer to be finished.
	<-done

	logger.WithFields(map[string]any{
		"total_repos": totalRepos,
		"duration":    time.Since(startTime).Truncate(time.Minute),
	}).Info("Finished enumeration")

	return nil
}
