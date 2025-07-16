package gentoo

import (
	"bufio"
	"fmt"
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

type GentooCollector struct {
	collector.CollecterInterface
}

func (hc *GentooCollector) Collect(outputPath string, downloadDir string) {
	adc := storage.GetDefaultAppDatabaseContext()
	err := hc.cloneGentooRepo(downloadDir)
	if err != nil {
		log.Printf("Error cloning Gentoo repository: %v\n", err)
		return
	}
	hc.ParseInfo(outputPath)
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

func extractNameAndVersion(fileName string) (string, string) {
	lastDashIndex := strings.LastIndex(fileName, "-")
	if lastDashIndex != -1 {
		versionPart := fileName[lastDashIndex+1:]
		if strings.HasPrefix(versionPart, "r") {
			secondLastDashIndex := strings.LastIndex(fileName[:lastDashIndex], "-")
			if secondLastDashIndex != -1 {
				name := fileName[:secondLastDashIndex]
				version := strings.TrimSuffix(fileName[secondLastDashIndex+1:], ".ebuild")
				return name, version
			}
		} else {
			name := fileName[:lastDashIndex]
			versionParts := strings.Split(fileName[lastDashIndex+1:], ".")
			if len(versionParts) > 0 {
				version := strings.TrimSuffix(versionParts[0], ".ebuild")
				return name, version
			}
		}
	}
	return "", ""
}

func (hc *GentooCollector) ParseEbuild(filePath string) (collector.PackageInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return collector.PackageInfo{}, fmt.Errorf("Error opening ebuild file: %v", err)
	}
	defer file.Close()

	var pkgInfo collector.PackageInfo
	scanner := bufio.NewScanner(file)

	fileName := filepath.Base(filePath)
	pkgInfo.Name, _ = extractNameAndVersion(fileName)

	reDescription := regexp.MustCompile(`^DESCRIPTION="(.+)"$`)
	reHomepage := regexp.MustCompile(`^HOMEPAGE="(.+)"$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if reDescription.MatchString(line) {
			matches := reDescription.FindStringSubmatch(line)
			pkgInfo.Description = matches[1]
		} else if reHomepage.MatchString(line) {
			matches := reHomepage.FindStringSubmatch(line)
			pkgInfo.Homepage = matches[1]
		}
	}

	if err := scanner.Err(); err != nil {
		return collector.PackageInfo{}, fmt.Errorf("Error reading ebuild file: %v", err)
	}
	dependencies, err := getDependenciesFromCommand(pkgInfo.Name)
	if err != nil {
		pkgInfo.DirectDepends = nil
	} else {
		pkgInfo.DirectDepends = dependencies
	}

	return pkgInfo, nil
}

func getDependenciesFromCommand(pkgName string) ([]string, error) {
	cmd := exec.Command("equery", "depgraph", "--depth=1", pkgName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Error executing command: %v", err)
	}

	var dependencies []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			if strings.Contains(line, "]") {
				parts := strings.SplitN(line, "]", 2)
				if len(parts) > 1 {
					dependency := strings.TrimSpace(parts[1])
					if lastSlashIndex := strings.LastIndex(dependency, "/"); lastSlashIndex != -1 {
						dependency = dependency[lastSlashIndex+1:]
					}
					name, _ := extractNameAndVersion(dependency)
					dependencies = append(dependencies, name)
				}
			}
		}
	}
	return dependencies, nil
}

func (hc *GentooCollector) FetchAndParseEbuildFiles(directory string) error {
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(info.Name(), ".ebuild") {
			pkgInfo, err := hc.ParseEbuild(path)
			if err != nil {
				return fmt.Errorf("failed to parse ebuild file %s: %v", path, err)
			}
			if pkgInfo.Name != "" {
				hc.SetPkgInfo(pkgInfo.Name, &pkgInfo)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk through ebuild directory: %v", err)
	}

	return nil
}

func (hc *GentooCollector) cloneGentooRepo(baseDirectory string) error {
	repoURL := "https://github.com/gentoo/gentoo.git"
	dir := filepath.Join(baseDirectory)

	if _, err := os.Stat(dir); err == nil {
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

func (hc *GentooCollector) ParseInfo(outputPath string) {
	cmd := exec.Command("emerge", "--sync")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error executing emerge --sync: %v\n", err)
	}

	err := hc.FetchAndParseEbuildFiles(outputPath)
	if err != nil {
		fmt.Printf("Error fetching package info: %v\n", err)
		return
	}

	fmt.Println("Fetched and parsed ebuild files successfully.")
}

func NewGentooCollector() *GentooCollector {
	return &GentooCollector{
		CollecterInterface: collector.NewCollector(repository.Gentoo, repository.DistPackageTablePrefix("gentoo")),
	}
}
