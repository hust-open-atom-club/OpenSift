package openkylin

import (
	"log"
	"regexp"
	"strings"

	collector "github.com/HUSTSecLab/criticality_score/pkg/collector/internal"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/HUSTSecLab/criticality_score/pkg/storage/repository"
)

type OpenKylinCollector struct {
	collector.CollecterInterface
}

func (dc *OpenKylinCollector) Collect(outputPath string) {
	adc := storage.GetDefaultAppDatabaseContext()
	data := dc.GetPackageInfo(collector.OpenKylinURL)
	dc.ParseInfo(data)
	dc.GetDep()
	dc.PageRank(0.85, 20)
	dc.GetDepCount()
	dc.UpdateDistRepoCount(adc)
	dc.CalculateDistImpact()
	dc.UpdateOrInsertDatabase(adc)
	dc.UpdateOrInsertDistDependencyDatabase(adc)
	if outputPath != "" {
		err := dc.GenerateDependencyGraph(outputPath)
		if err != nil {
			log.Printf("Error generating dependency graph: %v\n", err)
			return
		}
	}
}

func (dc *OpenKylinCollector) ParseInfo(data string) {
	var currentPkg *collector.PackageInfo
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "Package:"):
			if currentPkg != nil {
				dc.SetPkgInfo(currentPkg.Name, currentPkg)
			}
			currentPkg = &collector.PackageInfo{Name: strings.TrimSpace(strings.Split(line, ":")[1])}
		case strings.Contains(line, "Version:"):
			currentPkg.Version = strings.TrimSpace(strings.Split(line, ":")[1])
		case strings.Contains(line, "Description:"):
			currentPkg.Description = strings.TrimSpace(strings.Split(line, ":")[1])
		case strings.Contains(line, "Homepage:"):
			if strings.Contains(line, "<insert the upstream URL, if relevant>") {
				currentPkg.Homepage = ""
			} else {
				currentPkg.Homepage = strings.TrimSpace(strings.Split(line, ":")[1] + ":" + strings.Split(line, ":")[2])
			}
		case strings.Contains(line, "Depends:"):
			depLine := strings.TrimPrefix(line, "Depends: ")
			re := regexp.MustCompile(`[\w\-\.\+\|]+(?:\s*\([^)]+\))?`)
			matches := re.FindAllString(depLine, -1)

			var cleanedDeps []string
			for _, match := range matches {
				for _, subPkg := range strings.Split(match, "|") {
					subPkg = strings.TrimSpace(subPkg)
					if idx := strings.Index(subPkg, " ("); idx != -1 {
						subPkg = subPkg[:idx]
					}
					cleanedDeps = append(cleanedDeps, subPkg)
				}
			}
			currentPkg.DirectDepends = cleanedDeps
		}
	}
	if currentPkg != nil {
		dc.SetPkgInfo(currentPkg.Name, currentPkg)
	}
}

func NewOpenKylinCollector() *OpenKylinCollector {
	return &OpenKylinCollector{
		CollecterInterface: collector.NewCollector(repository.OpenKylin, repository.DistPackageTablePrefix("openkylin")),
	}
}
