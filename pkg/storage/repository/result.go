package repository

import (
	"iter"
	"time"

	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/HUSTSecLab/OpenSift/pkg/storage/sqlutil"
	"github.com/lib/pq"
)

type ResultRepository interface {
	/** QUERY **/
	CountByLink(search string) (int, error)
	QueryByLink(search string, skip int, take int) (iter.Seq[*Result], error)
	CountHistoriesByLink(link string) (int, error)
	QueryHistoriesByLink(link string, skip int, take int) (iter.Seq[*Result], error)
	GetByScoreID(scoreID int) (*Result, error)
	QueryGitDetailsByScoreID(scoreID int) (iter.Seq[*ResultGitDetail], error)
	QueryLangDetailsByScoreID(scoreID int) (iter.Seq[*ResultLangDetail], error)
	QueryDistDetailsByScoreID(scoreID int) (iter.Seq[*ResultDistDetail], error)
	QueryRankingCache(skip int, take int) (iter.Seq[*RankingResult], error)
	MakeRankingCache() error
}

type Result struct {
	GitLink    *string
	ScoreID    **int
	DistScore  **float64
	LangScore  **float64
	GitScore   **float64
	Score      **float64
	UpdateTime **time.Time
}

type RankingResult struct {
	GitLink    *string
	ScoreID    **int
	DistScore  **float64
	LangScore  **float64
	GitScore   **float64
	Score      **float64
	UpdateTime **time.Time
	Ranking    *int
}

type ResultGitDetail struct {
	License          **pq.StringArray
	Language         **pq.StringArray
	CommitFrequency  **float64
	CreatedSince     **time.Time
	UpdatedSince     **time.Time
	OrgCount         **int
	ContributorCount **int
	UpdateTime       **time.Time
}

type ResultLangDetail struct {
	Type          **int
	LangEcoImpact **float64
	DepCount      **int
	UpdateTime    **time.Time
}

type ResultDistDetail struct {
	Type       **int
	Count      **int
	Impact     **float64
	PageRank   **float64
	UpdateTime **time.Time
}

type resultRepository struct {
	ctx storage.AppDatabaseContext
}

// QueryRanking implements ResultRepository.
func (r *resultRepository) QueryRankingCache(skip int, take int) (iter.Seq[*RankingResult], error) {
	rows, err := sqlutil.Query[RankingResult](r.ctx, `select * from rankings_cache limit $1 offset $2`, take, skip)
	return rows, err
}

func (r *resultRepository) MakeRankingCache() error {
	_, err := r.ctx.Exec(`DROP TABLE IF EXISTS rankings_cache_tmp;
	CREATE TABLE rankings_cache_tmp AS
	SELECT * FROM rankings;
	DROP TABLE IF EXISTS rankings_cache;
	ALTER TABLE rankings_cache_tmp RENAME TO rankings_cache;
	`)
	return err
}

// QueryHistoriesByLink implements ResultRepository.
func (r *resultRepository) QueryHistoriesByLink(link string, skip int, take int) (iter.Seq[*Result], error) {
	rows, err := sqlutil.Query[Result](r.ctx, `select ag.git_link as git_link,
		s.id as score_id,
		s.dist_score as dist_score,
		s.lang_score as lang_score,
		s.git_score as git_score,
		s.score as score,
		s.update_time as update_time
	from all_gitlinks_cache ag
	left join scores s on ag.git_link = s.git_link
	where ag.git_link = $1 order by s.id desc limit $2 offset $3
	`, link, take, skip)
	return rows, err
}

// CountByLink implements ResultRepository.
func (r *resultRepository) CountByLink(search string) (int, error) {
	row := r.ctx.QueryRow(`select count(*) from all_gitlinks_cache where git_link like $1`, "%"+search+"%")
	var count int
	err := row.Scan(&count)
	return count, err
}

// CountHistoriesByLink implements ResultRepository.
func (r *resultRepository) CountHistoriesByLink(link string) (int, error) {
	row := r.ctx.QueryRow(`select count(*) from scores where git_link = $1`, link)
	var count int
	err := row.Scan(&count)
	return count, err
}

// QueryDistDetailsByScoreID implements ResultRepository.
func (r *resultRepository) QueryDistDetailsByScoreID(scoreID int) (iter.Seq[*ResultDistDetail], error) {
	return sqlutil.Query[ResultDistDetail](r.ctx, `select
		dd.type as type,
		dd.dep_count as count,
		dd.dep_impact as impact,
		dd.page_rank as page_rank,
		dd.update_time as update_time
	from scores_dist sd
	left join distribution_dependencies dd on sd.distribution_dependencies_id = dd.id
	where sd.score_id = $1`, scoreID)
}

// QueryGitDetailsByScoreID implements ResultRepository.
func (r *resultRepository) QueryGitDetailsByScoreID(scoreID int) (iter.Seq[*ResultGitDetail], error) {
	return sqlutil.Query[ResultGitDetail](r.ctx, `select
		gm.license as license,
		gm.language as language,
		gm.commit_frequency as commit_frequency,
		gm.created_since as created_since,
		gm.updated_since as updated_since,
		gm.org_count as org_count,
		gm.contributor_count as contributor_count,
		gm.update_time as update_time
	from scores_git sg
	left join git_metrics gm on sg.git_metrics_id = gm.id
	where sg.score_id = $1`, scoreID)
}

// QueryLangDetailsByScoreID implements ResultRepository.
func (r *resultRepository) QueryLangDetailsByScoreID(scoreID int) (iter.Seq[*ResultLangDetail], error) {
	return sqlutil.Query[ResultLangDetail](r.ctx, `select
		le.type as type,
		le.lang_eco_impact as lang_eco_impact,
		le.dep_count as dep_count,
		le.update_time as update_time
	from scores_lang sl
	left join lang_ecosystems le on sl.lang_ecosystems_id = le.id
	where sl.score_id = $1`, scoreID)
}

// QueryWithCountByLink implements ResultRepository.
func (r *resultRepository) QueryByLink(search string, skip int, take int) (iter.Seq[*Result], error) {
	rows, err := sqlutil.Query[Result](r.ctx, `select * from (
		select distinct on (ag.git_link)
			ag.git_link as git_link,
			s.id as score_id,
			s.dist_score as dist_score,
			s.lang_score as lang_score,
			s.git_score as git_score,
			s.score as score,
			s.update_time as update_time
		from all_gitlinks_cache ag
		left join scores s on ag.git_link = s.git_link
		where ag.git_link like $1
		order by ag.git_link, s.id desc) as t
	order by score desc nulls last
	limit $2 offset $3`, "%"+search+"%", take, skip)
	return rows, err
}

// GetByScoreID implements ResultRepository.
func (r *resultRepository) GetByScoreID(scoreID int) (*Result, error) {
	row, err := sqlutil.QueryFirst[Result](r.ctx, `select 
		ag.git_link as git_link,
		s.id as score_id,
		s.dist_score as dist_score,
		s.lang_score as lang_score,
		s.git_score as git_score,
		s.score as score,
		s.update_time as update_time
	from all_gitlinks_cache ag
	left join scores s on ag.git_link = s.git_link
	where s.id = $1
	`, scoreID)
	return row, err
}

var _ ResultRepository = (*resultRepository)(nil)

func NewResultRepository(appDb storage.AppDatabaseContext) ResultRepository {
	return &resultRepository{ctx: appDb}
}
