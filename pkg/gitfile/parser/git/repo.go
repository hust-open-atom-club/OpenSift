/*
 * @Author: 7erry
 * @Date: 2024-09-29 14:41:35
 * @LastEditTime: 2025-01-09 15:37:17
 * @Description: Parse Git Repositories to collect necessary metrics
 */

package git

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	parser "github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
	url "github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/url"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/licensecheck"
)

var (
	errUrlNotFound      = errors.New("repo URL not found")
	errWalkRepoFailed   = errors.New("walk repo failed")
	errWalkLogFailed    = errors.New("walk log failed")
	errPathNameNotFound = errors.New("repo pathname not found")
)

type Repo struct {
	Name    string
	Owner   string
	Source  string
	URL     string
	License []string
	//* is_maintained bool
	Languages        []string
	Ecosystems       []string
	CreatedSince     time.Time
	UpdatedSince     time.Time
	ContributorCount int
	OrgCount         int
	CommitFrequency  float64
	EcoDeps          map[*langeco.Package]*langeco.Dependencies
}

func NewRepo() Repo {
	return Repo{
		Name:             parser.UNKNOWN_NAME,
		Owner:            parser.UNKNOWN_OWNER,
		Source:           parser.UNKNOWN_SOURCE,
		URL:              parser.UNKNOWN_URL,
		License:          nil,
		Languages:        nil,
		Ecosystems:       nil,
		CreatedSince:     parser.UNKNOWN_TIME,
		UpdatedSince:     parser.UNKNOWN_TIME,
		ContributorCount: parser.UNKNOWN_COUNT,
		OrgCount:         parser.UNKNOWN_COUNT,
		CommitFrequency:  parser.UNKNOWN_FREQUENCY,
	}
}

func GetLicense(f *object.File) (string, error) {
	text, err := f.Contents()
	if err != nil {
		return "", err
	}
	cov := licensecheck.Scan([]byte(text))
	if len(cov.Match) == 0 {
		return "", nil
	}

	license := cov.Match[0].ID

	return license, nil
}

func getTopNKeys(m map[string]int64) []string {
	keys := make([]string, 0, len(m))

	for key := range m {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return m[keys[i]] > m[keys[j]]
	})

	if len(keys) > parser.TOP_N {
		return keys[:parser.TOP_N]
	}
	return keys
}

func (repo *Repo) WalkLog(r *git.Repository) error {
	cIter, err := r.Log(&git.LogOptions{
		//* From:  ref.Hash(),
		All:   true,
		Since: &parser.BEGIN_TIME,
		Until: &parser.END_TIME,
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return err
	}

	contributors := make(map[string]int, 0)
	orgs := make(map[string]int, 0)
	var commit_count float64 = 0

	latest_commit, err := cIter.Next()
	if err != nil {
		return err
	}

	author := fmt.Sprintf("%s(%s)", latest_commit.Author.Name, latest_commit.Author.Email)
	e := strings.Split(latest_commit.Author.Email, "@")
	org := e[len(e)-1]

	repo.UpdatedSince = latest_commit.Committer.When
	contributors[author]++
	orgs[org]++

	if latest_commit.Author.When.After(parser.LAST_YEAR) {
		commit_count++
	}

	created_since := latest_commit.Committer.When

	err = cIter.ForEach(func(c *object.Commit) error {
		author := fmt.Sprintf("%s(%s)", c.Author.Name, c.Author.Email)
		e = strings.Split(c.Author.Email, "@")
		org := e[len(e)-1]

		//! It made sense that this `if` statement is not necessary but sometimes there are errors
		if created_since.After(c.Committer.When) {
			created_since = c.Committer.When
		}
		contributors[author]++
		orgs[org]++

		if created_since.After(parser.LAST_YEAR) {
			commit_count++
		}

		return nil
	})

	if err != nil {
		return err
	}

	repo.CreatedSince = created_since
	repo.ContributorCount = len(contributors)
	repo.OrgCount = len(orgs)
	repo.CommitFrequency = commit_count / 52

	return nil
}

func (repo *Repo) WalkRepo(r *git.Repository) error {

	ref, err := r.Head()
	if err != nil {
		return err
	}

	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return err
	}

	tree, err := commit.Tree()
	if err != nil {
		return err
	}

	fIter := tree.Files()
	led := NewLangEcoDeps(repo)

	err = fIter.ForEach(func(f *object.File) error {
		led.Parse(f)
		filename := filepath.Base(f.Name)
		if repo.License == nil {
			if _, ok := parser.LICENSE_FILENAMES[filename]; ok {
				license, err := GetLicense(f)
				if err != nil {
					logger.Error(err)
				} else if license != "" {
					repo.License = []string{license}
				}
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	repo.Languages = getTopNKeys(led.languages)
	repo.Ecosystems = getTopNKeys(led.ecosystems)
	repo.EcoDeps = led.dependencies

	return nil
}

func (repo *Repo) Show() {
	fmt.Printf(
		"[%v]: %v\n"+
			"[%v]: %v    [%v]: %v    [%v]: %v\n"+
			"[%v]: %v\n"+
			"[%v]: %v\n"+
			"[%v]: %v\n"+
			"[%v]: %v\n"+
			"[%v]: %v\n"+
			"[%v]: %v    [%v]: %v\n"+
			"[%v]: %v\n",
		"Repository Name", repo.Name,
		"Source", repo.Source,
		"Owner", repo.Owner,
		"License", repo.License,
		"URL", repo.URL,
		"Languages", repo.Languages,
		"Ecosystems", repo.Ecosystems,
		"Created at", repo.CreatedSince,
		"Updated at", repo.UpdatedSince,
		"Contributor Count", repo.ContributorCount,
		"Organization Count", repo.OrgCount,
		"Commit Frequency", repo.CommitFrequency,
	)
}

func ParseRepo(r *git.Repository) (*Repo, error) {

	repo := NewRepo()

	u, err := GetURL(r)
	if err != nil {
		logger.Errorf("Failed to Get RepoURL for %v", err)
		return nil, err
	}
	if u == "" {
		return nil, errUrlNotFound
	}

	repo.URL = u

	uu, err := url.ParseURL(u)
	if err != nil {
		logger.Errorf("Failed to Parse RepoURL for %v", err)
		return nil, fmt.Errorf("failed to parse repo URL: %w", err)
	}

	if uu.Pathname == "" || uu.Resource == "" {
		return nil, errPathNameNotFound
	}

	path := strings.Split(uu.Pathname, "/")
	repo.Name = strings.Split(path[len(path)-1], ".")[0]
	repo.Owner = path[len(path)-2]
	repo.Source = uu.Resource

	err = repo.WalkRepo(r)
	if err != nil {
		logger.Errorf("Failed to Walk Repo for %v", err)
		return nil, errWalkRepoFailed
	}

	err = repo.WalkLog(r)
	if err != nil {
		logger.Errorf("Failed to Walk Log for %v", err)
		return nil, errWalkLogFailed
	}

	return &repo, nil
}
