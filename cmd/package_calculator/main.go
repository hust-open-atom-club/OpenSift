package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/HUSTSecLab/criticality_score/pkg/package_calculator"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
)

var flagConfigPath = flag.String("config", "config.json", "path to the config file")
var flagRepoName = flag.String("repo", "", "name of the repository")
var flagMethod = flag.String("method", "", "method to use for calculation (bfs or dfs)")

func main() {
	flag.Parse()

	if *flagRepoName == "" {
		log.Fatal("Repository name must be provided")
	}

	if *flagMethod != "bfs" && *flagMethod != "dfs" {
		log.Fatal("Method must be either 'bfs' or 'dfs'")
	}

	storage.InitializeDefaultAppDatabase(*flagConfigPath)
	db, err := storage.GetDefaultDatabaseConnection()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	Countquery := fmt.Sprintf("SELECT count(*) FROM %s_packages", *flagRepoName)
	var count int
	err = db.QueryRow(Countquery).Scan(&count)
	if err != nil {
		log.Fatalf("Error querying database: %v", err)
	}

	query := fmt.Sprintf("SELECT frompackage, topackage FROM %s_relationships", *flagRepoName)
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Error querying database: %v", err)
	}
	defer rows.Close()

	if err := package_calculator.CalculatePackages(rows, *flagMethod, count); err != nil {
		log.Fatalf("Error calculating packages: %v", err)
	}
}
