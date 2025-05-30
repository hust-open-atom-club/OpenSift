package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	collector "github.com/HUSTSecLab/criticality_score/pkg/gitfile/collector"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/git"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/bytedance/gopkg/util/gopool"
)

const (
	BATCH       = 1024
	BUFFER_SIZE = 1024
	POOL_SIZE   = 256
)

func main() {
	ch := make(chan *git.Repo, BUFFER_SIZE)
	wg := sync.WaitGroup{}
	gopool.SetCap(POOL_SIZE)

	inputFile, err := os.OpenFile("input.txt", os.O_RDONLY, 0)
	if err != nil {
		logger.Fatal(err)
	}
	defer inputFile.Close()

	FLAG := os.O_WRONLY | os.O_CREATE | os.O_APPEND

	infoFile, err := os.OpenFile("info.csv", FLAG, 0644)
	if err != nil {
		logger.Fatal(err)
	}
	defer infoFile.Close()

	relationshipFile, err := os.OpenFile("relationship.csv", FLAG, 0644)
	if err != nil {
		logger.Fatal(err)
	}
	defer relationshipFile.Close()

	reader := bufio.NewReader(inputFile)
	//infoWriter := bufio.NewWriter(infoFile)
	relationshipWriter := bufio.NewWriter(relationshipFile)

	gopool.Go(func() {
		pkgInfo := []string{}
		pkgRelationship := []string{}

		for repo := range ch {
			if len(repo.Ecosystems) == 0 {
				continue
			}

			for pkg, deps := range repo.EcoDeps {
				info := fmt.Sprintf("%v, %v, %v, %v", repo.URL, pkg.Name, pkg.Version, pkg.Eco)
				pkgInfo = append(pkgInfo, info)
				if deps == nil {
					continue
				}
				for _, dep := range *deps {
					relation := fmt.Sprintf("%v, %v, %v, %v, %v", pkg.Name, pkg.Version, dep.Name, dep.Version, pkg.Eco)
					pkgRelationship = append(pkgRelationship, relation)
				}
			}
			if len(pkgInfo) > BATCH {
				logger.Info("Save pkg Info")
				joined := strings.Join(pkgInfo, "\n")
				if _, err := infoFile.WriteString(joined); err != nil {
					logger.Fatal(err)
				}
				pkgInfo = []string{}
			}
			if len(pkgRelationship) > BATCH {
				logger.Info("Save pkg Relationship")
				joined := strings.Join(pkgRelationship, "\n")
				if _, err := relationshipWriter.WriteString(joined); err != nil {
					logger.Fatal(err)
				}
				pkgRelationship = []string{}
			}
		}
		if len(pkgInfo) >= 0 {
			logger.Info("Save pkg Info")

			joined := strings.Join(pkgInfo, "\n")
			if _, err := infoFile.WriteString(joined); err != nil {
				logger.Fatal(err)
			}
		}
		if len(pkgRelationship) > 0 {
			logger.Info("Save pkg Relationship")

			joined := strings.Join(pkgRelationship, "\n")
			if _, err := relationshipWriter.WriteString(joined); err != nil {
				logger.Fatal(err)
			}
		}
		wg.Done()
	})

	for {
		var path string
		if lineData, err := reader.ReadString('\n'); err != nil {
			break
		} else {
			wg.Add(1)
			path = strings.TrimRight(lineData, "\n")
		}

		gopool.Go(func() {
			r, err := collector.Open(path)
			if err != nil {
				return
				// logger.Fatal(path, err)
			}

			repo, err := git.ParseRepo(r)
			if err != nil {
				return
				// logger.Fatal(path, err)
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
