package homebrew

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	collector "github.com/HUSTSecLab/OpenSift/pkg/collector/internal"
	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/HUSTSecLab/OpenSift/pkg/storage/repository"
)

type HomebrewCollector struct {
	collector.CollecterInterface
}

func (hc *HomebrewCollector) Collect(outputPath string, downloadDir string) {
	adc := storage.GetDefaultAppDatabaseContext()
	err := hc.CloneHomebrewRepo(downloadDir)
	if err != nil {
		log.Printf("Error cloning homebrew repository: %v\n", err)
		return
	}
	hc.ParseInfo(downloadDir)
	hc.GetDep()
	hc.PageRank(0.85, 20)
	hc.GetDepCount()
	hc.UpdateRelationships(adc)
	hc.UpdateDistRepoCount(adc)
	hc.CalculateDistImpact()
	hc.UpdateOrInsertDatabase(adc)
	hc.UpdateOrInsertDistDependencyDatabase(adc)
	if outputPath != "" {
		err = hc.GenerateDependencyGraph(outputPath)
		if err != nil {
			log.Printf("Error generating dependency graph: %v\n", err)
			return
		}
	}
}

func (hc *HomebrewCollector) CloneHomebrewRepo(dir string) error {
	repoURL := "https://github.com/Homebrew/homebrew-core.git"

	if _, err := os.Stat(dir); err == nil {
		cmd := exec.Command("git", "-C", dir, "pull")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to pull repository: %v", err)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check directory: %v", err)
	}

	cmd := exec.Command("git", "clone", repoURL, dir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %v", err)
	}

	return nil
}

func (hc *HomebrewCollector) ParseInfo(dir string) error {

	formulaDir := filepath.Join(dir, "Formula")

	err := filepath.Walk(formulaDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".rb") {
			formulaContent, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read formula file %s: %v", path, err)
			}

			pkgInfo := hc.parseFormulaContent(string(formulaContent), info.Name())
			if pkgInfo.Name != "" {
				hc.SetPkgInfo(pkgInfo.Name, &pkgInfo)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk through formula directory: %v", err)
	}

	return nil
}

func (hc *HomebrewCollector) parseFormulaContent(content string, fileName string) collector.PackageInfo {
	var pkgInfo collector.PackageInfo

	baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	pkgInfo.Name = baseName

	descRe := regexp.MustCompile(`desc\s+"([^"]+)"`)
	if match := descRe.FindStringSubmatch(content); len(match) > 1 {
		pkgInfo.Description = match[1]
	}

	homepageRe := regexp.MustCompile(`homepage\s+"([^"]+)"`)
	if match := homepageRe.FindStringSubmatch(content); len(match) > 1 {
		pkgInfo.Homepage = match[1]
	}

	dependsRe := regexp.MustCompile(`depends_on\s+"([^"]+)"`)
	dependsMatches := dependsRe.FindAllStringSubmatch(content, -1)
	for _, match := range dependsMatches {
		if len(match) > 1 {
			pkgInfo.DirectDepends = append(pkgInfo.DirectDepends, match[1])
		}
	}

	return pkgInfo
}

func NewHomebrewCollector() *HomebrewCollector {
	return &HomebrewCollector{
		CollecterInterface: collector.NewCollector(repository.Homebrew, repository.DistPackageTablePrefix("homebrew")),
	}
}
