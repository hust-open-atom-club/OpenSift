package archlinux

import (
	"log"
	"strings"

	collector "github.com/HUSTSecLab/criticality_score/pkg/collector/internal"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/HUSTSecLab/criticality_score/pkg/storage/repository"
)

type ArchLinuxCollector struct {
	collector.CollecterInterface
}

func (al *ArchLinuxCollector) Collect(outputPath string) {
	adc := storage.GetDefaultAppDatabaseContext()
	data := al.GetPackageInfo(collector.ArchlinuxURL)
	al.ParseInfo(data)
	al.GetDep()
	al.PageRank(0.85, 20)
	al.GetDepCount()
	al.UpdateRelationships(adc)
	al.UpdateDistRepoCount(adc)
	al.CalculateDistImpact()
	al.UpdateOrInsertDatabase(adc)
	al.UpdateOrInsertDistDependencyDatabase(adc)
	if outputPath != "" {
		err := al.GenerateDependencyGraph(outputPath)
		if err != nil {
			log.Printf("Error generating dependency graph: %v\n", err)
			return
		}
	}
}

func (al *ArchLinuxCollector) ParseInfo(data string) {
	var currentPkg *collector.PackageInfo
	var depend bool

	lines := strings.Split(data, "\n")
	for idx, line := range lines {
		switch {
		case line == "%NAME%":
			if currentPkg != nil {
				al.SetPkgInfo(currentPkg.Name, currentPkg)
			}
			currentPkg = &collector.PackageInfo{Name: strings.TrimSpace(lines[idx+1])}
		case line == "%DESC%":
			currentPkg.Description = strings.TrimSpace(lines[idx+1])
		case line == "%VERSION%":
			currentPkg.Version = strings.TrimSpace(lines[idx+1])
		case line == "%URL%":
			currentPkg.Homepage = strings.TrimSpace(lines[idx+1])
		case line == "%DEPENDS%":
			depend = true
		case depend && (strings.Contains(line, "%") && line != "%DEPENDS%"):
			depend = false
		case depend && line != "":
			currentPkg.DirectDepends = append(currentPkg.DirectDepends, strings.TrimSpace(line))
		}
	}
	if currentPkg != nil {
		al.SetPkgInfo(currentPkg.Name, currentPkg)
	}
}

func NewArchLinuxCollector() *ArchLinuxCollector {
	return &ArchLinuxCollector{
		collector.NewCollector(repository.Arch, repository.DistPackageTablePrefix("arch")),
	}
}
