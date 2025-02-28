///usr/bin/true; exec /usr/bin/env go run "$0" "$@"

package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/HUSTSecLab/criticality_score/pkg/config"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/collector"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/git"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/url"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/util"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/HUSTSecLab/criticality_score/pkg/storage/repository"
	"github.com/HUSTSecLab/criticality_score/pkg/storage/sqlutil"
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/lib/pq"
	"github.com/spf13/pflag"
)

func fix(ctx storage.AppDatabaseContext, link string) {

	logger.WithFields(map[string]any{
		"link": link,
	}).Infof("Fixing language of %s", link)
	u := url.ParseURL(link)
	path := util.GetGitRepositoryPath(config.GetGitStoragePath(), &u)
	r, err := collector.Open(path)

	if err != nil {
		logger.Errorf("Open %s failed: %s", link, err)
		return
	}

	result := git.NewRepo()
	err = result.WalkRepo(r)

	if err != nil {
		logger.Errorf("WalkRepo %s failed: %s", link, err)
		return
	}

	// Get latest id of the link
	var id int
	srow := ctx.QueryRow(`select id from scores where git_link = $1 order by id desc limit 1`, link)
	if err = srow.Scan(&id); err != nil {
		logger.Errorf("Failed to get latest id of %s: %v", link, err)
		return
	}

	// Get git_metrics id
	srows, err := ctx.Query(`select git_metrics_id from scores_git where score_id = $1`, id)
	if err != nil {
		logger.Errorf("Failed to get git_metrics_id of %s: %v", link, err)
		return
	}

	for srows.Next() {
		var gmid int
		if err = srows.Scan(&gmid); err != nil {
			logger.Errorf("Failed to get git_metrics_id of %s: %v", link, err)
			continue
		}
		_, err := ctx.Exec(`UPDATE git_metrics SET
				language = $1,
				license = $2
				WHERE id = $3`,
			pq.StringArray(result.Languages),
			pq.StringArray(result.Licenses),
			gmid)

		if err != nil {
			logger.Errorf("Update database for %s failed: %v", link, err)
			continue
		}

		logger.WithFields(map[string]any{
			"link":           link,
			"id":             id,
			"git_metrics_id": gmid,
		}).Info("Success to fix %s", link)
	}

}

func main() {
	flagJobs := pflag.IntP("jobs", "j", 8, "Number of jobs to run concurrently")

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "This program fetch links from latest ranking, and fix the language of latest metadata.\n")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		pflag.PrintDefaults()
	}

	config.RegistCommonFlags(pflag.CommandLine)
	config.RegistGitStorageFlags(pflag.CommandLine)
	config.ParseFlags(pflag.CommandLine)

	logger.SetContext("git-metadata-fixer")

	ctx := storage.GetDefaultAppDatabaseContext()
	rr := repository.NewResultRepository(ctx)

	// refresh ranking cache
	err := rr.MakeRankingCache()

	if err != nil {
		logger.Fatalf("Failed to refresh ranking cache: %v\n", err)
		os.Exit(1)
	}

	// Fetch links from latest ranking
	items, err := sqlutil.Query[repository.RankingResult](ctx, `select * from rankings_cache`)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to fetch links from latest ranking: %v\n", err)
		os.Exit(1)
	}

	pool := gopool.NewPool("fix-metadata", int32(*flagJobs), &gopool.Config{})
	var wg sync.WaitGroup

	for row := range items {
		link := row.GitLink
		if link == nil {
			logger.Warnf("Link is nil, skipping")
			continue
		}
		wg.Add(1)
		pool.Go(func() {
			defer wg.Done()
			fix(ctx, *link)
		})
	}

	wg.Wait()
}
