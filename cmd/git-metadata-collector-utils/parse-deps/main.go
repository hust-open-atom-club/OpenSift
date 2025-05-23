package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	collector "github.com/HUSTSecLab/criticality_score/pkg/gitfile/collector"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/git"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/bytedance/gopkg/util/gopool"
)

func saveResults(pkgInfo, pkgRelationship []string) {
	flag := os.O_WRONLY | os.O_CREATE | os.O_APPEND

	infoFile, err := os.OpenFile("info.csv", flag, 0644)
	if err != nil {
		logger.Fatal(err)
	}
	defer infoFile.Close()

	relationshipFile, err := os.OpenFile("relationship.csv", flag, 0644)
	if err != nil {
		logger.Fatal(err)
	}
	defer relationshipFile.Close()

	joined := strings.Join(pkgInfo, "\n")
	if _, err := infoFile.WriteString(joined); err != nil {
		logger.Fatal(err)
	}

	joined = strings.Join(pkgRelationship, "\n")
	if _, err := relationshipFile.WriteString(joined); err != nil {
		logger.Fatal(err)
	}
}

func readPaths() []string {
	return []string{}
}

func main() {
	paths := readPaths()
	ch := make(chan *git.Repo)
	wg := sync.WaitGroup{}
	wg.Add(len(paths))
	gopool.Go(func() {
		pkgInfo := []string{}
		pkgRelationship := []string{}
		for repo := range ch {
			for pkg, deps := range repo.EcoDeps {
				pkgInfo = append(pkgInfo,
					fmt.Sprintf("%s, %s, %s, %s", repo.URL, pkg.Name, pkg.Version, pkg.Eco),
				)
				for _, dep := range *deps {
					pkgRelationship = append(pkgRelationship,
						fmt.Sprintf("%s, %s, %s, %s, %s", pkg.Name, pkg.Version, dep.Name, dep.Version, pkg.Eco),
					)
				}
			}
		}
		saveResults(pkgInfo, pkgRelationship)
		wg.Done()
	})
	for _, path := range paths {
		gopool.Go(func() {
			r, err := collector.Open(path)
			if err != nil {
				logger.Fatal(path, err)
			}

			repo, err := git.ParseRepo(r)
			if err != nil {
				logger.Fatal(path, err)
			}

			ch <- repo
			wg.Done()
		})
	}
	wg.Wait()
	close(ch)
	wg.Add(1)
	wg.Wait()
}
