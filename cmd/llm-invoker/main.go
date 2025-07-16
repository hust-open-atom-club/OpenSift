package main

import (
	"strings"

	"github.com/HUSTSecLab/OpenSift/pkg/config"
	"github.com/HUSTSecLab/OpenSift/pkg/llm"
	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/spf13/pflag"
)

var choice = pflag.String("choice", "", "which function to run etc. home2git, identifyCountry, identifyIndustry")
var repo = pflag.String("repo", "", "input repo name")
var url = pflag.String("url", "", "input LLM url")
var batchSize = pflag.Int("batch", 1000, "batch size")
var outputCsv = pflag.String("output", "", "output csv file")

func main() {
	config.RegistCommonFlags(pflag.CommandLine)
	config.RegistGithubTokenFlags(pflag.CommandLine)
	config.ParseFlags(pflag.CommandLine)

	pflag.Parse()
	if *choice == "home2git" {
		repolist := strings.Split(*repo, ",")
		llm.Home2git(storage.GetDefaultAppDatabaseContext(), repolist, *url, *batchSize, *outputCsv)
	}
	if *choice == "identifyIndustry" {
		llm.IndustryID(storage.GetDefaultAppDatabaseContext(), *url, *batchSize, *outputCsv)
	}
}
