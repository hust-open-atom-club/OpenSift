package collector_depsdev

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
)

type DependentInfo struct {
	DependentCount         int `json:"dependentCount"`
	DirectDependentCount   int `json:"directDependentCount"`
	IndirectDependentCount int `json:"indirectDependentCount"`
}

func updateDatabase(db *sql.DB, linkDepCountList map[string][]float64, batchSize int) error {
	var linkDepList [][]string
	for link, counts := range linkDepCountList {
		linkDepList = append(linkDepList, []string{link, fmt.Sprintf("%d", int(counts[1])), fmt.Sprintf("%f", counts[0])})
	}
	for i := 0; i < len(linkDepList); i += batchSize {
		end := i + batchSize
		if end > len(linkDepList) {
			end = len(linkDepList)
		}
		batch := linkDepList[i:end]
		query := "UPDATE git_metrics SET depsdev_count = CASE"
		args := []interface{}{}
		for j, link := range batch {
			depsdevCount, _ := strconv.Atoi(link[1])
			depsdevPagerank, _ := strconv.ParseFloat(link[2], 64)
			query += fmt.Sprintf(" WHEN git_link = $%d THEN $%d::Int", j*3+1, j*3+2)
			args = append(args, link[0], depsdevCount, depsdevPagerank)
		}
		query += " END, depsdev_pagerank = CASE"
		for j, _ := range batch {
			query += fmt.Sprintf(" WHEN git_link = $%d THEN $%d::Float8", j*3+1, j*3+3)
		}
		query += " END WHERE git_link IN ("
		for j := 0; j < len(batch); j++ {
			if j > 0 {
				query += ", "
			}
			query += fmt.Sprintf("$%d", j*3+1)
		}
		query += ")"
		result, err := db.Exec(query, args...)
		if err != nil {
			log.Printf("Error executing query: %v", err)
			return fmt.Errorf("failed to update batch: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Printf("Error retrieving rows affected: %v", err)
			return fmt.Errorf("failed to retrieve affected rows: %w", err)
		}
		log.Printf("Batch [%d - %d]: %d rows updated", i, end, rowsAffected)
	}
	return nil
}

type VersionInfo struct {
	VersionKey struct {
		Version string `json:"version"`
	} `json:"versionKey"`
	PublishedAt time.Time `json:"publishedAt"`
}

type PackageInfo struct {
	Versions []VersionInfo `json:"versions"`
}

func getLatestVersion(repo, projectType string) string {
	ctx := context.Background()

	url := fmt.Sprintf("https://api.deps.dev/v3alpha/systems/%s/packages/%s", projectType, repo)

	req, _ := http.NewRequest("GET", url, nil)
	client := &http.Client{}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		fmt.Println("Error fetching package information:", err)
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result PackageInfo
	json.Unmarshal(body, &result)

	var latestVersion string
	var latestDate time.Time
	for _, version := range result.Versions {
		if version.PublishedAt.After(latestDate) {
			latestDate = version.PublishedAt
			latestVersion = version.VersionKey.Version
		}
	}

	return latestVersion
}

func queryDepsDev(projectType, projectName, version string) int {
	url := fmt.Sprintf("https://api.deps.dev/v3alpha/systems/%s/packages/%s/versions/%s:dependents", projectType, projectName, version)
	resp, err := http.Get(url)
	if err != nil {
		version = getLatestVersion(projectName, projectType)
		url = fmt.Sprintf("https://api.deps.dev/v3alpha/systems/%s/packages/%s/versions/%s:dependents", projectType, projectName, version)
		resp, err = http.Get(url)
		if err != nil {
			fmt.Println("Error fetching package information:", err)
			return 0
		}
		defer resp.Body.Close()
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0
	}

	var info DependentInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		fmt.Println("Error decoding response:", err)
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

func queryDepsName(gitlink string, rdb *redis.Client) map[string][]string {
	depMap := make(map[string][]string)
	if strings.Contains(gitlink, ".git") {
		gitlink = strings.TrimSuffix(gitlink, ".git")
	}
	var repo, name string
	if len(strings.Split(gitlink, "/")) == 5 {
		repo = strings.Split(gitlink, "/")[3]
		name = strings.Split(gitlink, "/")[4]
	}
	url := fmt.Sprintf("https://api.deps.dev/v3alpha/projects/github.com%%2f%s%%2f%s:packageversions", repo, name)
	resp, err := http.Get(url)
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

	latestVersions := make(map[string]map[string]string)
	for _, item := range result.Versions {
		name := item.VersionKey.Name
		version := item.VersionKey.Version
		system := item.VersionKey.System
		if strings.Contains(name, "\u003E") {
			split := strings.Split(name, "\u003E")
			name = split[len(split)-1]
		}
		if strings.Contains(name, "/") {
			split := strings.Split(name, "/")
			name = split[len(split)-1]
		}
		if _, exists := latestVersions[name]; !exists {
			latestVersions[name] = make(map[string]string)
		}
		if currentVersion, exists := latestVersions[name][system]; !exists || version > currentVersion {
			latestVersions[name][system] = version
			depMap[name] = []string{system, version}
			storage.SetKeyValue(rdb, name, gitlink)
		}
	}
	return depMap
}

func Depsdev(configPath string, batchSize int, workerPoolSize int) {
	storage.InitializeDefaultAppDatabase(configPath)
	db, err := storage.GetDefaultDatabaseConnection()
	rdb, _ := storage.InitRedis()
	if err != nil {
		fmt.Errorf("Error initializing database: %v\n", err)
		return
	}
	defer db.Close()
	gitLinks := getGitlink(db)
	pkgMap := make(map[string][]string)
	gitMap := make(map[string][]float64)
	pkgDepMap := make(map[string]int)
	for _, gitlink := range gitLinks {
		depMap := queryDepsName(gitlink, rdb)
		for pkgName, pkgInfo := range depMap {
			pkgDepMap[pkgName] = queryDepsDev(pkgInfo[0], pkgName, pkgInfo[1])
		}
		pkgdepMap := fetchDep(depMap, workerPoolSize)
		for pkgName, pkgInfo := range pkgdepMap {
			pkgMap[pkgName] = pkgInfo
		}
		storage.PersistData(rdb)
	}
	page_rank := calculatePageRank(pkgMap, 100, 0.85)
	var mu sync.Mutex
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, workerPoolSize)
	for pkgName := range pkgMap {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(pkgName string) {
			defer wg.Done()
			defer func() { <-semaphore }()
			key, err := storage.GetKeyValue(rdb, pkgName)
			if err != nil {
				fmt.Println("Error getting key:", err)
				return
			}
			mu.Lock()
			if _, exists := gitMap[key]; !exists {
				gitMap[key] = []float64{0.0, 0.0}
			}
			gitMap[key][0] += page_rank[pkgName]
			gitMap[key][1] += float64(pkgDepMap[pkgName])
			mu.Unlock()
		}(pkgName)
	}
	wg.Wait()
	updateDatabase(db, gitMap, batchSize)
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

func fetchDep(depMap map[string][]string, threadnum int) map[string][]string {
	depMapNew := make(map[string][]string)
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, threadnum)
	var mu sync.Mutex
	for depName, depInfo := range depMap {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(depName string, depInfo []string) {
			defer wg.Done()
			defer func() { <-semaphore }()
			var system, name, version string
			system = depInfo[0]
			name = depName
			version = depInfo[1]
			result := getAndProcessDependencies(system, name, version)
			mu.Lock()
			depMapNew[name] = []string{}
			for _, node := range result.Nodes {
				if node.Relation == "DIRECT" {
					depMapNew[name] = append(depMapNew[name], node.VersionKey.Name)
				}
			}
			mu.Unlock()
		}(depName, depInfo)
	}
	wg.Wait()
	return depMapNew
}

func calculatePageRank(pkgInfoMap map[string][]string, iterations int, dampingFactor float64) map[string]float64 {
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

		for pkgName, deps := range pkgInfoMap {
			depNum := len(deps)
			for _, depName := range deps {
				if _, exists := pkgInfoMap[depName]; exists {
					newPageRank[depName] += dampingFactor * (pageRank[pkgName] / float64(depNum))
				}
			}
		}
		pageRank = newPageRank
	}
	return pageRank
}

func getAndProcessDependencies(system, name, version string) Dependencies {
	var result Dependencies
	url := fmt.Sprintf("https://api.deps.dev/v3alpha/systems/%s/packages/%s/versions/%s:dependencies", system, name, version)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error querying deps.dev:", err)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		version = getLatestVersion(name, system)
		url = fmt.Sprintf("https://api.deps.dev/v3alpha/systems/%s/packages/%s/versions/%s:dependencies", system, name, version)
		resp, err = http.Get(url)
		if err != nil {
			fmt.Println("Error querying deps.dev:", err)
			return result
		}
		defer resp.Body.Close()
	}
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
