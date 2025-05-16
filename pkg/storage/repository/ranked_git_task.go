package repository

import (
	"iter"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/HUSTSecLab/criticality_score/pkg/storage/sqlutil"
)

type RankedGitTask struct {
	GitLink *string
	Nice    *int
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
select git_link, nice
from (
    select git_link, 0 as nice from (
        select git_link from all_gitlinks
        except select git_link from git_files
    )
) union all (
    select git_link, 1 + EXP(EXTRACT(HOUR from (update_time - now()))) as nice from git_files 
    where success = false and update_time < now() - least(pow(2, failed_times), 60) * interval '1 day'
) union all (
    select git_link, 2 + EXP(EXTRACT(DAY from (update_time - now()))) as nice from git_files 
		where update_time < now() - interval '30 days'
) ORDER BY nice LIMIT $1
	`, limit)
}
