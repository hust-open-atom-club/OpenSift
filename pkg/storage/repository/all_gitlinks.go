package repository

import (
	"iter"

	"github.com/HUSTSecLab/OpenSift/pkg/storage"
)

type AllGitLinkRepository interface {
	/** QUERY **/
	Query() (iter.Seq[string], error)
	QueryByLink(search string) (iter.Seq[string], error)
	QueryCache() (iter.Seq[string], error)
	MakeCache() error
}

type allGitLinkRepository struct {
	ctx storage.AppDatabaseContext
}

// QueryByLink implements AllGitLinkRepository.
func (a *allGitLinkRepository) QueryByLink(search string) (iter.Seq[string], error) {
	return gitlinksQuery(a.ctx, "SELECT git_link FROM all_gitlinks WHERE git_link LIKE $1", search)
}

// MakeCache implements AllGitLinkRepository.
func (a *allGitLinkRepository) MakeCache() error {
	_, err := a.ctx.Exec(`DROP TABLE IF EXISTS all_gitlinks_cache;
	CREATE TABLE all_gitlinks_cache AS SELECT * FROM all_gitlinks;
	`)
	return err
}

// QueryCache implements AllGitLinkRepository.
func (a *allGitLinkRepository) QueryCache() (iter.Seq[string], error) {
	return gitlinksQuery(a.ctx, "SELECT git_link FROM all_gitlinks_cache")
}

var _ AllGitLinkRepository = (*allGitLinkRepository)(nil)

func NewAllGitLinkRepository(appDb storage.AppDatabaseContext) AllGitLinkRepository {
	return &allGitLinkRepository{ctx: appDb}
}

func gitlinksQuery(ctx storage.AppDatabaseContext, query string, args ...interface{}) (iter.Seq[string], error) {
	rows, err := ctx.Query(query, args...)

	if err != nil {
		return nil, err
	}

	return func(yield func(string) bool) {
		defer rows.Close()
		for rows.Next() {
			var link *string
			err := rows.Scan(&link)
			if err != nil {
				return
			}
			if link == nil {
				continue
			}

			if !yield(*link) {
				return
			}
		}
	}, nil
}

// Query implements AllLinkRepository.
func (a *allGitLinkRepository) Query() (iter.Seq[string], error) {
	return gitlinksQuery(a.ctx, "SELECT git_link FROM all_gitlinks")
}
