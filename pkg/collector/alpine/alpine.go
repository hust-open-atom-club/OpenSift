package alpine

import (
	"log"
	"strings"

	collector "github.com/HUSTSecLab/criticality_score/pkg/collector/internal"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/HUSTSecLab/criticality_score/pkg/storage/repository"
)

type AlpineCollector struct {
	collector.CollecterInterface
}

func (ac *AlpineCollector) Collect(outputPath string) {
	adc := storage.GetDefaultAppDatabaseContext()
	data := ac.GetPackageInfo(collector.AlpineURL)
	ac.ParseInfo(data)
	ac.GetDep()
	ac.PageRank(0.85, 20)
	ac.GetDepCount()
	ac.UpdateRelationships(adc)
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

func (ac *AlpineCollector) ParseInfo(data string) {
	entries := strings.Split(data, "\n\n")
	for _, entry := range entries {
		lines := strings.Split(entry, "\n")
		var pkg collector.PackageInfo
		for _, line := range lines {
			if len(line) < 2 {
				continue
			}
			switch line[0:2] {
			case "P:":
				pkg.Name = line[2:]
			case "V:":
				pkg.Version = line[2:]
			case "D:":
				depends := strings.Fields(line[2:])
				for _, dep := range depends {
					if idx := strings.Index(dep, ":"); idx != -1 {
						dep = dep[idx+1:]
					}
					if idx := strings.Index(dep, "="); idx != -1 {
						dep = dep[:idx]
					}
					pkg.DirectDepends = append(pkg.DirectDepends, dep)
				}
			case "T:":
				pkg.Description = line[2:]
			case "U:":
				pkg.Homepage = line[2:]
			}
		}
		if pkg.Name != "" {
			ac.SetPkgInfo(pkg.Name, &pkg)
		}
	}
}

func NewAlpineCollector() *AlpineCollector {
	return &AlpineCollector{
		CollecterInterface: collector.NewCollector(repository.Alpine, repository.DistPackageTablePrefix("alpine")),
	}
}
