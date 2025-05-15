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
	url "github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/url"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var (
	errUrlNotFound      = errors.New("repo URL not found")
	errWalkRepoFailed   = errors.New("walk repo failed")
	errWalkLogFailed    = errors.New("walk log failed")
	errPathNameNotFound = errors.New("repo pathname not found")
)

type Repo struct {
	Name   string
	Owner  string
	Source string
	URL    string
	//*	Licenses []string
	//* is_maintained bool
	Languages        []string
	Ecosystems       []string
	CreatedSince     time.Time
	UpdatedSince     time.Time
	ContributorCount int
	OrgCount         int
	CommitFrequency  float64
}

func NewRepo() Repo {
	return Repo{
		Name:   parser.UNKNOWN_NAME,
		Owner:  parser.UNKNOWN_OWNER,
		Source: parser.UNKNOWN_SOURCE,
		URL:    parser.UNKNOWN_URL,
		//*	Licenses:         nil,
		Languages:        nil,
		Ecosystems:       nil,
		CreatedSince:     parser.UNKNOWN_TIME,
		UpdatedSince:     parser.UNKNOWN_TIME,
		ContributorCount: parser.UNKNOWN_COUNT,
		OrgCount:         parser.UNKNOWN_COUNT,
		CommitFrequency:  parser.UNKNOWN_FREQUENCY,
	}
}

func GetBlobs(r *git.Repository) (*[]*object.Blob, error) {
	bIter, err := r.BlobObjects()

	if err != nil {
		return nil, err
	}

	blobs := make([]*object.Blob, 0)

	err = bIter.ForEach(func(b *object.Blob) error {
		// fmt.Println(b)
		blobs = append(blobs, b)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &blobs, nil
}

func GetBranches(r *git.Repository) (*[]*plumbing.Reference, error) {
	rIter, err := r.Branches()

	if err != nil {
		return nil, err
	}

	refs := make([]*plumbing.Reference, 0)
	err = rIter.ForEach(func(r *plumbing.Reference) error {
		// fmt.Println(r)
		refs = append(refs, r)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &refs, nil
}

func GetCommits(r *git.Repository) (*[]*object.Commit, error) {
	cIter, err := r.CommitObjects()

	if err != nil {
		return nil, err
	}

	commits := make([]*object.Commit, 0)
	err = cIter.ForEach(func(c *object.Commit) error {
		// fmt.Println(c)
		commits = append(commits, c)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &commits, nil
}

func GetConfig(r *git.Repository) (*gitconfig.Config, error) {
	c, err := r.Config()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func GetObjects(r *git.Repository) (*[]*object.Object, error) {
	oIter, err := r.Objects()

	if err != nil {
		return nil, err
	}

	objs := make([]*object.Object, 0)
	err = oIter.ForEach(func(o object.Object) error {
		// fmt.Println(o)
		objs = append(objs, &o)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &objs, nil
}

func GetReferences(r *git.Repository) (*[]*plumbing.Reference, error) {
	rIter, err := r.References()

	if err != nil {
		return nil, err
	}

	refs := make([]*plumbing.Reference, 0)
	err = rIter.ForEach(func(r *plumbing.Reference) error {
		// fmt.Println(r)
		refs = append(refs, r)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &refs, nil
}

func GetRemotes(r *git.Repository) (*[]*git.Remote, error) {
	remotes, err := r.Remotes()

	if err != nil {
		return nil, err
	}

	return &remotes, nil
}

func GetTags(r *git.Repository) (*[]*object.Tag, error) {
	tIter, err := r.TagObjects()

	if err != nil {
		return nil, err
	}

	tags := make([]*object.Tag, 0)
	err = tIter.ForEach(func(t *object.Tag) error {
		// fmt.Println(t)
		tags = append(tags, t)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &tags, nil
}

func GetTagRefs(r *git.Repository) (*[]*plumbing.Reference, error) {
	rIter, err := r.Tags()

	if err != nil {
		return nil, err
	}

	refs := make([]*plumbing.Reference, 0)
	err = rIter.ForEach(func(r *plumbing.Reference) error {
		// fmt.Println(r)
		refs = append(refs, r)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &refs, nil
}

func GetTrees(r *git.Repository) (*[]*object.Tree, error) {
	tIter, err := r.TreeObjects()

	if err != nil {
		return nil, err
	}

	trees := make([]*object.Tree, 0)
	err = tIter.ForEach(func(t *object.Tree) error {
		// fmt.Println(t)
		trees = append(trees, t)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &trees, nil
}

func GetWorkTree(r *git.Repository) (*git.Worktree, error) {
	wt, err := r.Worktree()

	if err != nil {
		return nil, err
	}

	return wt, nil
}

func GetURL(r *git.Repository) (string, error) {
	//? In most cases, the Remote URLs of Git Fetch and Git Push are the same, but we take the former one
	remotes, err := GetRemotes(r)

	if err != nil {
		logger.Error(err)
		return "", err
	}

	if len(*remotes) == 0 {
		return "", nil
	}

	var u string

	if len((*remotes)[0].Config().URLs) > 0 {
		u = (*remotes)[0].Config().URLs[0]
	}

	for _, remote := range *remotes {
		if remote.Config().Name == parser.DEFAULT_REMOTE_NAME {
			if len(remote.Config().URLs) > 0 {
				u = remote.Config().URLs[0]
				break
			}
		}
	}

	return u, nil
}

func GetLanguages(filename string, filesize int64, l *map[string]int64) {
	v, ok := parser.LANGUAGE_FILENAMES[filename]
	if ok {
		(*l)[v] += filesize
	} else {
		ex := filepath.Ext(filename)
		v, ok = parser.LANGUAGE_EXTENSIONS[ex]
		if ok {
			(*l)[v] += filesize
		}
	}
}

func GetEcosystem(filename string, filesize int64, e *map[string]int64) {
	v, ok := parser.ECOSYSTEM_MAP[filename]
	if ok {
		(*e)[v] += filesize
	}
}

/*
* License Info is now disabled but preserved
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
*/

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

		//! It made sense that this `if`` statement is not necessary but sometimes there are errors
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

	languages := make(map[string]int64, 0)
	ecosystems := make(map[string]int64, 0)

	fIter := tree.Files()

	err = fIter.ForEach(func(f *object.File) error {
		filename := filepath.Base(f.Name)
		filesize := f.Size
		GetLanguages(filename, filesize, &languages)
		GetEcosystem(filename, filesize, &ecosystems)
		/*
			* License Info is now disabled but preserved
			if repo.Licenses == nil {
				if _, ok := parser.LICENSE_FILENAMES[filename]; ok {
					license, err := GetLicense(f)
					if err != nil {
						logger.Error(err)
					} else if license != "" {
						repo.Licenses = []string{license}
					}
				}
			}
		*/
		return nil
	})
	if err != nil {
		return err
	}

	repo.Languages = getTopNKeys(languages)
	repo.Ecosystems = getTopNKeys(ecosystems)

	return nil
}

func (repo *Repo) Show() {
	fmt.Printf(
		"[%v]: %v\n"+
			"[%v]: %v    [%v]: %v    [%v]: %v\n"+
			"[%v]: %v\n"+
			"[%v]: %v\n"+
			//* "[%v]: %v\n"+
			"[%v]: %v\n"+
			"[%v]: %v\n"+
			"[%v]: %v    [%v]: %v\n"+
			"[%v]: %v\n",
		"Repository Name", repo.Name,
		"Source", repo.Source,
		"Owner", repo.Owner,
		//* "License", repo.Licenses,
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

	uu, _ := url.ParseURL(u)

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
