package repositories_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/HUSTSecLab/criticality_score/pkg/storage/repositories"
	"github.com/samber/lo"
)

func TestGithubInsert(t *testing.T) {

	db, err := storage.NewAppDatabase("/home/chengziqiu/Workspace/criticality_score/config.json")

	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}

	repo := repositories.NewGitMetricsRepository(db)

	var error error

	url := "https://github.com/neovim/neovim"

	error = repo.UpdateGitMetrics(&repositories.GitMetrics{
		GitLink:      lo.ToPtr(url),
		CreatedSince: lo.ToPtr(time.Now()),
	})

	if error != nil {
		t.Fatalf("failed to insert: %v", error)
	}

	testTime, _ := time.Parse("2006-01-02", "2021-01-01")

	error = repo.UpdateGitMetrics(&repositories.GitMetrics{
		GitLink:         lo.ToPtr(url),
		CreatedSince:    lo.ToPtr(testTime),
		CommitFrequency: lo.ToPtr(123.0),
	})

	if error != nil {
		t.Fatalf("failed to insert: %v", error)
	}

	d, err := repo.GetGitMetricsByLink(url)
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}

	if *d.CommitFrequency == 123.0 && *d.CreatedSince == testTime {
		t.Fatalf("failed to get: %v", d)
	}
}

func TestGithubQuery(t *testing.T) {

	db, err := storage.NewAppDatabase("/home/chengziqiu/Workspace/criticality_score/config.json")

	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}

	repo := repositories.NewGitMetricsRepository(db)

	url := "https://github.com/neovim/neovim"

	d, err := repo.GetGitMetricsByLink(url)
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}

	fmt.Println(*d.Language)
}
