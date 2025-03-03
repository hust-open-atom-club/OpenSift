package main

import (
	"fmt"

	"github.com/HUSTSecLab/criticality_score/pkg/config"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	scores "github.com/HUSTSecLab/criticality_score/pkg/score"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	_ "github.com/lib/pq"
	"github.com/spf13/pflag"
)

var (
	batchSize     = pflag.Int("batch", 1000, "batch size")
	calcType      = pflag.String("calc", "all", "calculation type: distro, git, langeco, all")
	normalization = pflag.String("normalization", "log", "normalization type: log, sigmoid")
)

func main() {
	config.RegistCommonFlags(pflag.CommandLine)
	config.ParseFlags(pflag.CommandLine)
	ac := storage.GetDefaultAppDatabaseContext()
	scores.UpdatePackageList(ac)
	linksMap := scores.FetchGitLink(ac)
	// linksMap := []string{"https://sourceware.org/git/glibc.git"}
	gitMeticMap := scores.FetchGitMetrics(ac)
	langEcoMetricMap := scores.FetchLangEcoMetadata(ac)
	distMetricMap := scores.FetchDistMetadata(ac)
	var gitMetadataScore = make(map[string]*scores.GitMetadataScore)

	packageScore := make(map[string]*scores.LinkScore)
	round := scores.GetRound(ac)

	for _, link := range linksMap {
		if _, ok := distMetricMap[link]; !ok {
			distMetricMap[link] = scores.NewDistScore()
		}
		distMetricMap[link].CalculateDistScore(*normalization)

		if _, ok := langEcoMetricMap[link]; !ok {
			langEcoMetricMap[link] = scores.NewLangEcoScore()
		}
		langEcoMetricMap[link].CalculateLangEcoScore(*normalization)

		gitMetadataScore[link] = scores.NewGitMetadataScore()
		if _, ok := gitMeticMap[link]; !ok {
			fmt.Println("No git metadata for ", link)
		} else {
			gitMetadataScore[link].CalculateGitMetadataScore(gitMeticMap[link], *normalization)
		}
		packageScore[link] = scores.NewLinkScore(gitMetadataScore[link], distMetricMap[link], langEcoMetricMap[link], round+1)
		packageScore[link].CalculateScore(*normalization)
	}
	logger.Println("Updating database...")
	scores.UpdateScore(ac, packageScore)
}
