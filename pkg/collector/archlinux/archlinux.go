package archlinux

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/lib/pq"
)

func updateOrInsertDatabase(pkgInfoMap map[string]DepInfo) error {
	db, err := storage.GetDefaultDatabaseConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	for pkgName, pkgInfo := range pkgInfoMap {
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM arch_packages WHERE package = $1)", pkgName).Scan(&exists)
		if err != nil {
			return err
		}

		if !exists {
			_, err := db.Exec("INSERT INTO arch_packages (package, depends_count, description, homepage, version, page_rank) VALUES ($1, $2, $3, $4, $5, $6)",
				pkgName, pkgInfo.DependsCount, pkgInfo.Description, pkgInfo.Homepage, pkgInfo.Version, pkgInfo.PageRank)
			if err != nil {
				return err
			}
		} else {
			_, err := db.Exec("UPDATE arch_packages SET depends_count = $1, description = $2, homepage = $3, version = $4, page_rank = $5 WHERE package = $6",
				pkgInfo.DependsCount, pkgInfo.Description, pkgInfo.Homepage, pkgInfo.Version, pkgInfo.PageRank, pkgName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func storeDependenciesInDatabase(pkgName string, dependencies []DepInfo) error {
	db, err := storage.GetDefaultDatabaseConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	for _, dep := range dependencies {
		_, err := db.Exec("INSERT INTO arch_relationships (frompackage, topackage) VALUES ($1, $2)", pkgName, dep.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

type DepInfo struct {
	Name         string
	Arch         string
	Version      string
	Description  string
	Homepage     string
	DependsCount int
	PageRank     float64
}

func toDep(dep string, rawContent string) DepInfo {
	re := regexp.MustCompile(`^([^=><!]+?)(?:([=><!]+)([^:]+))?(?::(.+?))?(?:\s*\((.+)\))?$`)
	matches := re.FindStringSubmatch(dep)

	depInfo := DepInfo{Name: dep, Arch: "", Version: "", Description: "", Homepage: ""}

	if matches != nil {
		depInfo.Name = matches[1]
		depInfo.Version = matches[2] + matches[3]
		depInfo.Arch = matches[4]
	}

	descriptionRegex := regexp.MustCompile(`(?m)^%DESC%\s*(.+)$`)
	homepageRegex := regexp.MustCompile(`(?m)^%URL%\s*(.+)$`)

	if descMatches := descriptionRegex.FindStringSubmatch(rawContent); len(descMatches) > 1 {
		depInfo.Description = descMatches[1]
	}

	if homeMatches := homepageRegex.FindStringSubmatch(rawContent); len(homeMatches) > 1 {
		depInfo.Homepage = homeMatches[1]
	}

	return depInfo
}

func extractTarGz(gzipStream io.Reader, dest string) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err
	}
	defer uncompressedStream.Close()

	tarReader := tar.NewReader(uncompressedStream)
	hasFiles := false

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		hasFiles = true
		target := filepath.Join(dest, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(file, tarReader); err != nil {
				file.Close()
				return err
			}
			file.Close()
		}
	}

	if !hasFiles {
		return fmt.Errorf("empty tar archive")
	}

	return nil
}

func readDescFile(descPath string) (DepInfo, []DepInfo, error) {
	file, err := os.Open(descPath)
	if err != nil {
		return DepInfo{}, nil, err
	}
	defer file.Close()

	var pkgInfo DepInfo
	var dependencies []DepInfo
	var inPackageSection, inDependSection bool
	var rawContent strings.Builder
	var expectNextLine string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "%NAME%" {
			inPackageSection = true
			continue
		}
		if line == "%VERSION%" {
			expectNextLine = "version"
			inPackageSection = false
			inDependSection = false
			continue
		}
		if line == "%DEPENDS%" {
			inDependSection = true
			inPackageSection = false
			continue
		}
		if strings.HasPrefix(line, "%") {
			inPackageSection = false
			inDependSection = false
		}

		if expectNextLine == "version" {
			pkgInfo.Version = line
			expectNextLine = ""
			continue
		}

		if inPackageSection && line != "" {
			rawContent.WriteString(line + "\n")
			// log.Println(rawContent.String())
			pkgInfo = toDep(line, rawContent.String())
		}

		if inDependSection && line != "" {
			rawContent.WriteString(line + "\n")
			dependencies = append(dependencies, toDep(line, rawContent.String()))
		}

		if line == "%URL%" {
			expectNextLine = "url"
		} else if line == "%DESC%" {
			expectNextLine = "desc"
		} else if expectNextLine == "url" {
			pkgInfo.Homepage = line
			expectNextLine = ""
		} else if expectNextLine == "desc" {
			pkgInfo.Description = line
			expectNextLine = ""
		}
	}

	if err := scanner.Err(); err != nil {
		return DepInfo{}, nil, err
	}
	return pkgInfo, dependencies, nil
}

func generateDependencyGraph(packages map[string]map[string]interface{}, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	writer.WriteString("digraph {\n")

	packageIndices := make(map[string]int)
	index := 0

	for pkgName, pkgInfo := range packages {
		packageIndices[pkgName] = index
		label := fmt.Sprintf("%s@%s", pkgName, pkgInfo["Info"].(DepInfo).Version)
		writer.WriteString(fmt.Sprintf("  %d [label=\"%s\"];\n", index, label))
		index++
	}

	for pkgName, pkgInfo := range packages {
		pkgIndex := packageIndices[pkgName]
		if depends, ok := pkgInfo["Depends"].([]DepInfo); ok {
			for _, dep := range depends {
				if depIndex, ok := packageIndices[dep.Name]; ok {
					writer.WriteString(fmt.Sprintf("  %d -> %d [label=\"%s\"];\n", pkgIndex, depIndex, dep.Version))
				}
			}
		}
	}

	writer.WriteString("}\n")
	writer.Flush()
	return nil
}

func getAllDep(packages map[string]map[string]interface{}, pkgName string, deps []string) []string {
	deps = append(deps, pkgName)
	if pkg, ok := packages[pkgName]; ok {
		if depends, ok := pkg["Depends"].([]DepInfo); ok {
			for _, dep := range depends {
				pkgname := dep.Name
				if !contains(deps, pkgname) {
					deps = getAllDep(packages, pkgname, deps)
				}
			}
		}
	}
	return deps
}
func calculatePageRank(packages map[string]map[string]interface{}, iterations int, dampingFactor float64) map[string]float64 {
	pageRank := make(map[string]float64)
	outgoingLinks := make(map[string]int)

	for pkgName, pkgInfo := range packages {
		pageRank[pkgName] = 1.0 / float64(len(packages))
		if depends, ok := pkgInfo["Depends"].([]DepInfo); ok {
			outgoingLinks[pkgName] = len(depends)
		} else {
			outgoingLinks[pkgName] = 0
		}
	}

	for i := 0; i < iterations; i++ {
		newPageRank := make(map[string]float64)
		for pkgName := range packages {
			newPageRank[pkgName] = (1 - dampingFactor) / float64(len(packages))
		}

		for pkgName, pkgInfo := range packages {
			var depNum int
			if depends, ok := pkgInfo["Depends"].([]DepInfo); ok {
				for _, dep := range depends {
					if outgoingLinks[dep.Name] > 0 {
						depNum++
					}
				}
			}
			if depends, ok := pkgInfo["Depends"].([]DepInfo); ok {
				for _, dep := range depends {
					if outgoingLinks[dep.Name] > 0 {
						newPageRank[dep.Name] += dampingFactor * pageRank[pkgName] / float64(depNum)
					}
				}
			}
		}

		pageRank = newPageRank
	}

	return pageRank
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func Archlinux(outputPath string) {
	downloadDir := "./download"

	// if _, err := os.Stat(downloadDir); os.IsNotExist(err) {
	// 	log.Println("Download directory not found, starting download...")
	DownloadFiles()
	// }

	log.Println("Getting package list...")
	extractDir := "./extracted"
	packages := make(map[string]map[string]interface{})
	packageNamePattern := regexp.MustCompile(`^([a-zA-Z0-9\-_]+)-([0-9\._]+)`)

	if _, err := os.Stat(extractDir); os.IsNotExist(err) {
		err := os.Mkdir(extractDir, 0o755)
		if err != nil {
			log.Printf("Error creating extract directory: %v\n", err)
			return
		}
	}

	err := filepath.Walk(downloadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".tar.gz") {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			err = extractTarGz(file, extractDir)
			if err != nil {
				if err.Error() == "empty tar archive" {
					return nil
				}
				return nil
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("Error walking through download directory: %v\n", err)
		return
	}

	err = filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), "desc") {
			packageName := packageNamePattern.FindStringSubmatch(filepath.Base(filepath.Dir(path)))
			if packageName != nil {
				pkgInfo, dependencies, err := readDescFile(path)
				if err != nil {
					return err
				}
				if _, ok := packages[pkgInfo.Name]; !ok {
					packages[pkgInfo.Name] = make(map[string]interface{})
				}
				packages[pkgInfo.Name]["Depends"] = dependencies
				packages[pkgInfo.Name]["Info"] = pkgInfo
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("Error walking through extracted directory: %v\n", err)
		return
	}
	log.Printf("Done, total: %d packages.\n", len(packages))

	if outputPath != "" {
		err := generateDependencyGraph(packages, outputPath)
		if err != nil {
			log.Printf("Error generating dependency graph: %v\n", err)
			return
		}
		log.Println("Dependency graph generated successfully.")
	}
	log.Println("Building dependencies graph...")
	keys := make([]string, 0, len(packages))
	for k := range packages {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	depMap := make(map[string][]string)
	for _, pkgName := range keys {
		deps := getAllDep(packages, pkgName, []string{})
		depMap[pkgName] = deps
	}

	pagerank := calculatePageRank(packages, 20, 0.85)
	log.Println("Calculating dependencies count...")
	countMap := make(map[string]int)
	for _, deps := range depMap {
		for _, dep := range deps {
			countMap[dep]++
		}
	}

	pkgInfoMap := make(map[string]DepInfo)

	for pkgName, pkgInfo := range packages {
		depCount := countMap[pkgName]

		var description, homepage, version string

		if info, ok := pkgInfo["Info"].(DepInfo); ok {
			description = info.Description
			homepage = info.Homepage
			version = info.Version
		} else {
			description = ""
			homepage = ""
		}
		pageRank, _ := pagerank[pkgName]

		pkgInfoMap[pkgName] = DepInfo{
			Name:         pkgName,
			DependsCount: depCount,
			Description:  description,
			Homepage:     homepage,
			Version:      version,
			PageRank:     pageRank,
		}
	}

	err = updateOrInsertDatabase(pkgInfoMap)
	if err != nil {
		log.Printf("Error updating database: %v\n", err)
		return
	}
	for _, pkgInfo := range packages {
		if packageInfo, ok := pkgInfo["Info"].(DepInfo); ok {
			packageName := packageInfo.Name
			if depends, ok := pkgInfo["Depends"].([]DepInfo); ok {
				if err := storeDependenciesInDatabase(packageName, depends); err != nil {
					if isUniqueViolation(err) {
						continue
					}
					log.Printf("Error storing dependencies for package %s: %v\n", packageName, err)
					return
				}
			} else {
				log.Printf("No valid dependencies found for package %s\n", packageName)
			}
		} else {
			log.Printf("Invalid package name for pkgInfo: %v\n", pkgInfo)
		}
	}
	log.Println("Database updated successfully.")
}
func isUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}
