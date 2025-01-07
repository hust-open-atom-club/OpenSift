package main

import (
	"flag"
	"log"

	"github.com/HUSTSecLab/criticality_score/pkg/gitmetricsync"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
)

var flagConfigPath = flag.String("config", "config.json", "path to the config file")
var batchSize = flag.Int("batch", 1000, "batch size")

func main() {
	flag.Parse()
	storage.InitializeDefaultAppDatabase(*flagConfigPath)

	log.Println("Starting synchronization...")
	gitmetricsync.Run()
	log.Println("Synchronization complete.")
	gitmetricsync.Union_repo(*batchSize)
}
