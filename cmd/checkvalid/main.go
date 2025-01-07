package main

import (
	"flag"
	"log"

	"github.com/HUSTSecLab/criticality_score/pkg/checkvalid"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
)

var flagConfigPath = flag.String("config", "config.json", "path to the config file")
var outputFile = flag.String("output", "output.csv", "path to the output file")
var checkCloneValid = flag.Bool("checkCloneValid", false, "check clone valid")
var maxThreads = flag.Int("maxThreads", 10, "max threads")

func main() {
	flag.Parse()
	storage.InitializeDefaultAppDatabase(*flagConfigPath)
	db, err := storage.GetDefaultDatabaseConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer db.Close()
	invalidLinks := checkvalid.CheckVaild(db, *checkCloneValid, *maxThreads)
	checkvalid.WriteCsv(invalidLinks, *outputFile)
	log.Println("checkvalid finished")
}
