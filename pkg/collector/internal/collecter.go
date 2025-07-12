package collector

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/HUSTSecLab/OpenSift/pkg/storage/repository"
	"github.com/samber/lo"
)

type CollecterInterface interface {
	GetPackageInfo(urls PackageURL) string
	UpdateOrInsertDatabase(ac storage.AppDatabaseContext)
	UpdateOrInsertDistDependencyDatabase(ac storage.AppDatabaseContext)
	GenerateDependencyGraph(outputPath string) error
	GetAllDep(pkgName string, visited map[string]bool, deps []string) []string
	PageRank(d float64, iterations int)
	ParseInfo(data string)
	GetDepCount()
	GetDep()
	SetPkgInfo(pkgName string, pkgInfo *PackageInfo)
	GetPkgInfo(pkgName string) *PackageInfo
	CalculateDistImpact()
	UpdateDistRepoCount(ac storage.AppDatabaseContext)
}

type Collecter struct {
	PkgInfoMap             map[string]PackageInfo
	DistRepoCount          int
	Type                   repository.DistType
	DistPackageTablePrefix repository.DistPackageTablePrefix
}

func NewCollector(Type repository.DistType, DistPackageTablePrefix repository.DistPackageTablePrefix) CollecterInterface {
	return &Collecter{
		PkgInfoMap:             make(map[string]PackageInfo),
		Type:                   Type,
		DistPackageTablePrefix: DistPackageTablePrefix,
	}
}

func (cl *Collecter) UpdateOrInsertDatabase(ac storage.AppDatabaseContext) {
	for _, pkgInfo := range cl.PkgInfoMap {
		distPackage := pkgInfo.ParseDistPackage()
		repo := repository.NewDistPackageRepository(ac, cl.DistPackageTablePrefix)

		if pkgInfo.Name == "" {
			continue
		}

		err := repo.InsertOrUpdate(distPackage)
		if err != nil {
			log.Println("Error inserting package info into database:", err)
		}
	}
}

func (cl *Collecter) GenerateDependencyGraph(outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	writer.WriteString("digraph {\n")

	packageIndices := make(map[string]int)
	index := 0

	for pkgName, pkgInfo := range cl.PkgInfoMap {
		packageIndices[pkgName] = index
		label := fmt.Sprintf("%s@%s", pkgName, pkgInfo.Description)
		writer.WriteString(fmt.Sprintf("  %d [label=\"%s\"];\n", index, label))
		index++
	}

	for pkgName, pkgInfo := range cl.PkgInfoMap {
		pkgIndex := packageIndices[pkgName]
		for _, depName := range pkgInfo.IndirectDepends {
			if depIndex, ok := packageIndices[depName]; ok {
				writer.WriteString(fmt.Sprintf("  %d -> %d;\n", pkgIndex, depIndex))
			}
		}
	}

	writer.WriteString("}\n")
	writer.Flush()
	return nil
}

func (cl *Collecter) GetAllDep(pkgName string, visited map[string]bool, deps []string) []string {
	if visited[pkgName] {
		return deps
	}

	visited[pkgName] = true
	deps = append(deps, pkgName)

	if pkg, ok := cl.PkgInfoMap[pkgName]; ok {
		for _, depName := range pkg.DirectDepends {
			deps = cl.GetAllDep(depName, visited, deps)
		}
	}
	return deps
}

func (cl *Collecter) PageRank(d float64, iterations int) {
	ranks := make(map[string]float64)
	N := float64(len(cl.PkgInfoMap))

	for pkgName := range cl.PkgInfoMap {
		ranks[pkgName] = 1.0 / N
	}

	for i := 0; i < iterations; i++ {
		newRanks := make(map[string]float64)

		for pkgName := range cl.PkgInfoMap {
			newRanks[pkgName] = (1 - d) / N
		}

		for pkgName, pkgInfo := range cl.PkgInfoMap {
			var depNum int
			for _, depName := range pkgInfo.DirectDepends {
				if _, exists := cl.PkgInfoMap[depName]; exists {
					depNum++
				}
			}

			share := ranks[pkgName] / float64(depNum)
			for _, dep := range pkgInfo.DirectDepends {
				if _, exists := cl.PkgInfoMap[dep]; exists {
					newRanks[dep] += d * share
				}
			}
		}

		ranks = newRanks
	}

	for pkgName, rank := range ranks {
		pkgInfo := cl.PkgInfoMap[pkgName]
		pkgInfo.PageRank = rank
		cl.PkgInfoMap[pkgName] = pkgInfo
	}
}

func (cl *Collecter) ParseInfo(pkgInfo string) {
	log.Println("Parsing package info for", pkgInfo)
}

func (cl *Collecter) GetPackageInfo(urls PackageURL) string {
	var result string

	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Error making HTTP request:", err)
			continue
		}
		defer resp.Body.Close()

		switch {
		case strings.HasSuffix(url, ".tar.gz"):
			gzipReader, err := gzip.NewReader(resp.Body)
			if err != nil {
				fmt.Println("Error creating gzip reader:", err)
				continue
			}
			defer gzipReader.Close()

			tarReader := tar.NewReader(gzipReader)
			var body strings.Builder

			for {
				_, err := tarReader.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					fmt.Println("Error reading tar content:", err)
					break
				}

				buf := new(strings.Builder)
				if _, err := io.Copy(buf, tarReader); err != nil {
					fmt.Println("Error extracting tar content:", err)
					break
				}
				body.WriteString(buf.String())
			}
			result += body.String()

		case strings.HasSuffix(url, ".gz"):
			gzipReader, err := gzip.NewReader(resp.Body)
			if err != nil {
				fmt.Println("Error creating gzip reader:", err)
				continue
			}
			defer gzipReader.Close()

			reader := bufio.NewReader(gzipReader)
			var body strings.Builder

			for {
				line, err := reader.ReadString('\n')
				body.WriteString(line)
				if err == io.EOF {
					break
				}
				if err != nil {
					fmt.Println("Error reading response body:", err)
					break
				}
			}
			result += body.String()

		default:
			fmt.Println("Unsupported file type:", url)
			continue
		}
	}

	return result
}

func (cl *Collecter) GetDepCount() {
	countMap := make(map[string]int)

	for _, deps := range cl.PkgInfoMap {
		for _, dep := range deps.IndirectDepends {
			if idx := strings.Index(dep, ">"); idx != -1 {
				dep = dep[:idx]
			} else if idx := strings.Index(dep, "="); idx != -1 {
				dep = dep[:idx]
			} else if idx := strings.Index(dep, "<"); idx != -1 {
				dep = dep[:idx]
			}
			countMap[dep]++
		}
	}

	for pkgName, count := range countMap {
		if pkgInfo, ok := cl.PkgInfoMap[pkgName]; ok {
			pkgInfo.DependsCount = count
			cl.PkgInfoMap[pkgName] = pkgInfo
		}
	}
}

func (cl *Collecter) GetDep() {
	for pkgName := range cl.PkgInfoMap {
		visited := make(map[string]bool)
		deps := cl.GetAllDep(pkgName, visited, []string{})
		pkgInfo := cl.PkgInfoMap[pkgName]
		pkgInfo.IndirectDepends = deps
		cl.PkgInfoMap[pkgName] = pkgInfo
	}
}

func (cl *Collecter) SetPkgInfo(pkgName string, pkgInfo *PackageInfo) {
	pkgInfo.Type = cl.Type
	pkgInfo.DistPackageTablePrefix = cl.DistPackageTablePrefix
	cl.PkgInfoMap[pkgName] = *pkgInfo
}

func (cl *Collecter) GetPkgInfo(pkgName string) *PackageInfo {
	if _, ok := cl.PkgInfoMap[pkgName]; !ok {
		return nil
	}
	return lo.ToPtr(cl.PkgInfoMap[pkgName])
}

func (cl *Collecter) UpdateOrInsertDistDependencyDatabase(ac storage.AppDatabaseContext) {
	var distMap = make(map[string]*repository.DistDependency)
	for _, pkgInfo := range cl.PkgInfoMap {
		if pkgInfo.Name == "" {
			continue
		}

		pkgInfo.GetGitlinkByPkg(ac)
		distPackage := pkgInfo.ParseDistLinkInfo()
		if pkgInfo.Gitlink != "" && pkgInfo.Gitlink != "NA" && pkgInfo.Gitlink != "NaN" {
			if _, ok := distMap[*distPackage.GitLink]; !ok {
				distMap[*distPackage.GitLink] = distPackage
			} else {
				distMap[*distPackage.GitLink].DepCount = lo.ToPtr(*distMap[*distPackage.GitLink].DepCount + *distPackage.DepCount)
				distMap[*distPackage.GitLink].DepImpact = lo.ToPtr(*distMap[*distPackage.GitLink].DepImpact + *distPackage.DepImpact)
				distMap[*distPackage.GitLink].PageRank = lo.ToPtr(*distMap[*distPackage.GitLink].PageRank + *distPackage.PageRank)
			}
		}
	}

	for _, distPackage := range distMap {
		repo := repository.NewDistDependencyRepository(ac)
		err := repo.InsertOrUpdate(distPackage)
		if err != nil {
			log.Println("Error inserting package info into database:", err)
		}
	}
}

func (cl *Collecter) CalculateDistImpact() {
	for _, pkgInfo := range cl.PkgInfoMap {
		pkgInfo.CalculateImpact(cl.DistRepoCount)
		cl.PkgInfoMap[pkgInfo.Name] = pkgInfo
	}
}

func (cl *Collecter) UpdateDistRepoCount(ac storage.AppDatabaseContext) {
	repo := repository.NewDistDependencyRepository(ac)
	count, err := repo.QueryDistCountByType(cl.Type)
	if err != nil {
		log.Println("Error getting count from dist dependency repository:", err)
		return
	}

	cl.DistRepoCount = count
}
