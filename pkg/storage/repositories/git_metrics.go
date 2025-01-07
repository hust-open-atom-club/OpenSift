package repositories

import (
	"fmt"
	"reflect"
	"time"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/lib/pq"
)

type GitMetricsFrom int

type GitMetricsRepository struct {
	AppDb *storage.AppDatabase
}

// const (
// 	GitMetricsFromPacakge   GitMetricsFrom = 0
// 	GitMetricsFromGithub    GitMetricsFrom = 1
// 	GitMetricsFromGitlab    GitMetricsFrom = 2
// 	GitMetricsFromBitbucket GitMetricsFrom = 3
// )

type GitMetrics struct {
	SeqId            *int64          `column:"id"`
	GitLink          *string         `column:"git_link"`
	EcoSystem        *string         `column:"ecosystem"`
	CreatedSince     *time.Time      `column:"created_since"`
	UpdatedSince     *time.Time      `column:"updated_since"`
	ContributorCount *int            `column:"contributor_count"`
	CommitFrequency  *float64        `column:"commit_frequency"`
	DepsDevCount     *int            `column:"depsdev_count"`
	DepsDistro       *string         `column:"deps_distro"`
	OrgCount         *int            `column:"org_count"`
	License          *string         `column:"license"`
	Language         *pq.StringArray `column:"language"`
	CloneValid       *bool           `column:"clone_valid"`
	DepsDevPageRank  *float64        `column:"depsdev_pagerank"`
	Scores           *float64        `column:"scores"`
	IsDeleted        *bool           `column:"is_deleted"`

	// follwing is update by repository layer, any change by caller will be ignored
	UpdateTimeGitMetadata  *time.Time `column:"update_time_git_metadata"`
	UpdateTimeDepsDev      *time.Time `column:"update_time_deps_dev"`
	UpdateTimeDistribution *time.Time `column:"update_time_distribution"`
	UpdateTimeScores       *time.Time `column:"update_time_scores"`
	UpdateTime             *time.Time `column:"update_time"`
}

func NewGitMetricsRepository(appDb *storage.AppDatabase) *GitMetricsRepository {
	return &GitMetricsRepository{AppDb: appDb}
}

func (r *GitMetricsRepository) GetGitMetricsByLinkIncludingDeleted(gitLink string) (*GitMetrics, error) {
	rows, err := getDataFromTable[GitMetrics](r.AppDb, "git_metrics_history", "WHERE git_link = $1 ORDER BY id DESC LIMIT 1", gitLink)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return rows.Next()
}

func (r *GitMetricsRepository) GetGitMetricsByLink(gitLink string) (*GitMetrics, error) {
	rows, err := getDataFromTable[GitMetrics](r.AppDb, "git_metrics_history", "WHERE git_link = $1 and is_deleted = FALSE ORDER BY id DESC LIMIT 1", gitLink)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return rows.Next()
}

func (r *GitMetricsRepository) UpdateGitMetrics(data *GitMetrics) error {
	if data.GitLink == nil {
		return fmt.Errorf("GitLink is required")
	}

	oldData, err := r.GetGitMetricsByLink(*data.GitLink)

	if err != nil {
		return err
	}

	if oldData != nil {
		// use reflection to update the map
		reflectType := reflect.TypeOf(*data)
		dataReflectVal := reflect.ValueOf(data).Elem()
		oldDataReflectVal := reflect.ValueOf(oldData).Elem()

		for i := 0; i < reflectType.NumField(); i++ {
			// field := reflectType.Field(i)
			if dataReflectVal.Field(i).IsNil() && !oldDataReflectVal.Field(i).IsNil() {
				val := oldDataReflectVal.Field(i).Interface()
				dataReflectVal.Field(i).Set(reflect.ValueOf(val))
			}
		}
	}

	if data.EcoSystem != nil || data.CreatedSince != nil || data.UpdatedSince != nil || data.ContributorCount != nil || data.CommitFrequency != nil || data.OrgCount != nil || data.License != nil || data.Language != nil || data.CloneValid != nil {
		t := time.Now()
		data.UpdateTimeGitMetadata = &t
	}

	t := time.Now()
	if data.DepsDistro != nil {
		data.UpdateTimeDistribution = &t
	}

	if data.DepsDevCount != nil || data.DepsDevPageRank != nil {
		data.UpdateTimeDepsDev = &t
	}

	isDeleted := false
	data.IsDeleted = &isDeleted
	data.UpdateTime = &t
	data.SeqId = nil

	return insertDataIntoTable(r.AppDb, "git_metrics_history", data)
}

func (r *GitMetricsRepository) DeleteGitMetricsByLink(gitLink string) error {
	link, _ := r.GetGitMetricsByLink(gitLink)
	if link == nil {
		return fmt.Errorf("GitMetrics not found")
	}

	*link.IsDeleted = true
	link.SeqId = nil
	return insertDataIntoTable(r.AppDb, "git_metrics_history", link)
}
