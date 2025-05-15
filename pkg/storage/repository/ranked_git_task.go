package repository

import (
	"iter"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/HUSTSecLab/criticality_score/pkg/storage/sqlutil"
)

type RankedGitTask struct {
	GitLink *string
	Type    *int
}

type RankedGitTaskRepository interface {
	/** QUERY **/
	Query(limit int) (iter.Seq[*RankedGitTask], error)
}

type rankedGitTaskRepository struct {
	ctx storage.AppDatabaseContext
}

var _ RankedGitTaskRepository = (*rankedGitTaskRepository)(nil)

func NewRankedGitTaskRepository(ctx storage.AppDatabaseContext) RankedGitTaskRepository {
	return &rankedGitTaskRepository{ctx: ctx}
}

// query implements rankedgittaskrepository.
func (r *rankedGitTaskRepository) Query(limit int) (iter.Seq[*RankedGitTask], error) {
	return sqlutil.Query[RankedGitTask](r.ctx, `
select git_link, type
from (
    select git_link, 0 as type from (
        select git_link from all_gitlinks
        except select git_link from git_files
    ) order by git_link
) union (
    select git_link, 1 as type from git_files 
    where success = false and update_time < now() - least(pow(2, failed_times), 60) * interval '1 day'
    order by update_time desc
) union (
    select git_link, 2 as type from (
        select distinct on (git_link) git_link, update_time, commit_frequency from
        git_metrics order by git_link, id desc
    ) as t
    where update_time < now() - interval '30 days'
) LIMIT $1
	`, limit)
}
