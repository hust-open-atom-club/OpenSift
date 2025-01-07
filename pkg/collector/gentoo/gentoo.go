package gentoo

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/lib/pq"
)

type PackageInfo struct {
	Name         string
	Version      string
	Description  string
	Homepage     string
	Depends      []string
	DependsCount int
	URL          string
	GitRepo      string
	PageRank     float64
}

func storeDependenciesInDatabase(pkgName string, dependencies []string) error {
	db, err := storage.GetDefaultDatabaseConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	for _, dep := range dependencies {
		_, err := db.Exec("INSERT INTO gentoo_relationships (frompackage, topackage) VALUES ($1, $2)", pkgName, dep)
		if err != nil {
			return err
		}
	}
	return nil
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

func ParseEbuild(filePath string) (PackageInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return PackageInfo{}, fmt.Errorf("Error opening ebuild file: %v", err)
	}
	defer file.Close()

	var pkgInfo PackageInfo
	scanner := bufio.NewScanner(file)

	fileName := filepath.Base(filePath)
	pkgInfo.Name, pkgInfo.Version = extractNameAndVersion(fileName)

	reDescription := regexp.MustCompile(`^DESCRIPTION="(.+)"$`)
	reHomepage := regexp.MustCompile(`^HOMEPAGE="(.+)"$`)
	reURL := regexp.MustCompile(`^SRC_URI="(.+)"$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if reDescription.MatchString(line) {
			matches := reDescription.FindStringSubmatch(line)
			pkgInfo.Description = matches[1]
		} else if reHomepage.MatchString(line) {
			matches := reHomepage.FindStringSubmatch(line)
			pkgInfo.Homepage = matches[1]
		} else if reURL.MatchString(line) {
			matches := reURL.FindStringSubmatch(line)
			pkgInfo.URL = matches[1]
			if strings.HasPrefix(pkgInfo.URL, "https://github.com/") {
				parts := strings.Split(pkgInfo.URL, "/")
				if len(parts) >= 5 {
					orgName := parts[3]
					repoName := parts[4]
					if strings.Contains(orgName, "${PN}") {
						orgName = strings.Replace(orgName, "${PN}", pkgInfo.Name, -1)
					}
					if strings.Contains(repoName, "${PN}") {
						repoName = strings.Replace(repoName, "${PN}", pkgInfo.Name, -1)
					}
					pkgInfo.GitRepo = fmt.Sprintf("https://github.com/%s/%s.git", orgName, repoName)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return PackageInfo{}, fmt.Errorf("Error reading ebuild file: %v", err)
	}
	dependencies, err := getDependenciesFromCommand(pkgInfo.Name + "-" + pkgInfo.Version)
	if err != nil {
		pkgInfo.Depends = nil
	} else {
		pkgInfo.Depends = dependencies
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

func FetchAndParseEbuildFiles(directory string) (map[string]PackageInfo, error) {
	pkgInfoMap := make(map[string]PackageInfo)

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(info.Name(), ".ebuild") {
			pkgInfo, err := ParseEbuild(path)
			if err != nil {
				return fmt.Errorf("failed to parse ebuild file %s: %v", path, err)
			}
			if pkgInfo.Name != "" {
				pkgInfoMap[pkgInfo.Name] = pkgInfo
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk through ebuild directory: %v", err)
	}

	return pkgInfoMap, nil
}

func UpdateOrInsertDatabase(pkgInfoMap map[string]PackageInfo) error {
	db, err := storage.GetDefaultDatabaseConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	for pkgName, pkgInfo := range pkgInfoMap {
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM gentoo_packages WHERE package = $1)", pkgName).Scan(&exists)
		if err != nil {
			return err
		}
		if !exists {
			_, err := db.Exec("INSERT INTO gentoo_packages (package, version, depends_count, description, homepage, page_rank) VALUES ($1, $2, $3, $4, $5, $6)",
				pkgName, pkgInfo.Version, pkgInfo.DependsCount, pkgInfo.Description, pkgInfo.Homepage, pkgInfo.PageRank)
			if err != nil {
				return err
			}
		} else {
			_, err := db.Exec("UPDATE gentoo_packages SET version = $1, depends_count = $2, description = $3, homepage = $4, page_rank = $5 WHERE package = $6",
				pkgInfo.Version, pkgInfo.DependsCount, pkgInfo.Description, pkgInfo.Homepage, pkgInfo.PageRank, pkgName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func Gentoo(outputPath string) {
	baseDirectory := "gentoo"
	err := cloneGentooRepo(baseDirectory)
	if err != nil {
		fmt.Printf("Error cloning Gentoo repository: %v\n", err)
		return
	}

	cmd := exec.Command("emerge", "--sync")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error executing emerge --sync: %v\n", err)
	}

	pkgInfoMap, err := FetchAndParseEbuildFiles(baseDirectory)
	if err != nil {
		fmt.Printf("Error fetching package info: %v\n", err)
		return
	}

	fmt.Println("Fetched and parsed ebuild files successfully.")

	depMap := make(map[string][]string)
	for pkgName := range pkgInfoMap {
		visited := make(map[string]bool)
		deps := getAllDep(pkgInfoMap, pkgName, visited, []string{})
		depMap[pkgName] = deps
	}

	countMap := make(map[string]int)
	for _, deps := range depMap {
		for _, dep := range deps {
			countMap[dep]++
		}
	}

	pageRankMap := pageRank(pkgInfoMap, 0.85, 20)

	for pkgName, pkgInfo := range pkgInfoMap {
		pkgInfo.PageRank = pageRankMap[pkgName]
		depCount := countMap[pkgName]
		pkgInfo.DependsCount = depCount
		pkgInfoMap[pkgName] = pkgInfo
	}

	err = UpdateOrInsertDatabase(pkgInfoMap)
	if err != nil {
		fmt.Printf("Error updating database: %v\n", err)
		return
	}
	for pkgName, pkgInfo := range pkgInfoMap {
		fmt.Println("Storing dependencies for package", pkgName, pkgInfo.Depends)
		if err := storeDependenciesInDatabase(pkgName, pkgInfo.Depends); err != nil {
			if isUniqueViolation(err) {
				continue
			}
			fmt.Printf("Error storing dependencies for package %s: %v\n", pkgName, err)
		}
	}
	fmt.Println("Database updated successfully.")

	if outputPath != "" {
		err := generateDependencyGraph(pkgInfoMap, outputPath)
		if err != nil {
			fmt.Printf("Error generating dependency graph: %v\n", err)
			return
		}
		fmt.Println("Dependency graph generated successfully.")
	}
}

func cloneGentooRepo(baseDirectory string) error {
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

func getAllDep(packages map[string]PackageInfo, pkgName string, visited map[string]bool, deps []string) []string {
	if visited[pkgName] {
		return deps
	}

	visited[pkgName] = true
	deps = append(deps, pkgName)

	if pkg, ok := packages[pkgName]; ok {
		for _, depName := range pkg.Depends {
			deps = getAllDep(packages, depName, visited, deps)
		}
	}
	return deps
}

func pageRank(pkgInfoMap map[string]PackageInfo, d float64, iterations int) map[string]float64 {
	ranks := make(map[string]float64)
	N := float64(len(pkgInfoMap))

	// Initialize ranks
	for pkgName := range pkgInfoMap {
		ranks[pkgName] = 1.0 / N
	}

	for i := 0; i < iterations; i++ {
		newRanks := make(map[string]float64)
		for pkgName := range pkgInfoMap {
			newRanks[pkgName] = (1 - d) / N
		}

		for pkgName, pkgInfo := range pkgInfoMap {
			var depNum int
			for _, depName := range pkgInfo.Depends {
				if _, exists := pkgInfoMap[depName]; exists {
					depNum++
				}
			}
			share := ranks[pkgName] / float64(depNum)
			for _, dep := range pkgInfo.Depends {
				if _, exists := pkgInfoMap[dep]; exists {
					newRanks[dep] += d * share
				}
			}
		}

		ranks = newRanks
	}

	return ranks
}

func generateDependencyGraph(pkgInfoMap map[string]PackageInfo, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	writer.WriteString("digraph {\n")

	packageIndices := make(map[string]int)
	index := 0

	for pkgName, pkgInfo := range pkgInfoMap {
		packageIndices[pkgName] = index
		label := fmt.Sprintf("%s@%s", pkgName, pkgInfo.Description)
		writer.WriteString(fmt.Sprintf("  %d [label=\"%s\"];\n", index, label))
		index++
	}

	for pkgName, pkgInfo := range pkgInfoMap {
		pkgIndex := packageIndices[pkgName]
		for _, depName := range pkgInfo.Depends {
			if depIndex, ok := packageIndices[depName]; ok {
				writer.WriteString(fmt.Sprintf("  %d -> %d;\n", pkgIndex, depIndex))
			}
		}
	}

	writer.WriteString("}\n")
	writer.Flush()
	return nil
}

func isUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}
