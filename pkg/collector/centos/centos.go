package centos

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/lib/pq"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
)

var URL = "https://mirrors.aliyun.com/centos/7/os/x86_64/repodata/2b479c0f3efa73f75b7fb76c82687744275fff78e4a138b5b3efba95f91e099e-primary.xml.gz"

type PackageInfo struct {
	Depends      []string
	DependsCount int
	Description  string
	GitRepo      string
	Homepage     string
	Name         string
	PageRank     float64
	URL          string
	Version      string
}

func Centos(outputPath string) {
	pkgInfoMap, err := getDependencies()
	if err != nil {
		log.Fatal(err)
	}

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

	pagerank := calculatePageRank(pkgInfoMap, 20, 0.85)

	for pkgName, pkgInfo := range pkgInfoMap {
		pagerankVal := pagerank[pkgName]
		depCount := countMap[pkgName]
		pkgInfo.PageRank = pagerankVal
		pkgInfo.DependsCount = depCount
		pkgInfoMap[pkgName] = pkgInfo
	}
	err = updateOrInsertDatabase(pkgInfoMap)
	if err != nil {
		fmt.Printf("Error updating database: %v\n", err)
		return
	}
	for pkgName, pkgInfo := range pkgInfoMap {
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

func decompressGzip(data []byte) (string, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer reader.Close()

	uncompressedData, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(uncompressedData), nil
}
func getDependencies() (map[string]PackageInfo, error) {
	resp, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	data, err := decompressGzip(body)
	if err != nil {
		return nil, err
	}
	data = strings.Replace(data, "\x00", "", -1)
	packages := make(map[string]PackageInfo)
	decoder := xml.NewDecoder(strings.NewReader(data))
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		if charset == "utf-8" {
			return input, nil
		}
		return nil, fmt.Errorf("unsupported charset: %s", charset)
	}
	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		switch se := tok.(type) {
		case xml.StartElement:
			if se.Name.Local == "package" {
				var pkgData struct {
					Type string `xml:"type,attr"`
					XML  string `xml:",innerxml"`
				}
				err := decoder.DecodeElement(&pkgData, &se)
				if err != nil {
					return nil, err
				}

				if pkgData.Type == "rpm" {
					lines := strings.Split(pkgData.XML, "\n")
					for i, line := range lines {
						if len(line) > 2 {
							lines[i] = line[2:]
						}
					}
					trimmedXML := strings.Join(lines, "\n")
					pkgInfo, err := parsePackageXML(trimmedXML[1:])
					if err != nil {
						return nil, err
					}

					if _, exists := packages[pkgInfo.Name]; !exists {
						packages[pkgInfo.Name] = pkgInfo
					}
				}
			}
		}
	}
	return packages, nil
}

func parsePackageXML(data string) (PackageInfo, error) {
	data = strings.Map(func(r rune) rune {
		if r == '\x00' || r > 127 {
			return -1
		}
		return r
	}, data)
	decoder := xml.NewDecoder(strings.NewReader(data))
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		if charset == "utf-8" {
			return input, nil
		}
		return nil, fmt.Errorf("unsupported charset: %s", charset)
	}
	var pkgInfo PackageInfo
	var depends []string

	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return PackageInfo{}, err
		}

		switch se := tok.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "name":
				var name string
				if err := decoder.DecodeElement(&name, &se); err != nil {
					return PackageInfo{}, err
				}
				pkgInfo.Name = name
			case "description":
				var description string
				if err := decoder.DecodeElement(&description, &se); err != nil {
					return PackageInfo{}, err
				}
				if len(description) > 255 {
					description = description[:254]
				}
				pkgInfo.Description = description
			case "url":
				var url string
				if err := decoder.DecodeElement(&url, &se); err != nil {
					return PackageInfo{}, err
				}
				pkgInfo.URL = url
			case "version":
				var version struct {
					Epoch string `xml:"epoch,attr"`
					Ver   string `xml:"ver,attr"`
					Rel   string `xml:"rel,attr"`
				}
				if err := decoder.DecodeElement(&version, &se); err != nil {
					return PackageInfo{}, err
				}
				pkgInfo.Version = fmt.Sprintf("%s:%s-%s", version.Epoch, version.Ver, version.Rel)
			case "entry":
				var entry struct {
					Name string `xml:"name,attr"`
				}
				if err := decoder.DecodeElement(&entry, &se); err != nil {
					return PackageInfo{}, err
				}
				depends = append(depends, entry.Name)
			}
		}
	}

	pkgInfo.Depends = depends
	return pkgInfo, nil
}

func calculatePageRank(pkgInfoMap map[string]PackageInfo, iterations int, dampingFactor float64) map[string]float64 {
	pageRank := make(map[string]float64)
	numPackages := len(pkgInfoMap)

	for pkgName := range pkgInfoMap {
		pageRank[pkgName] = 1.0 / float64(numPackages)
	}

	for i := 0; i < iterations; i++ {
		newPageRank := make(map[string]float64)

		for pkgName := range pkgInfoMap {
			newPageRank[pkgName] = (1 - dampingFactor) / float64(numPackages)
		}

		for pkgName, pkgInfo := range pkgInfoMap {
			var depNum int
			for _, depName := range pkgInfo.Depends {
				if _, exists := pkgInfoMap[depName]; exists {
					depNum++
				}
			}
			for _, depName := range pkgInfo.Depends {
				if _, exists := pkgInfoMap[depName]; exists {
					newPageRank[depName] += dampingFactor * (pageRank[pkgName] / float64(depNum))
				}
			}
		}
		pageRank = newPageRank
	}
	return pageRank
}

func isUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}

func updateOrInsertDatabase(pkgInfoMap map[string]PackageInfo) error {
	db, err := storage.GetDefaultDatabaseConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	for pkgName, pkgInfo := range pkgInfoMap {
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM centos_packages WHERE package = $1)", pkgName).Scan(&exists)
		if err != nil {
			return err
		}
		if !exists {
			_, err := db.Exec("INSERT INTO centos_packages (package, depends_count, description, homepage, page_rank, version) VALUES ($1, $2, $3, $4, $5, $6)",
				pkgName, pkgInfo.DependsCount, pkgInfo.Description, pkgInfo.URL, pkgInfo.PageRank, pkgInfo.Version)
			if err != nil {
				return err
			}
		} else {
			_, err := db.Exec("UPDATE centos_packages SET depends_count = $1, description = $2, homepage = $3, page_rank = $4, version = $5 WHERE package = $6",
				pkgInfo.DependsCount, pkgInfo.Description, pkgInfo.URL, pkgInfo.PageRank, pkgInfo.Version, pkgName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func storeDependenciesInDatabase(pkgName string, dependencies []string) error {
	db, err := storage.GetDefaultDatabaseConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	for _, dep := range dependencies {
		_, err := db.Exec("INSERT INTO centos_relationships (frompackage, topackage) VALUES ($1, $2)", pkgName, dep)
		if err != nil {
			return err
		}
	}
	return nil
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
