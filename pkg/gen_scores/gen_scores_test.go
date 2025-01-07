package scores

import (
	"flag"
	"fmt"
	"log"
	"testing"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	_ "github.com/lib/pq"
)

var flagConfigPath = flag.String("config", "config.json", "path to the config file")

func TestCalculateScore(t *testing.T) {
	fmt.Println("Testing CalculateScore")
	flag.Parse()
	storage.InitializeDefaultAppDatabase(*flagConfigPath)
	db, err := storage.GetDefaultDatabaseConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer db.Close()
	CalculaterepoCount(db)
	var packageScore = make(map[string]LinkScore)
	links := []string{
		"https://github.com/gcc-mirror/gcc.git",
	}
	for _, link := range links {
		linkCount := make(map[string]map[string]int)
		for repo := range PackageList {
			linkCount[repo] = FetchdLinkCount(repo, db)
		}
		fmt.Println(linkCount["debian_packages"]["https://github.com/gcc-mirror/gcc.git"])
		distro_scores := CalculateDepsdistro(link, linkCount)
		data, err := FetchProjectData(db, link)
		if err != nil {
			log.Printf("Failed to fetch project data for %s: %v", link, err)
			return
		}
		score := CalculateScore(*data, distro_scores) * 100
		packageScore[link] = LinkScore{Distro_scores: distro_scores, Score: score}
	}
	fmt.Println(packageScore)
}
