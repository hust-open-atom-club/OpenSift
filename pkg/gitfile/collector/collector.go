/*
 * @Author: 7erry
 * @Date: 2024-09-29 14:41:35
 * @LastEditTime: 2025-03-03 17:50:44
 * @Description: Collect Git Repositories - Download and Read by go-git
 */

package collector

import (
	"fmt"

	parser "github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser"
	url "github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/url"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
)

// clone or update the repository, and collect metadata
func Collect(u *url.RepoURL, storagePath string) (*gogit.Repository, error) {
	r, err := Clone(u, storagePath)

	if err == gogit.ErrRepositoryAlreadyExists {
		r, err = Update(u, storagePath)
		if err != nil {
			logger.Errorf("Failed to Update %s, %v", u.URL, err)
		}
	} else if err != nil {
		logger.Errorf("Failed to Clone %s, %v", u.URL, err)
	}

	return r, err
}

// mem clone the repository, and collect metadata
func EzCollect(u *url.RepoURL) (*gogit.Repository, error) {
	r, err := MemClone(u)

	if err != nil {
		logger.Errorf("Failed to Clone %s", u.URL)
	}

	return r, err
}

// only clone the repository, if it exists, return error
func Clone(u *url.RepoURL, storagePath string) (*gogit.Repository, error) {
	path := fmt.Sprintf("%s/%s%s", storagePath, u.Resource, u.Pathname)

	r, err := gogit.PlainClone(path, false, &gogit.CloneOptions{
		URL: u.URL,
		// Progress:     os.Stdout,
		SingleBranch: false,
	})

	return r, err
}

// only clone the repository into memory
func MemClone(u *url.RepoURL) (*gogit.Repository, error) {
	r, err := gogit.Clone(memory.NewStorage(), nil, &gogit.CloneOptions{
		URL: u.URL,
		// Progress:     os.Stdout,
		SingleBranch: false,
	})

	return r, err
}

// open the repository
func Open(path string) (*gogit.Repository, error) {
	r, err := gogit.PlainOpen(path)
	return r, err
}

// pull the repository
func Pull(r *gogit.Repository, url string) error {
	wt, err := r.Worktree()

	if err != nil {
		return err
	}

	remotes, err := r.Remotes()

	if err != nil {
		return err
	}

	var remote, u string

	if len(remotes) > 0 {
		remote = (remotes)[0].Config().Name
		urls := (remotes)[0].Config().URLs
		if len(urls) > 0 {
			u = urls[0]
		}
	}

	if remote == "" {
		remote = parser.DEFAULT_REMOTE_NAME
	}

	if u == "" {
		u = url
	}

	err = wt.Pull(&gogit.PullOptions{
		RemoteName:   remote,
		RemoteURL:    u,
		SingleBranch: true,
		Force:        true,
	})

	return err
}

/*
func Fetch(r *gogit.Repository, path string) error {
	var u string

	remotes := git.GetRemotes(r)
	if len(*remotes) > 0 {
		us := (*remotes)[0].Config().URLs
		if len(us) > 0 {
			u = us[0]
		}
	}

	if u == "" {
		u = "https://" + parser.DEFAULT_SOURCE + path
	}

	err := r.Fetch(&gogit.FetchOptions{
		RemoteURL: u,
		RefSpecs:  []gogitconfig.RefSpec{"refs/*:refs/*", "HEAD:ref/heads/HEAD"},
		// Progress:  os.Stdout,
	})
	return err
}
*/

func Update(u *url.RepoURL, storagePath string) (*gogit.Repository, error) {
	path := fmt.Sprintf("%s/%s%s", storagePath, u.Resource, u.Pathname)
	url := u.URL
	r, err := Open(path)

	if err != nil {
		logger.Errorf("Failed to open %s, %v", path, err)
		return r, err
	}

	err = Pull(r, url)

	// err := Fetch(r)
	if err == gogit.NoErrAlreadyUpToDate {
		err = nil
	} else {
		logger.Errorf("Failed to pull %s, %v", path, err)
	}

	return r, err
}

// Check if the url provided is available
func Check(u *url.RepoURL) error {
	_, err := gogit.Clone(memory.NewStorage(), nil, &gogit.CloneOptions{
		URL: u.URL,
		// Progress:     os.Stdout,
		SingleBranch: true,
		Depth:        0,
	})
	return err

}
