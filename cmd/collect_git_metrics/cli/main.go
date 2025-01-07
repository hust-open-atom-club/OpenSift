/*
 * @Date: 2024-09-06 21:09:14
 * @LastEditTime: 2025-01-07 19:12:42
 * @Description: The Cli for collector
 */
package main

import (
	"os"
	"strings"
	"sync"

	collector "github.com/HUSTSecLab/criticality_score/pkg/collector_git/collector"
	"github.com/HUSTSecLab/criticality_score/pkg/collector_git/logger"
	git "github.com/HUSTSecLab/criticality_score/pkg/collector_git/parser/git"
	url "github.com/HUSTSecLab/criticality_score/pkg/collector_git/parser/url"
	"github.com/bytedance/gopkg/util/gopool"
	gogit "github.com/go-git/go-git/v5"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "grm-collector",
		Usage: "Collect Git-based Repository Metrics",
		Action: func(c *cli.Context) error {
			inputs := []string{}
			for i := 0; i < c.NArg(); i++ {
				inputs = append(inputs, c.Args().Get(i))
			}

			var wg sync.WaitGroup
			wg.Add(len(inputs))

			repos := make([]*git.Repo, 0)

			for index, input := range inputs {
				gopool.Go(func() {
					defer wg.Done()
					logger.Infof("[%d] Collecting %s", index, input)

					r := &gogit.Repository{}
					var err error

					//* if the input is url, parse and clone the repo
					//* if not, open the repo
					if strings.Contains(input, "://") {
						u := url.ParseURL(input)
						r, err = collector.EzCollect(&u)
						if err != nil {
							logger.Panicf("[%d] Collecting %s Failed", index, input)
						}
					} else {
						r, err = collector.Open(input)
						if err != nil {
							logger.Panicf("[%d] Opening %s Failed", index, input)
						}
					}

					repo, err := git.ParseRepo(r)
					if err != nil {
						logger.Panicf("[%d] Parsing %s Failed", index, input)
					}

					repos = append(repos, repo)
					logger.Infof("[%d] %s Collected", index, repo.Name)
				})
			}

			wg.Wait()
			for _, repo := range repos {
				repo.Show()
			}
			return nil
		}}
	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}
}
