package repository

import (
	"fmt"
	"iter"
	"strconv"
	"strings"
	"time"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/HUSTSecLab/criticality_score/pkg/storage/sqlutil"
	"github.com/lib/pq"
)

type GitMetricsRepository interface {
	/** QUERY **/
	Query() (iter.Seq[*GitMetric], error)
	QueryByLink(link string) (*GitMetric, error)

	/** INSERT/UPDATE **/
	// NOTE: update_time will be updated automatically
	InsertOrUpdate(data *GitMetric) error
	// NOTE: update_time will be updated automatically
	// and the data will not copy from old data
	BatchInsertOrUpdate(data []*GitMetric) error

	// successFilter:
	// 0: no filter, 1: success, 2: fail, 3: never success
	//
	QueryGitFiles(linkQuery string, successFilter int, skip int, take int) (iter.Seq[*GitFile], int, error)
	GetGitFileByLink(link string) (*GitFile, error)
	GetGitFilesStatistics() (*GitFileStatisticsResult, error)

	// times will be updated automatically
	InsertOrUpdateGitFile(data *GitFile, success bool) error
	DeleteGitFile(link string) error
}

type GitMetric struct {
	ID               *int64 `generated:"true"`
	GitLink          *string
	CreatedSince     **time.Time
	UpdatedSince     **time.Time
	ContributorCount **int
	CommitFrequency  **float64
	OrgCount         **int
	License          **pq.StringArray
	Language         **pq.StringArray
	CloneValid       **bool
	UpdateTime       **time.Time
}

type GitFile struct {
	GitLink     *string `pk:"true"`
	FilePath    *string
	Success     *bool
	Message     **string
	UpdateTime  **time.Time
	FailedTimes **int
	LastSuccess **time.Time
	TakeTimeMs  **int64
	TakeStorage **int64
}

type GitFileStatisticsResult struct {
	Total        *int
	Success      *int
	Fail         *int
	NeverSuccess *int
}

const GitMetricTableName = "git_metrics"
const GitFilesTableName = "git_files"

type gitmetricsRepository struct {
	ctx storage.AppDatabaseContext
}

// GetGitFileByLink implements GitMetricsRepository.
func (g *gitmetricsRepository) GetGitFileByLink(link string) (*GitFile, error) {
	return sqlutil.QueryCommonFirst[GitFile](g.ctx, "git_files", "WHERE git_link = $1", link)
}

// CountGitFiles implements GitMetricsRepository.
// func (g *gitmetricsRepository) CountGitFiles(linkQuery string) (int, error) {
// 	var cnt int
// 	r := g.ctx.QueryRow("SELECT COUNT(*) FROM git_files WHERE git_link like $1", "%"+linkQuery+"%")
// 	err := r.Scan(&cnt)
// 	return cnt, err
// }

var _ GitMetricsRepository = (*gitmetricsRepository)(nil)

// BatchInsertOrUpdate implements GitMetricsRepository.
func (g *gitmetricsRepository) BatchInsertOrUpdate(data []*GitMetric) error {
	for _, d := range data {
		d.UpdateTime = sqlutil.ToNullable(time.Now())
	}
	return sqlutil.BatchInsert(g.ctx, string(GitMetricTableName), data)
}

// InsertOrUpdate implements GitMetricsRepository.
func (g *gitmetricsRepository) InsertOrUpdate(data *GitMetric) error {
	oldData, err := g.QueryByLink(*data.GitLink)
	if err != nil {
		sqlutil.MergeStruct(oldData, data)
	}
	data.UpdateTime = sqlutil.ToNullable(time.Now())
	return sqlutil.Insert(g.ctx, string(GitMetricTableName), data)
}

// Query implements GitMetricsRepository.
func (g *gitmetricsRepository) Query() (iter.Seq[*GitMetric], error) {
	subQuery := fmt.Sprintf(`(SELECT DISTINCT ON (git_link)
	 * 
	FROM %s
	ORDER BY git_link, id DESC)`, GitMetricTableName)
	return sqlutil.QueryCommon[GitMetric](g.ctx, subQuery, "")
}

// QueryByLink implements GitMetricsRepository.
func (g *gitmetricsRepository) QueryByLink(link string) (*GitMetric, error) {
	return sqlutil.QueryCommonFirst[GitMetric](g.ctx, GitMetricTableName, "WHERE git_link = $1 ORDER BY id DESC", link)
}

// DeleteGitFile implements GitMetricsRepository.
func (g *gitmetricsRepository) DeleteGitFile(link string) error {
	return sqlutil.Delete(g.ctx, GitFilesTableName, &GitFile{GitLink: &link})
}

// InsertOrUpdateGitFile implements GitMetricsRepository.
func (g *gitmetricsRepository) InsertOrUpdateGitFile(data *GitFile, success bool) error {
	if success {
		_, err := g.ctx.Exec(`INSERT INTO `+GitFilesTableName+` (git_link, file_path, success, message, update_time, failed_times, last_success, take_time_ms)
		VALUES ($1, $2, true, $3, $4, 0, $4, $5)
		ON CONFLICT (git_link) DO UPDATE SET file_path = $2, success = true, message = $3, update_time = $4, failed_times = 0, last_success = $4, take_time_ms = $5`,
			data.GitLink, data.FilePath, data.Message, data.UpdateTime, data.TakeTimeMs)
		return err

	} else {
		_, err := g.ctx.Exec(`INSERT INTO `+GitFilesTableName+` (git_link, file_path, success, message, update_time, failed_times, last_success, take_time_ms)
		VALUES ($1, $2, false, $3, $4, 1, NULL, $5)
		ON CONFLICT (git_link) DO UPDATE SET file_path = $2, success = false, message = $3, update_time = $4, failed_times = `+GitFilesTableName+`.failed_times + 1, take_time_ms = $5`,
			data.GitLink, data.FilePath, data.Message, data.UpdateTime, data.TakeTimeMs)
		return err
	}
}

// GetGitFilesStatistics implements GitMetricsRepository.
func (g *gitmetricsRepository) GetGitFilesStatistics() (*GitFileStatisticsResult, error) {
	return sqlutil.QueryFirst[GitFileStatisticsResult](g.ctx, `SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE success = TRUE) AS success,
			COUNT(*) FILTER (WHERE success = FALSE) AS fail,
			COUNT(*) FILTER (WHERE last_success is null) AS never_success
		FROM git_files

	`)
}

// QueryGitFiles implements GitMetricsRepository.
func (g *gitmetricsRepository) QueryGitFiles(linkQuery string, successFilter, skip int, take int) (iter.Seq[*GitFile], int, error) {
	var whereSentences = make([]string, 0)
	var args = make([]any, 0)

	if linkQuery != "" {
		whereSentences = append(whereSentences, "git_link LIKE $"+strconv.Itoa(len(args)+1))
		args = append(args, "%"+linkQuery+"%")
	}

	if successFilter != 0 {
		switch successFilter {
		case 1:
			whereSentences = append(whereSentences, "success = true")
		case 2:
			whereSentences = append(whereSentences, "success = false")
		case 3:
			whereSentences = append(whereSentences, "last_success is null")
		}
	}

	paginationSentence := " OFFSET $" + strconv.Itoa(len(args)+1) + " LIMIT $" + strconv.Itoa(len(args)+2)
	var whereSentence string
	if len(whereSentences) != 0 {
		whereSentence = " WHERE "
		whereSentence += strings.Join(whereSentences, " AND ")
	}
	args = append(args, skip, take)

	var cnt int
	r := g.ctx.QueryRow("SELECT COUNT(*) FROM git_files "+whereSentence, args[0:len(args)-2]...)
	err := r.Scan(&cnt)
	if err != nil {
		return nil, cnt, err
	}

	d, err := sqlutil.QueryCommon[GitFile](g.ctx, "git_files", whereSentence+paginationSentence, args...)
	return d, cnt, err
}

func NewGitMetricsRepository(appDb storage.AppDatabaseContext) GitMetricsRepository {
	return &gitmetricsRepository{ctx: appDb}
}
