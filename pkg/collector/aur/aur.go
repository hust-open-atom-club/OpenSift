package aur

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	collector "github.com/HUSTSecLab/OpenSift/pkg/collector/internal"
	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/HUSTSecLab/OpenSift/pkg/storage/repository"
)

type AurCollector struct {
	collector.CollecterInterface
}

func (ac *AurCollector) Collect(outputPath string) {
	adc := storage.GetDefaultAppDatabaseContext()
	data := ac.GetPackageInfo(collector.AurURL)
	ac.ParseInfo(data)
	ac.GetDep()
	ac.PageRank(0.85, 20)
	ac.GetDepCount()
	ac.UpdateDistRepoCount(adc)
	ac.CalculateDistImpact()
	ac.UpdateOrInsertDatabase(adc)
	ac.UpdateOrInsertDistDependencyDatabase(adc)
	if outputPath != "" {
		err := ac.GenerateDependencyGraph(outputPath)
		if err != nil {
			log.Printf("Error generating dependency graph: %v\n", err)
			return
		}
	}
}

func (ac *AurCollector) ParseInfo(data string) error {
	var packages []collector.PackageInfo
	err := json.Unmarshal([]byte(data), &packages)
	if err != nil {
		return err
	}

	for _, pkg := range packages {
		ac.SetPkgInfo(pkg.Name, &pkg)
	}
	return nil
}

func (ac *AurCollector) GetPackageInfo(urls collector.PackageURL) string {
	resp, err := http.Get(urls[0])
	if err != nil {
		log.Printf("Error making HTTP request: %v\n", err)
	}
	defer resp.Body.Close()

	var result strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		result.WriteString(scanner.Text() + "\n")
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading response body: %v\n", err)
	}
	return result.String()
}

func NewAurCollector() *AurCollector {
	return &AurCollector{
		collector.NewCollector(repository.Aur, repository.DistPackageTablePrefix("aur")),
	}
}
