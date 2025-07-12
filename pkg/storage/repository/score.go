package repository

import (
	"fmt"
	"iter"
	"time"

	"github.com/HUSTSecLab/OpenSift/pkg/logger"
	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/HUSTSecLab/OpenSift/pkg/storage/sqlutil"
	"github.com/samber/lo"
)

type ScoreRepository interface {
	/** QUERY **/

	Query() (iter.Seq[*Score], error)
	GetByGitLink(distID int64) (*Score, error)
	GetRound() (int, error)

	/** INSERT/UPDATE **/

	// When inserting, make sure all score is properly calculated
	// Any data will not be copied from old data
	// update_time will be updated automatically
	//
	// NOTE: This function only observe thd id field in DistDependencies,
	//       LangEcosystems, GitMetrics
	InsertOrUpdate(score *Score) error
	// When inserting, make sure all score is properly calculated
	// Any data will not be copied from old data
	// update_time will be updated automatically
	//
	// NOTE: This function only observe thd id field in DistDependencies,
	//       LangEcosystems, GitMetrics
	BatchInsertOrUpdate(scores []*Score) error
}

type Score struct {
	ID               *int64 `pk:"true" generated:"true"`
	GitLink          *string
	DistDependencies []*DistDependency `ignore:"true"`
	DistScore        *float64
	LangEcosystems   []*LangEcosystem `ignore:"true"`
	LangScore        *float64
	GitMetrics       []*GitMetric `ignore:"true"`
	GitScore         *float64
	Score            *float64
	UpdateTime       *time.Time
	Round            *int
}

const ScoreTableName = "scores"
const ScoreDistTableName = "scores_dist"
const ScoreLangTableName = "scores_lang"
const ScoreGitTableName = "scores_git"

var _ ScoreRepository = (*scoreRepository)(nil)

type scoreRepository struct {
	ctx storage.AppDatabaseContext
}

// BatchInsertOrUpdate implements ScoreRepository.
func (s *scoreRepository) BatchInsertOrUpdate(scores []*Score) error {
	// TODO: Implement this function use batch method

	// FIXME: This is a temporary solution
	cnt := 0
	for _, score := range scores {
		err := s.InsertOrUpdate(score)
		cnt++
		if cnt%1000 == 0 {
			logger.WithFields(map[string]any{"cnt": cnt}).Info("Inserted scores")
		}

		if err != nil {
			return err
		}
	}
	return nil

	// autoCommitSize := 1000

	// for i := 0; i < len(scores); i += autoCommitSize {
	// 	var sql string
	// 	values := make([]interface{}, 0, len(scores)*5)
	// 	sql = `BEGIN;
	// 	DO $$
	// 	DECLARE
	// 		sid int8;
	// 	BEGIN`

	// 	currentScores := scores[i:min(i+autoCommitSize, len(scores))]
	// 	for _, score := range currentScores {

	// 		sql += `INSERT INTO ` + ScoreTableName + ` (git_link, dist_score, lang_score, git_score, score, update_time) VALUES `
	// 		ph := 1 // placeholder
	// 		sql += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, now()) RETURNING id INTO sid;\n", ph, ph+1, ph+2, ph+3, ph+4)
	// 		values = append(values, score.GitLink, score.DistScore, score.LangScore, score.GitScore, score.Score)
	// 		ph += 5

	// 		// insert DistDependencies
	// 		sql += `INSERT INTO ` + ScoreDistTableName + ` (score_id, dist_id) VALUES `
	// 		for _, score := range scores {
	// 			for _, dist := range score.DistDependencies {
	// 				sql += fmt.Sprintf("($%d, sid),", ph)
	// 				values = append(values, dist.ID)
	// 				ph++
	// 			}
	// 		}

	// 		// insert LangEcosystems
	// 		sql += `INSERT INTO ` + ScoreLangTableName + ` (score_id, lang_id) VALUES `
	// 		for _, score := range scores {
	// 			for _, lang := range score.LangEcosystems {
	// 				sql += fmt.Sprintf("($%d, sid),", ph)
	// 				values = append(values, lang.ID)
	// 				ph++
	// 			}
	// 		}

	// 		// insert GitMetrics
	// 		sql += `INSERT INTO ` + ScoreGitTableName + ` (score_id, git_id) VALUES `
	// 		for _, score := range scores {
	// 			for _, git := range score.GitMetrics {
	// 				sql += fmt.Sprintf("($%d, sid),", ph)
	// 				values = append(values, git.ID)
	// 				ph++
	// 			}
	// 		}
	// 	}

	// 	sql += `END$$;
	// 	COMMIT;`

	// 	_, err := s.ctx.Exec(sql, values...)
	// 	return err
	// }
	// return nil
}

// BatchIntLink implements ScoreRepository.
func (s *scoreRepository) GetByGitLink(distID int64) (*Score, error) {
	return sqlutil.QueryCommonFirst[Score](s.ctx, ScoreTableName,
		`WHERE git_link = $1 ORDER BY DESC ID`,
		distID)
}

// InsertOrUpdate implements ScoreRepository.
func (s *scoreRepository) InsertOrUpdate(score *Score) error {
	score.UpdateTime = lo.ToPtr(time.Now())
	err := sqlutil.Insert(s.ctx, ScoreTableName, score)
	if err != nil {
		return err
	}
	pid := score.ID
	if pid == nil {
		return fmt.Errorf("ID is not generated when inserting")
	}
	id := *pid

	// Insert DistDependencies
	for _, dist := range score.DistDependencies {
		cid := dist.ID
		if cid == nil {
			continue
		}
		_, err := s.ctx.Exec(`INSERT INTO `+ScoreDistTableName+` (score_id, distribution_dependencies_id) VALUES ($1, $2)`, id, *cid)
		if err != nil {
			return err
		}
	}

	// Insert LangEcosystems
	for _, lang := range score.LangEcosystems {
		cid := lang.ID
		if cid == nil {
			continue
		}
		_, err := s.ctx.Exec(`INSERT INTO `+ScoreLangTableName+` (score_id, lang_ecosystems_id) VALUES ($1, $2)`, id, *cid)
		if err != nil {
			return err
		}
	}

	// Insert GitMetrics
	for _, git := range score.GitMetrics {
		cid := git.ID
		if cid == nil {
			continue
		}
		_, err := s.ctx.Exec(`INSERT INTO `+ScoreGitTableName+` (score_id, git_metrics_id) VALUES ($1, $2)`, id, *cid)
		if err != nil {
			return err
		}
	}

	return nil
}

// Query implements ScoreRepository.
func (s *scoreRepository) Query() (iter.Seq[*Score], error) {
	subQuery := `(SELECT DISTINCT ON (git_link) * FROM ` + ScoreTableName + ` ORDER BY git_link, id DESC)`
	return sqlutil.QueryCommon[Score](s.ctx, subQuery, "")
}

// GetRound implements ScoreRepository.
func (s *scoreRepository) GetRound() (int, error) {
	var result int
	row := s.ctx.QueryRow(`SELECT MAX(round) FROM ` + ScoreTableName)
	err := row.Scan(&result)
	return result, err
}

func NewScoreRepository(appDb storage.AppDatabaseContext) ScoreRepository {
	return &scoreRepository{
		ctx: appDb,
	}
}
