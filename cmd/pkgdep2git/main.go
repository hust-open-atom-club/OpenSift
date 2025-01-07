package main

import (
	"flag"
	"log"

	"github.com/HUSTSecLab/criticality_score/pkg/pkgdep2git"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
)

var flagConfigPath = flag.String("config", "config.json", "path to the config file")
var batchSize = flag.Int("batch", 1000, "batch size for updating scores")

func main() {
	flag.Parse()
	storage.InitializeDefaultAppDatabase(*flagConfigPath)
	db, err := storage.GetDefaultDatabaseConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer db.Close()

	depMap := pkgdep2git.FetchAlldep(db)
	gitdepMap := pkgdep2git.GenGitDep(db, depMap)
	err = pkgdep2git.BatchUpdate(db, *batchSize, gitdepMap)
	if err != nil {
		log.Fatalf("Error updating database: %v", err)
	}
}
