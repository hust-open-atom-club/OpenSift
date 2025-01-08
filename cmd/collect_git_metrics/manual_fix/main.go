/*
 * @Date: 2024-09-06 21:09:14
 * @LastEditTime: 2024-12-16 18:58:31
 * @Description: The Cli for collector
 */
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	collector "github.com/HUSTSecLab/criticality_score/internal/collector"
	git "github.com/HUSTSecLab/criticality_score/internal/parser/git"
	url "github.com/HUSTSecLab/criticality_score/internal/parser/url"
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
			storage.InitializeDefaultAppDatabase(configPath)
			db, err := storage.GetDefaultDatabaseConnection()
			if err != nil {
				log.Fatalf("Failed to connect to database: %v", err)
			}

			defer db.Close()
			for _, repo := range repos {
				repo.Show()
				depsdev_count := FetchDepsdev(db, repo.URL)
				projectData := scores.ProjectData{
					CommitFrequency:  &repo.CommitFrequency,
					ContributorCount: &repo.ContributorCount,
					CreatedSince:     &repo.CreatedSince,
					UpdatedSince:     &repo.UpdatedSince,
					Org_Count:        &repo.OrgCount,
					Pkg_Manager:      &repo.Ecosystems,
					DepsdevCount:     &depsdev_count,
				}
				linkCount := make(map[string]map[string]int)
				scores.CalculaterepoCount(db)
				for pkg := range scores.PackageList {
					linkCount[pkg] = make(map[string]int)
					count := scores.FetchdLinkCountSingle(pkg, repo.URL, db)
					linkCount[pkg][strings.ToLower(repo.URL)] = count
				}
				deps_distro := scores.CalculateDepsdistro(repo.URL, linkCount)
				score := scores.CalculateScore(projectData, deps_distro)
				if updateDB {
					err = updateGitMetrics(db, repo, score, deps_distro)
					if err != nil {
						logger.Fatal(err)
					}
				}
			}
			return nil
		}}
	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}
}
func updateGitMetrics(db *sql.DB, repo *git.Repo, score float64, depsDistro float64) error {
	query := `
		UPDATE git_metrics 
		SET created_since = $1, updated_since = $2, contributor_count = $3, commit_frequency = $4, org_count = $5, scores = $6, deps_distro = $7
		WHERE git_link = $8
	`
	_, err := db.Exec(query, repo.CreatedSince, repo.UpdatedSince, repo.ContributorCount, repo.CommitFrequency, repo.OrgCount, score, depsDistro, repo.URL)
	if err != nil {
		return err
	}
	return nil
}

func FetchDepsdev(db *sql.DB, git_link string) int {
	query := fmt.Sprintf("SELECT depsdev_count FROM git_metrics WHERE git_link = '%s'", git_link)
	var depsdev_count int
	err := db.QueryRow(query).Scan(&depsdev_count)
	if err != nil {
		log.Fatal(err)
	}
	return depsdev_count
}
