package main

import (
	"flag"

	"github.com/HUSTSecLab/criticality_score/pkg/collector_depsdev"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
)

var flagConfigPath = flag.String("config", "config.json", "path to the config file")
var flagBatchSize = flag.Int("batch", 100, "batch size")
var workerCount = flag.Int("workers", 10, "number of workers")

func main() {
	flag.Parse()
	storage.InitializeDefaultAppDatabase(*flagConfigPath)

	collector_depsdev.Depsdev(*flagConfigPath, *flagBatchSize, *workerCount)
}
