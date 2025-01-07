package main

import (
	"flag"
	"log"

	scores "github.com/HUSTSecLab/criticality_score/pkg/gen_scores"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	_ "github.com/lib/pq"
)

var flagConfigPath = flag.String("config", "config.json", "path to the config file")
var batchSize = flag.Int("batch", 1000, "batch size")

func main() {
	flag.Parse()
	storage.InitializeDefaultAppDatabase(*flagConfigPath)
	db, err := storage.GetDefaultDatabaseConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer db.Close()

	links, err := scores.FetchAllLinks(db)
	if err != nil {
		log.Fatalf("Failed to fetch git links: %v", err)
	}
	scores.CalculaterepoCount(db)
	packageScore := make(map[string]scores.LinkScore)
	linkCount := make(map[string]map[string]int)
	for repo := range scores.PackageList {
		linkCount[repo] = scores.FetchdLinkCount(repo, db)
	}
	for _, link := range links {
		distro_scores := scores.CalculateDepsdistro(link, linkCount)
		data, err := scores.FetchProjectData(db, link)
		if err != nil {
			log.Printf("Failed to fetch project data for %s: %v", link, err)
			continue
		}
		score := scores.CalculateScore(*data, distro_scores)
		packageScore[link] = scores.LinkScore{
			Distro_scores: distro_scores,
			Score:         score * 100,
		}
	}
	log.Println("Updating database...")
	scores.UpdateScore(db, packageScore, *batchSize)
}
