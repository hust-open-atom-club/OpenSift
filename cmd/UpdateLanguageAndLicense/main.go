package main

import (
	"flag"
	"fmt"
	"log"
	"sync"

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

	gopool.SetCap(int32(*flagJobsCount))

	for _, input := range urls {

		gopool.Go(func() {
			defer wg.Done()
			u := url.ParseURL(input)

			path := fmt.Sprintf("%s/%s%s", config.STORAGE_PATH, u.Resource, u.Pathname)
			r, err := collector.Open(path)

			if err != nil || r == nil {
				logger.Errorf("Open %s failed: %s", u.URL, err)
				return
			}

			result := git.NewRepo()
			err = result.WalkRepo(r)

			if err != nil {
				logger.Errorf("WalkRepo %s failed: %s", input, err)
				return
			}

			sqlResult, err := db.Exec(`UPDATE git_metrics SET
				ecosystem = $1,
				license = $2,
				language = $3
				WHERE git_link = $4`,
				result.Ecosystems,
				result.License,
				result.Languages,
				input)

			if err != nil {
				logger.Errorf("Update database for %s failed: %v", input, err)
				return
			}

			rowAffected, err := sqlResult.RowsAffected()

			if err != nil {
				logger.Errorf("Get RowsAffected for %s Failed: %v", input, err)
				return
			}

			if rowAffected == 0 {
				logger.Warnf("Update %s failed: row affected = 0", input)
				return
			}

			logger.Infof("Success: %s", input)

		})
	}
	wg.Wait()
}
