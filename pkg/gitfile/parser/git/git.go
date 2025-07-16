package git

import (
	parser "github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser"
	"github.com/HUSTSecLab/OpenSift/pkg/logger"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

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
