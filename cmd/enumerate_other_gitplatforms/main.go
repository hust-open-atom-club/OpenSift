package main

import (
	"flag"
	"strings"

	"github.com/HUSTSecLab/criticality_score/pkg/enumerator/enumerator"
	"github.com/HUSTSecLab/criticality_score/pkg/enumerator/writer"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
)

func printGithubUsage() {
	logger.Warn("NOTE: This util is not support for github, please use `enumerate_github` instead.")
}

func main() {
	// flags
	var (
		flagConfig      = flag.String("config", "", "path to the configuration file, when output type is db")
		flagPlatforms   = flag.String("platforms", "", "comma separated list of platforms to enumerate")
		flagOutputType  = flag.String("output", "stdout", "output type: allow stdout, file, db")
		flagOutputFilev = flag.String("output-file", "", "output file")
		flagJobs        = flag.Int("jobs", 10, "number of concurrent jobs")
		flagTake        = flag.Int("take", 1000, "number of repositories to enumerate")
	)

	flag.Parse()
	configPath := *flagConfig
	platforms := strings.Split(*flagPlatforms, ",")

	for _, platform := range platforms {
		var w writer.Writer
		var tableName string
		var en enumerator.Enumerator

		switch platform {
		case "github":
			// NOTE: github is not supported in this util
			printGithubUsage()
		case "gitlab":
			tableName = "gitlab_links"
			en = enumerator.NewGitlabEnumerator(*flagTake, *flagJobs)
		case "bitbucket":
			tableName = "bitbucket_links"
			en = enumerator.NewBitBucketEnumerator(*flagTake)
		default:
			panic("unknown platform")
		}

		switch *flagOutputType {
		case "stdout":
			w = writer.NewStdOutWriter()
		case "file":
			w = writer.NewTextFileWriter(*flagOutputFilev)
		case "db":
			w = writer.NewDatabaseWriter(*&configPath, tableName)
		default:
			panic("unknown output type")
		}

		en.SetWriter(w)

		err := en.Enumerate()

		if err != nil {
			logger.Errorf("failed to enumerate %s: %v", platform, err)
		}

	}

}
