// This file is used to clone the git repo to memory, and
// print its metadata to the console.
package main

import (
	"os"
	"strings"
	"sync"

	"github.com/HUSTSecLab/criticality_score/pkg/config"
	collector "github.com/HUSTSecLab/criticality_score/pkg/gitfile/collector"
	git "github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/git"
	url "github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/url"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/bytedance/gopkg/util/gopool"
	gogit "github.com/go-git/go-git/v5"
	"github.com/spf13/pflag"
)

func main() {
	logger.ConfigAsCommandLineTool()
	config.RegistGitStorageFlags(pflag.CommandLine)
	inMemoryFlag := pflag.BoolP("memory", "m", false, "Whether to clone the repository in memory")

	pflag.Usage = func() {
		logger.Printf("This tool is used to collect git metadata in storage path, but not clone the repository.\n")
		logger.Printf("Usage: %s [url1] [url2] ...\n", os.Args[0])
		pflag.PrintDefaults()
	}
	pflag.Parse()

	if pflag.NArg() == 0 {
		pflag.Usage()
		os.Exit(1)
	}

	inputs := []string{}
	for i := 0; i < pflag.NArg(); i++ {
		inputs = append(inputs, pflag.Arg(i))
	}

	var wg sync.WaitGroup
	wg.Add(len(inputs))

	repos := make([]*git.Repo, 0)

	for _, input := range inputs {
		gopool.Go(func() {
			defer wg.Done()
			logger.Infof("%s: Start collecting", input)

			r := &gogit.Repository{}
			var err error

			if !strings.Contains(input, "://") {
				if *inMemoryFlag {
					logger.Warnf("%s: In memory flag is set, but the input is not a URL", input)
				}
				r, err = collector.Open(input)
				if err != nil {
					logger.Errorf("%s: Opening failed: %s", input, err)
					return
				}
				err = collector.Pull(r, "")
			} else if *inMemoryFlag {
				u, _ := url.ParseURL(input)
				r, err = collector.EzCollect(&u)
			} else {
				if config.GetGitStoragePath() == "" {
					logger.Errorf("Storage path is not set")
					return
				}
				u, _ := url.ParseURL(input)
				r, err = collector.Collect(&u, config.GetGitStoragePath())
			}

			if err != nil {
				logger.Errorf("%s: Collect failed: %s", input, err)
				return
			}

			repo, err := git.ParseRepo(r)
			if err != nil {
				logger.Errorf("%s: Parsing failed: %s", input, err)
				return
			}

			repos = append(repos, repo)
			logger.Infof("%s Collected", repo.Name)
		})
	}

	wg.Wait()
	for _, repo := range repos {
		repo.Show()
	}
}
