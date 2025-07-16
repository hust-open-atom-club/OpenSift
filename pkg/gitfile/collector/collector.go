package collector

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/HUSTSecLab/OpenSift/pkg/config"
	parser "github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser"
	url "github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser/url"
	"github.com/HUSTSecLab/OpenSift/pkg/logger"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
)

// clone or update the repository, and collect metadata
func Collect(u *url.RepoURL, path string, progress io.Writer) (*gogit.Repository, error) {
	_, err := Open(path)
	if err != nil { // not exsists
		r, err := CloneOutProcess(u, path, progress)
		if err != nil {
			logger.Errorf("Failed to Clone %s, %v", u.URL, err)
		}
		return r, err
	} else {
		r, err := Update(u, path, progress)
		if err != nil {
			logger.Errorf("Failed to Update %s, %v", u.URL, err)
		}
		return r, err
	}
}

// mem clone the repository, and collect metadata
func EzCollect(u *url.RepoURL) (*gogit.Repository, error) {
	r, err := MemClone(u)

	if err != nil {
		logger.Errorf("Failed to Clone %s", u.URL)
	}

	return r, err
}

// clone use system git
func CloneOutProcess(u *url.RepoURL, path string, progress io.Writer) (*gogit.Repository, error) {
	// get parent path of the path
	tmpPath := filepath.Dir(path)
	if config.GetGitStoragePath() != "" {
		tmpPath = config.GetGitStoragePath()
	}

	err := os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return nil, err
	}
	tmpDir, err := os.MkdirTemp(
		tmpPath,
		"clone-tmp-",
	)
	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(tmpDir)

	cmd := exec.Command("git", "clone", "--mirror", "--progress", u.URL, tmpDir)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	cmd.Stderr = progress
	cmd.Stdout = progress
	devnull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0755)
	cmd.Stdin = devnull
	err = cmd.Run()

	if err != nil {
		return nil, err
	}

	// If the path exists, remove it before moving the tmpDir
	if _, err := os.Stat(path); err == nil {
		if err := os.RemoveAll(path); err != nil {
			return nil, err
		}
	}

	if err = os.Rename(tmpDir, path); err != nil {
		return nil, err
	}

	return Open(path)
}

// Very slow, USE CloneOutProcess Instead
// only clone the repository, if it exists, return error
func Clone(u *url.RepoURL, path string, progress io.Writer) (*gogit.Repository, error) {
	// get parent path of the path
	tmpPath := filepath.Dir(path)
	if config.GetGitStoragePath() != "" {
		tmpPath = config.GetGitStoragePath()
	}

	err := os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return nil, err
	}
	tmpDir, err := os.MkdirTemp(
		tmpPath,
		"clone-tmp-",
	)
	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(tmpDir)

	_, err = gogit.PlainClone(tmpDir, false, &gogit.CloneOptions{
		URL:          u.URL,
		Progress:     progress,
		SingleBranch: false,
		Mirror:       true,
		//* NoCheckout: true,
	})

	if err != nil {
		return nil, err
	}

	// If the path exists, remove it before moving the tmpDir
	if _, err := os.Stat(path); err == nil {
		if err := os.RemoveAll(path); err != nil {
			return nil, err
		}
	}

	if err = os.Rename(tmpDir, path); err != nil {
		return nil, err
	}

	return Open(path)
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

// ToDo Check if Pull needs fixing with Clone adapting to be able to rollback when failed to clone
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
		RemoteName: remote,
		RemoteURL:  u,
		//* SingleBranch: true,
		//* Force: true,
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

func Update(u *url.RepoURL, path string, progress io.Writer) (*gogit.Repository, error) {
	url := u.URL
	r, err := Open(path)
	if err != nil {
		logger.Errorf("Failed to open %s, %v", path, err)
		return r, err
	}

	remoteRefs, err := r.Remotes()
	if err != nil {
		return r, err
	}

	for _, remoteRef := range remoteRefs {
		err := remoteRef.Fetch(&gogit.FetchOptions{
			RemoteURL: url,
			Progress:  progress,
		})
		if err != nil && err != gogit.NoErrAlreadyUpToDate {
			return r, err
		}
	}

	return r, nil

	// Following is old worktree-style

	// err = Pull(r, url)

	// // err := Fetch(r)
	// if err == gogit.NoErrAlreadyUpToDate {
	// 	err = nil
	// } else {
	// 	logger.Errorf("Failed to pull %s, %v", path, err)
	// }

	// return r, err
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
