package main

import (
	"fmt"
	"log"

	"github.com/HUSTSecLab/OpenSift/cmd/archives/package_calculator/internal/package_calculator"
	"github.com/HUSTSecLab/OpenSift/pkg/config"
	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/spf13/pflag"
)

var flagRepoName = pflag.String("repo", "", "name of the repository")
var flagMethod = pflag.String("method", "", "method to use for calculation (bfs or dfs)")

func main() {
	config.RegistCommonFlags(pflag.CommandLine)
	config.ParseFlags(pflag.CommandLine)

	if *flagRepoName == "" {
		log.Fatal("Repository name must be provided")
	}

	if *flagMethod != "bfs" && *flagMethod != "dfs" {
		log.Fatal("Method must be either 'bfs' or 'dfs'")
	}

	db, err := storage.GetDefaultAppDatabaseContext().GetDatabaseConnection()
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
