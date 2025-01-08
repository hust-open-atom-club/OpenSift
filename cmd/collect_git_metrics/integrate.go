/*
 * @Date: 2023-11-11 22:44:26
 * @LastEditTime: 2025-01-07 19:15:24
 * @Description: Integrate into Criticality Score system
 */
package main

import (
	"flag"
	"log"
	"sync"
	"time"

	"github.com/HUSTSecLab/criticality_score/pkg/collector_git/collector"
	"github.com/HUSTSecLab/criticality_score/pkg/collector_git/config"
	git "github.com/HUSTSecLab/criticality_score/pkg/collector_git/parser/git"
	url "github.com/HUSTSecLab/criticality_score/pkg/collector_git/parser/url"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/bytedance/gopkg/util/gopool"
)

var flagConfigPath = flag.String("config", "config.json", "path to the config file")
var flagStoragePath = flag.String("storage", "./storage", "path to git storage location")
var flagJobsCount = flag.Int("jobs", 256, "jobs count")
var flagForceUpdateAll = flag.Bool("force-update-all", false, "force update all repositories")

func getUrls() ([]string, error) {
	conn, err := storage.GetDefaultDatabaseConnection()
	if err != nil {
		return nil, err
	}

	var sqlStatement string

	if *flagForceUpdateAll {
		sqlStatement = `SELECT git_link from git_metrics`
	} else {
		sqlStatement = `SELECT git_link from git_metrics where need_update = true`
	}

	rows, err := conn.Query(sqlStatement)
	if err != nil {
		return nil, err
	}
	var ret []string
	for rows.Next() {
		var link string
		rows.Scan(&link)
		ret = append(ret, link)
	}
	return ret, nil
}

func main() {
	flag.Parse()
	storage.InitializeDefaultAppDatabase(*flagConfigPath)
	config.SetStoragetPath(*flagStoragePath)

	urls, err := getUrls()
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	logger.Infof("%d urls in total", len(urls))
	wg.Add(len(urls))

	db, err := storage.GetDefaultDatabaseConnection()
	if err != nil {
		logger.Fatal("Connecting Database Failed")
	}
	// psql.CreateTable(db)
	gopool.SetCap(int32(*flagJobsCount))

	for index, input := range urls {
		if index%10 == 0 {
			time.Sleep(5 * time.Second)
		} else {
			time.Sleep(2 * time.Second)
		}

		gopool.Go(func() {
			defer wg.Done()
			u := url.ParseURL(input)
			r, err := collector.Collect(&u)
			if err != nil {
				logger.Panicf("Collecting %s Failed", u.URL)
			}
			logger.Infof("[*] %s Collected", input)

			repo, err := git.ParseRepo(r)
			if err != nil {
				logger.Panicf("Parsing %s Failed", input)
			}

			result, err := db.Exec(`UPDATE git_metrics SET
				_name = $1,
				_owner = $2,
				_source = $3,
				ecosystem = $4,
				created_since = $5,
				updated_since = $6,
				contributor_count = $7,
				commit_frequency = $8,
				license = $9,
				language = $10,
				need_update = FALSE WHERE git_link = $11`,
				repo.Name,
				repo.Owner,
				repo.Source,
				repo.Ecosystems,
				repo.CreatedSince,
				repo.UpdatedSince,
				repo.ContributorCount,
				repo.CommitFrequency,
				repo.License,
				repo.Languages,
				input)

			if err != nil {
				logger.Errorf("Update database for %s Failed: %v", input, err)
				return
			}

			rowAffected, err := result.RowsAffected()

			if err != nil {
				logger.Errorf("Get RowsAffected for %s Failed: %v", input, err)
				return
			}

			if rowAffected == 0 {
				logger.Errorf("Update %s Failed", input)
			}
		})
	}
	wg.Wait()
}
