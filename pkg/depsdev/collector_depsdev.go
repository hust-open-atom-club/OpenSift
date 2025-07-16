package depsdev

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/HUSTSecLab/OpenSift/pkg/storage/repository"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/samber/lo"
)

var Pkg2GitLink sync.Map

var PackageCounts = map[repository.LangEcosystemType]int{
	repository.Npm:    3.37e6,
	repository.Go:     1.29e6,
	repository.Maven:  668e3,
	repository.Pypi:   574e3,
	repository.NuGet:  430e3,
	repository.Cargo:  168e3,
	repository.Others: 1,
}

type DependentInfo struct {
	DependentCount         int `json:"dependentCount"`
	DirectDependentCount   int `json:"directDependentCount"`
	IndirectDependentCount int `json:"indirectDependentCount"`
}

type VersionInfo struct {
	VersionKey struct {
		Version string `json:"version"`
	} `json:"versionKey"`
	PublishedAt  time.Time `json:"publishedAt"`
	IsDefault    bool      `json:"isDefault"`
	IsDeprecated bool      `json:"isDeprecated"`
}

type PackageInfo struct {
	Versions []VersionInfo `json:"versions"`
}

type Node struct {
	VersionKey Version  `json:"versionKey"`
	Bundled    bool     `json:"bundled"`
	Relation   string   `json:"relation"`
	Errors     []string `json:"errors"`
}

type Edge struct {
	FromNode    int    `json:"fromNode"`
	ToNode      int    `json:"toNode"`
	Requirement string `json:"requirement"`
}

type Dependencies struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
	Error string `json:"error"`
}

type Version struct {
	System  string `json:"system"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type PkgInfo struct {
	VersionKey         Version
	RelationType       string   `json:"relationType"`
	RelationProvenance string   `json:"relationProvenance"`
	SlsaProvenances    []string `json:"slsaProvenances"`
	Attestations       []string `json:"attestations"`
}

type DepsDevInfo struct {
	Versions []PkgInfo `json:"versions"`
}

type EcoSystemRatio struct {
	NpmRatio   float64
	GoRatio    float64
	MavenRatio float64
	PyPiRatio  float64
	NuGetRatio float64
	CargoRatio float64
}

func getLatestVersion(repo, projectType string) string {
	ctx := context.Background()

	url := fmt.Sprintf("https://api.deps.dev/v3alpha/systems/%s/packages/%s", projectType, repo)

	req, _ := http.NewRequest("GET", url, nil)
	client := &http.Client{}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result PackageInfo
	json.Unmarshal(body, &result)

	var latestVersion string
	var latestDate time.Time
	for _, version := range result.Versions {
		if version.PublishedAt == (time.Time{}) {
			latestVersion = version.VersionKey.Version
			break
		}
		if version.PublishedAt.After(latestDate) && version.IsDefault {
			latestDate = version.PublishedAt
			latestVersion = version.VersionKey.Version
		}
	}

	return latestVersion
}

func queryDepsDev(projectType, projectName, version string) int {
	projectName = url.QueryEscape(projectName)
	// version = getLatestVersion(projectName, projectType)
	url := fmt.Sprintf("https://api.deps.dev/v3alpha/systems/%s/packages/%s/versions/%s:dependents", projectType, projectName, version)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		// fmt.Println("Error fetching package information:", err)
		return 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0
	}

	var info DependentInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		// fmt.Println("Error decoding response:", err)
		return 0
	}
	return info.DependentCount
}

func getGitlink(db *sql.DB) []string {
	rows, err := db.Query("SELECT git_link FROM git_metrics")
	if err != nil {
		fmt.Println("Error querying git_metrics:", err)
		return nil
	}
	defer rows.Close()
	var gitLinks []string
	for rows.Next() {
		var gitLink string
		if err := rows.Scan(&gitLink); err != nil {
			fmt.Println("Error scanning git_link:", err)
			return nil
		}
		gitLinks = append(gitLinks, gitLink)
	}
	return gitLinks
}

func queryDepsName(gitlink string, rdb *redis.Client, workerPoolSize int) *sync.Map {
	depMap := &sync.Map{}
	if strings.Contains(gitlink, ".git") {
		gitlink = strings.TrimSuffix(gitlink, ".git")
	}
	var repo, name string
	if len(strings.Split(gitlink, "/")) == 5 {
		repo = strings.Split(gitlink, "/")[3]
		name = strings.Split(gitlink, "/")[4]
	}
	urlstr := fmt.Sprintf("https://api.deps.dev/v3alpha/projects/github.com%%2f%s%%2f%s:packageversions", repo, name)
	resp, err := http.Get(urlstr)
	if err != nil {
		fmt.Println("Error querying deps.dev:", err)
		return depMap
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return depMap
	}
	var result DepsDevInfo

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("Error decoding response:", err)
		return depMap
	}
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, workerPoolSize)

	for _, item := range result.Versions {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(item PkgInfo) {
			defer wg.Done()
			defer func() { <-semaphore }()
			name := item.VersionKey.Name
			system := item.VersionKey.System
			version := getLatestVersion(url.QueryEscape(name), system)
			if strings.Contains(name, "\u003E") {
				name = strings.ReplaceAll(name, "\u003E", ">")
			}
			depMap.Store(name, Version{Name: name, System: system, Version: version})
			storage.SetKeyValue(rdb, name, gitlink)
			if _, exists := Pkg2GitLink.Load(name); !exists {
				Pkg2GitLink.Store(name, &sync.Map{})
			}
			gitLinks, _ := Pkg2GitLink.Load(name)
			gitLinks.(*sync.Map).Store(gitlink, struct{}{})
		}(item)
	}
	wg.Wait()
	return depMap
}

type GitMetrics struct {
	LangEcoImpact   float64
	LangEcoPageRank float64
}

func Depsdev(batchSize int, workerPoolSize int, calculatePageRankFlag bool, debugMode bool) {
	ac := storage.GetDefaultAppDatabaseContext()
	repo := repository.NewLangEcoLinkRepository(ac)
	rdb, _ := storage.InitRedis()
	// gitLinks := []string{"https://github.com/facebook/react"}
	gitLinks := fetchGitLink(ac, lo.ToPtr(0))
	pkgMap := &sync.Map{}
	pkgDepMap := &sync.Map{}
	var count int
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, workerPoolSize)

	for _, gitlink := range gitLinks {
		count++
		log.Println("Processing gitlink:", gitlink, "Count: ", count, "/", len(gitLinks))
		wg.Add(1)
		semaphore <- struct{}{}
		go func(gitlink string) {
			defer wg.Done()
			defer func() { <-semaphore }()
			depMap := queryDepsName(gitlink, rdb, workerPoolSize)
			var depWg sync.WaitGroup
			depSemaphore := make(chan struct{}, workerPoolSize)
			depMap.Range(func(pkgName, pkgInfo interface{}) bool {
				depWg.Add(1)
				depSemaphore <- struct{}{}
				go func(pkgName string, pkgInfo Version) {
					defer depWg.Done()
					defer func() { <-depSemaphore }()
					if _, exists := pkgDepMap.Load(pkgInfo.System); !exists {
						pkgDepMap.Store(pkgInfo.System, &sync.Map{})
					}
					systemMap, _ := pkgDepMap.Load(pkgInfo.System)
					systemMap.(*sync.Map).Store(pkgName, queryDepsDev(pkgInfo.System, pkgInfo.Name, pkgInfo.Version))
				}(pkgName.(string), pkgInfo.(Version))
				return true
			})
			depWg.Wait()
			if calculatePageRankFlag {
				pkgdepMap := fetchDep(depMap, workerPoolSize)
				var depWg sync.WaitGroup
				depSemaphore := make(chan struct{}, workerPoolSize)
				pkgdepMap.Range(func(pkgName, pkgInfo interface{}) bool {
					depWg.Add(1)
					depSemaphore <- struct{}{}
					go func(pkgName string, pkgInfo []Version) {
						defer depWg.Done()
						defer func() { <-depSemaphore }()
						if _, exists := pkgMap.Load(pkgName); !exists {
							pkgMap.Store(pkgName, []Version{})
						}
						current, _ := pkgMap.Load(pkgName)
						pkgMap.Store(pkgName, append(current.([]Version), pkgInfo...))
					}(pkgName.(string), pkgInfo.([]Version))
					return true
				})
				depWg.Wait()
				storage.PersistData(rdb)
			}
		}(gitlink)
	}
	wg.Wait()
	var pageRank map[string]float64
	if calculatePageRankFlag {
		pkgMapCopy := make(map[string][]Version)
		pkgMap.Range(func(key, value interface{}) bool {
			pkgMapCopy[key.(string)] = value.([]Version)
			return true
		})
		pageRank = calculatePageRank(pkgMapCopy, 100, 0.85)
	} else {
		pageRank = make(map[string]float64)
		pkgDepMap.Range(func(_, systemMap interface{}) bool {
			systemMap.(*sync.Map).Range(func(pkgName, _ interface{}) bool {
				pageRank[pkgName.(string)] = 0.0
				return true
			})
			return true
		})
	}
	type langEcoKey struct {
		gitLink string
		ltype   repository.LangEcosystemType
	}
	langEco := &sync.Map{}

	// pkgDepMap.Range(func(system, systemMap interface{}) bool {
	// 	systemMap.(*sync.Map).Range(func(pkgName, depCount interface{}) bool {
	// 		if gitLinks, ok := Pkg2GitLink.Load(pkgName); ok {
	// 			gitLinks.(*sync.Map).Range(func(gitlink, _ interface{}) bool {
	// 				log.Printf("System: %s, Package: %s, DepCount: %d, GitLink: %s\n", system.(string), pkgName.(string), depCount.(int), gitlink.(string))
	// 				return true
	// 			})
	// 		} else {
	// 			log.Printf("System: %s, Package: %s, DepCount: %d, GitLink: Not Found\n", system.(string), pkgName.(string), depCount.(int))
	// 		}
	// 		return true
	// 	})
	// 	return true
	// })

	pkgDepMap.Range(func(system, systemMap interface{}) bool {
		systemMap.(*sync.Map).Range(func(pkgName, _ interface{}) bool {
			wg.Add(1)
			semaphore <- struct{}{}
			go func(system, pkgName string) {
				defer wg.Done()
				defer func() { <-semaphore }()
				gitlinks, _ := Pkg2GitLink.Load(pkgName)
				gitlinks.(*sync.Map).Range(func(gitlink, _ interface{}) bool {
					var ltype repository.LangEcosystemType
					switch strings.ToLower(system) {
					case "cargo":
						ltype = repository.Cargo
					case "go":
						ltype = repository.Go
					case "maven":
						ltype = repository.Maven
					case "npm":
						ltype = repository.Npm
					case "nuget":
						ltype = repository.NuGet
					case "pypi":
						ltype = repository.Pypi
					}

					key := langEcoKey{
						gitLink: gitlink.(string),
						ltype:   ltype,
					}

					if depCount, ok := systemMap.(*sync.Map).Load(pkgName); ok {
						if value, exists := langEco.Load(key); !exists {
							langEco.Store(key, GitMetrics{
								LangEcoImpact:   float64(depCount.(int)) / float64(PackageCounts[ltype]),
								LangEcoPageRank: pageRank[pkgName],
							})
						} else {
							langEco.Store(key, GitMetrics{
								LangEcoImpact:   value.(GitMetrics).LangEcoImpact + float64(depCount.(int))/float64(PackageCounts[ltype]),
								LangEcoPageRank: value.(GitMetrics).LangEcoPageRank + pageRank[pkgName],
							})
						}
					}
					return true
				})
			}(system.(string), pkgName.(string))
			return true
		})
		return true
	})
	wg.Wait()

	existingLinks := make(map[string]bool)
	langEco.Range(func(key, _ interface{}) bool {
		existingLinks[key.(langEcoKey).gitLink] = true
		return true
	})

	for _, link := range gitLinks {
		if !existingLinks[link] {
			key := langEcoKey{
				gitLink: link,
				ltype:   repository.LangEcosystemType(6),
			}
			langEco.Store(key, GitMetrics{
				LangEcoImpact:   0,
				LangEcoPageRank: 0,
			})
		}
	}

	var toUpdateList []*repository.LangEcosystem
	langEco.Range(func(key, info interface{}) bool {
		toUpdateList = append(toUpdateList, lo.ToPtr(repository.LangEcosystem{
			GitLink:           lo.ToPtr(key.(langEcoKey).gitLink),
			Type:              lo.ToPtr(key.(langEcoKey).ltype),
			DepCount:          lo.ToPtr(int(info.(GitMetrics).LangEcoImpact * float64(PackageCounts[key.(langEcoKey).ltype]))),
			LangEcoImpact:     lo.ToPtr(info.(GitMetrics).LangEcoImpact),
			Lang_eco_pagerank: lo.ToPtr(info.(GitMetrics).LangEcoPageRank),
		}))
		return true
	})

	err := repo.BatchInsertOrUpdate(toUpdateList)
	if err != nil {
		fmt.Printf("Error updating database: %v\n", err)
	}
}

func fetchDep(depMap *sync.Map, threadnum int) *sync.Map {
	depMapNew := &sync.Map{}
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, threadnum)
	depMap.Range(func(depName, depInfo interface{}) bool {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(depName string, depInfo Version) {
			defer wg.Done()
			defer func() { <-semaphore }()
			var system, name, version string
			system = depInfo.System
			name = depName
			version = depInfo.Version
			result := getAndProcessDependencies(system, name, version)
			depMapNew.Store(name, []Version{})
			for _, node := range result.Nodes {
				if node.Relation == "DIRECT" {
					if deps, ok := depMapNew.Load(name); ok {
						depMapNew.Store(name, append(deps.([]Version), node.VersionKey))
					}
				}
			}
		}(depName.(string), depInfo.(Version))
		return true
	})
	wg.Wait()
	return depMapNew
}

func calculatePageRank(pkgInfoMap map[string][]Version, iterations int, dampingFactor float64) map[string]float64 {
	pageRank := &sync.Map{}
	numPackages := len(pkgInfoMap)

	for pkgName := range pkgInfoMap {
		pageRank.Store(pkgName, 1.0/float64(numPackages))
	}

	for i := 0; i < iterations; i++ {
		newPageRank := &sync.Map{}

		for pkgName := range pkgInfoMap {
			newPageRank.Store(pkgName, (1-dampingFactor)/float64(numPackages))
		}

		var wg sync.WaitGroup
		for pkgName, deps := range pkgInfoMap {
			wg.Add(1)
			go func(pkgName string, deps []Version) {
				defer wg.Done()
				depNum := len(deps)
				for _, depName := range deps {
					if _, exists := pkgInfoMap[depName.Name]; exists {
						if val, ok := pageRank.Load(pkgName); ok {
							if newVal, ok := newPageRank.Load(depName.Name); ok {
								newPageRank.Store(depName.Name, newVal.(float64)+dampingFactor*(val.(float64)/float64(depNum)))
							}
						}
					}
				}
			}(pkgName, deps)
		}
		wg.Wait()
		pageRank = newPageRank
	}

	result := make(map[string]float64)
	pageRank.Range(func(key, value interface{}) bool {
		result[key.(string)] = value.(float64)
		return true
	})

	return result
}

func getAndProcessDependencies(system, name, version string) Dependencies {
	var result Dependencies
	name = url.QueryEscape(name)
	// version = getLatestVersion(name, system)
	url := fmt.Sprintf("https://api.deps.dev/v3alpha/systems/%s/packages/%s/versions/%s:dependencies", system, name, version)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error querying deps.dev:", err)
		return result
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return result
	}
	cleanedBody := removeInvisibleChars(string(body))
	err = json.Unmarshal([]byte(cleanedBody), &result)
	if err != nil {
		return result
	}

	return result
}

func removeInvisibleChars(input string) string {
	re := regexp.MustCompile(`[[:cntrl:]]+`)
	return re.ReplaceAllString(input, "")
}

func fetchGitLink(ac storage.AppDatabaseContext, limit *int) []string {
	repo := repository.NewAllGitLinkRepository(ac)
	linksIter, err := repo.Query()
	if err != nil {
		log.Fatalf("Failed to fetch git links: %v", err)
	}
	links := []string{}
	count := 0
	for link := range linksIter {
		if *limit > 0 && count >= *limit {
			break
		}
		links = append(links, link)
		count++
	}
	return links
}
