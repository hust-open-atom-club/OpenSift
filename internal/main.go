/*
 * @Date: 2024-09-06 21:09:14
 * @LastEditTime: 2025-01-07 19:08:48
 * @Description: The Cli for collector
 */
package main

import (
	"log"
	"os"
	"strings"
	"sync"

	collector "github.com/HUSTSecLab/criticality_score/pkg/collector_git/collector"
	git "github.com/HUSTSecLab/criticality_score/pkg/collector_git/parser/git"
	url "github.com/HUSTSecLab/criticality_score/pkg/collector_git/parser/url"
	scores "github.com/HUSTSecLab/criticality_score/pkg/gen_scores"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/bytedance/gopkg/util/gopool"
	gogit "github.com/go-git/go-git/v5"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "collector_git",
		Usage: "Collect Git-based Repository Metrics",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Usage:    "Path to config.json",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "update-db",
				Usage: "Whether to update the database",
			},
		},
		Action: func(c *cli.Context) error {
			configPath := c.String("config")
			updateDB := c.Bool("update-db")
			paths := c.Args().Slice()

			var wg sync.WaitGroup
			wg.Add(len(paths))

			repos := make([]*git.Repo, 0)

			for _, path := range paths {
				gopool.Go(func() {
					defer wg.Done()
					logger.Infof("Collecting %s", path)

					r := &gogit.Repository{}
					var err error

					if strings.Contains(path, "://") {
						u := url.ParseURL(path)
						r, err = collector.EzCollect(&u)
						if err != nil {
							logger.Panicf("Collecting %s Failed", u.URL)
						}
					} else {
						r, err = collector.Open(path)
						if err != nil {
							logger.Panicf("Opening %s Failed", path)
						}
					}

					repo, err := git.ParseRepo(r)
					if err != nil {
						logger.Panicf("Parsing %s Failed", path)
					}

					repos = append(repos, repo)
					logger.Infof("%s Collected", repo.Name)
				})
			}

			wg.Wait()

			if updateDB {
				storage.InitializeDefaultAppDatabase(configPath)
				db, err := storage.GetDefaultDatabaseConnection()
				if err != nil {
					log.Fatalf("Failed to connect to database: %v", err)
				}

				defer db.Close()
				for _, repo := range repos {
					repo.Show()
					projectData := scores.ProjectData{
						CommitFrequency:  &repo.CommitFrequency,
						ContributorCount: &repo.ContributorCount,
						CreatedSince:     &repo.CreatedSince,
						UpdatedSince:     &repo.UpdatedSince,
						Org_Count:        &repo.OrgCount,
						Pkg_Manager:      &repo.Ecosystems,
					}
					var totalnum float64
					linkCount := make(map[string]map[string]int)
					for pkg := range scores.PackageList {
						linkCount[pkg] = make(map[string]int)
						count := scores.FetchdLinkCountSingle(pkg, repo.URL, db)
						linkCount[pkg][repo.URL] = count
						distro_scores := scores.CalculateDepsdistro(repo.URL, linkCount)
						totalnum += distro_scores
					}
					score := scores.CalculateScore(projectData, totalnum)
					log.Println(score)
				}
			}
			return nil
		}}
	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}
}
