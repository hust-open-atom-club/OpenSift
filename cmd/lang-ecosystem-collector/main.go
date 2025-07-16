package main

import (
	"github.com/HUSTSecLab/OpenSift/pkg/config"
	"github.com/HUSTSecLab/OpenSift/pkg/depsdev"
	"github.com/spf13/pflag"
)

var (
	flagBatchSize     = pflag.Int("batch", 100, "batch size")
	workerCount       = pflag.Int("workers", 50, "number of workers")
	calculatePageRank = pflag.Bool("pagerank", false, "calculate page rank")
	debugMode         = pflag.Bool("debug", false, "debug mode")
)

func main() {
	config.RegistCommonFlags(pflag.CommandLine)
	config.ParseFlags(pflag.CommandLine)

	depsdev.Depsdev(*flagBatchSize, *workerCount, *calculatePageRank, *debugMode)
}
