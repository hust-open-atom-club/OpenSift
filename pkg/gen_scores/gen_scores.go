package scores

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
)

type ProjectData struct {
	StarCount        *int
	ForkCount        *int
	CreatedSince     *time.Time
	UpdatedSince     *time.Time
	ContributorCount *int
	CommitFrequency  *float64
	DepsdevCount     *int
	deps_distro      *float64
	Pkg_Manager      *string
	Org_Count        *int
}

type LinkScore struct {
	Distro_scores float64
	Score         float64
}

type UpdateData struct {
	Link         string
	DistroScores float64
	Score        float64
}

// Define weights (αi) and max thresholds (Ti)
var weights = map[string]float64{
	// "star_count":        1,
	// "fork_count":        1,
	"created_since":     1,
	"updated_since":     -1,
	"contributor_count": 2,
	"commit_frequency":  1,
	"depsdev_ratios":    5,
	"deps_distro":       10,
	"org_count":         1,
}

var thresholds = map[string]float64{
	// "star_count":        10000,
	// "fork_count":        5000,
	"created_since":     120, // in months
	"updated_since":     120, // in months
	"contributor_count": 40000,
	"commit_frequency":  1000,
	"depsdev_ratios":    30,
	"deps_distro":       1,
	"org_count":         8400,
}

var PackageManagerData = map[string]int{
	"npm":   3400000,
	"go":    1230000,
	"maven": 636000,
	"pypi":  538000,
	"nuget": 406000,
	"cargo": 155000,
}

var PackageList = map[string]int{
	"debian_packages":   0,
	"arch_packages":     0,
	"nix_packages":      0,
	"homebrew_packages": 0,
	"gentoo_packages":   0,
}

func CalculateDependencyRatio(link, packageType string, linkCount map[string]map[string]int) (float64, error) {
	if _, exist := linkCount[packageType][strings.ToLower(link)]; !exist {
		return 0, nil
	}
	return float64(linkCount[packageType][strings.ToLower(link)]) / float64(PackageList[packageType]), nil
}

func CalculaterepoCount(db *sql.DB) {
	for repo := range PackageList {
		var count int
		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", repo)).Scan(&count)
		if err != nil {
			fmt.Println("Error querying project type:", err)
			return
		}
		PackageList[repo] = count
	}
}

func GetProjectTypeFromDB(link string) string {
	var projectType string
	db, err := storage.GetDefaultDatabaseConnection()
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return ""
	}
	defer db.Close()
	err = db.QueryRow("SELECT ecosystem FROM git_metrics WHERE git_link = $1", link).Scan(&projectType)
	if err != nil {
		fmt.Println("Error querying project type:", err)
		return ""
	}

	return projectType
}

func CalculateScore(data ProjectData, distro_scores float64) float64 {
	score := 0.0
	var createdSinceScore, updatedSinceScore, contributorCountScore, commitFrequencyScore, Org_CountScore float64
	if data.CreatedSince != nil {
		monthsSinceCreation := time.Since(*data.CreatedSince).Hours() / (24 * 30)
		normalized := math.Log(monthsSinceCreation+1) / math.Log(math.Max(monthsSinceCreation, thresholds["created_since"])+1)
		createdSinceScore = weights["created_since"] * normalized
		score += createdSinceScore
	}

	if data.UpdatedSince != nil {
		monthsSinceUpdate := time.Since(*data.UpdatedSince).Hours() / (24 * 30)
		normalized := math.Log(monthsSinceUpdate+1) / math.Log(math.Max(monthsSinceUpdate, thresholds["updated_since"])+1)
		updatedSinceScore = weights["updated_since"] * normalized
		score += updatedSinceScore
	}

	if data.ContributorCount != nil {
		normalized := math.Log(float64(*data.ContributorCount)+1) / math.Log(math.Max(float64(*data.ContributorCount), thresholds["contributor_count"])+1)
		contributorCountScore = weights["contributor_count"] * normalized
		score += contributorCountScore
	}

	if data.CommitFrequency != nil {
		normalized := math.Log(float64(*data.CommitFrequency)+1) / math.Log(math.Max(float64(*data.CommitFrequency), thresholds["commit_frequency"])+1)
		commitFrequencyScore = weights["commit_frequency"] * normalized
		score += commitFrequencyScore
	}
	if data.Org_Count != nil {
		normalized := math.Log(float64(*data.Org_Count)+1) / math.Log(math.Max(float64(*data.Org_Count), thresholds["org_count"])+1)
		Org_CountScore = weights["org_count"] * normalized
		score += Org_CountScore
	}
	if data.Pkg_Manager != nil {
		pkgManager, ok := PackageManagerData[*data.Pkg_Manager]
		if ok && data.DepsdevCount != nil {
			ratios := float64(*data.DepsdevCount) / float64(pkgManager) * 100
			normalized := math.Log(ratios+1) / math.Log(math.Max(ratios, thresholds["depsdev_ratios"])+1)
			score += weights["depsdev_ratios"] * normalized
		}
	}
	normalized := math.Log(distro_scores+1) / math.Log(math.Max(distro_scores, thresholds["deps_distro"])+1)
	score += weights["deps_distro"] * normalized

	var totalnum float64
	for _, weight := range weights {
		totalnum += weight
	}
	return score / totalnum
}

func UpdateScore(db *sql.DB, packageScore map[string]LinkScore, batchSize int) error {
	updates := make([]UpdateData, 0, len(packageScore))

	for link, score := range packageScore {
		updates = append(updates, UpdateData{
			Link:         link,
			DistroScores: float64(score.Distro_scores),
			Score:        float64(score.Score),
		})
	}

	for i := 0; i < len(updates); i += batchSize {
		end := i + batchSize
		if end > len(updates) {
			end = len(updates)
		}
		batch := updates[i:end]

		query := "UPDATE git_metrics SET deps_distro = CASE git_link"
		args := []interface{}{}
		for j, update := range batch {
			query += fmt.Sprintf(" WHEN $%d THEN $%d::double precision ", j*3+1, j*3+2)
			args = append(args, update.Link, update.DistroScores, update.Score)
		}
		query += " END, scores = CASE git_link"
		for j, _ := range batch {
			query += fmt.Sprintf(" WHEN $%d THEN $%d::double precision ", j*3+1, j*3+3)
		}
		query += " END WHERE git_link IN ("
		for j, _ := range batch {
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

func FetchProjectData(db *sql.DB, gitLink string) (*ProjectData, error) {
	row := db.QueryRow("SELECT created_since, updated_since, contributor_count, commit_frequency, depsdev_count, ecosystem, org_count FROM git_metrics WHERE git_link = $1", gitLink)
	var data ProjectData
	err := row.Scan(&data.CreatedSince, &data.UpdatedSince, &data.ContributorCount, &data.CommitFrequency, &data.DepsdevCount, &data.Pkg_Manager, &data.Org_Count)
	if err != nil {
		log.Printf("Failed to fetch data for git link %s: %v", gitLink, err)
		return nil, err
	}
	return &data, nil
}

func CalculateDepsdistro(link string, linkCount map[string]map[string]int) float64 {
	totalRatio := 0.0
	for repo := range PackageList {
		depRatio, err := CalculateDependencyRatio(link, repo, linkCount)
		if err == nil {
			totalRatio += depRatio
		}
	}
	return totalRatio
}

func FetchdLinkCount(repo string, db *sql.DB) map[string]int {
	rows, err := db.Query("SELECT git_link, depends_count FROM " + repo)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	gitLinks := make(map[string]int)
	for rows.Next() {
		var gitLink sql.NullString
		var dependsCount int

		err := rows.Scan(&gitLink, &dependsCount)
		if err != nil {
			log.Fatal(err)
		}

		if gitLink.Valid {
			link := strings.ToLower(gitLink.String)
			if !strings.HasSuffix(link, ".git") {
				link += ".git"
			}

			if _, exist := gitLinks[link]; !exist {
				gitLinks[link] = dependsCount
			}
			gitLinks[link] += dependsCount
		}
	}
	return gitLinks
}

func FetchAllLinks(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT git_link FROM git_metrics")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []string
	for rows.Next() {
		var link string
		if err := rows.Scan(&link); err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, nil
}

func FetchdLinkCountSingle(repo string, link string, db *sql.DB) int {
	url := fmt.Sprintf("SELECT git_link, depends_count FROM %s WHERE git_link = '%s'", repo, link)
	rows, err := db.Query(url)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var gitLink sql.NullString
		var dependsCount int

		err := rows.Scan(&gitLink, &dependsCount)
		if err != nil {
			log.Fatal(err)
		}

		if gitLink.Valid {
			count += dependsCount
		}
	}
	return count
}
