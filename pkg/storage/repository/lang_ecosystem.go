package repository

import (
	"iter"
	"time"

	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/HUSTSecLab/OpenSift/pkg/storage/sqlutil"
	"github.com/samber/lo"
)

type LangEcoLinkRepository interface {
	/** QUERY **/
	QueryByLink(link string) (iter.Seq[*LangEcosystem], error)
	GetByLinkAndType(link string, typ LangEcosystemType) (*LangEcosystem, error)
	Query() (iter.Seq[*LangEcosystem], error) // Get all LangEcosystem Information in order to calculate the score.

	/** INSERT/UPDATE **/
	// NOTE: update_time will be updated automatically
	InsertOrUpdate(data *LangEcosystem) error
	// NOTE: update_time will be updated automatically
	// and the data will not copy from old data
	BatchInsertOrUpdate(data []*LangEcosystem) error
}

type LangEcosystemType int

const (
	Npm LangEcosystemType = iota
	Go
	Maven
	Pypi
	NuGet
	Cargo
	Others
)

type langEcoLinkRepository struct {
	appDb storage.AppDatabaseContext
}

type LangEcosystem struct {
	ID                *int64 `generated:"true"`
	GitLink           *string
	Type              *LangEcosystemType
	LangEcoImpact     *float64
	Lang_eco_pagerank *float64
	DepCount          *int
	UpdateTime        *time.Time
}

const LangEcosystemTableName = "lang_ecosystems"

var _ LangEcoLinkRepository = (*langEcoLinkRepository)(nil)

func NewLangEcoLinkRepository(appDb storage.AppDatabaseContext) LangEcoLinkRepository {
	return &langEcoLinkRepository{
		appDb: appDb,
	}
}

// Query implements LangEcoLinkRepository.
func (l *langEcoLinkRepository) Query() (iter.Seq[*LangEcosystem], error) {
	return sqlutil.Query[LangEcosystem](l.appDb, `SELECT DISTINCT ON (git_link, type)
		id, git_link, type, lang_eco_impact, lang_eco_pagerank, dep_count, update_time
		FROM lang_ecosystems ORDER BY git_link, type, id DESC`)
}

// BatchInsertOrUpdate implements LangEcoLinkRepository.
func (l *langEcoLinkRepository) BatchInsertOrUpdate(data []*LangEcosystem) error {
	for _, d := range data {
		if d.GitLink == nil || *d.GitLink == "" || d.Type == nil {
			return ErrInvalidInput
		}
	}

	return sqlutil.BatchInsert(l.appDb, LangEcosystemTableName, data)
}

// GetByLinkAndType implements LangEcoLinkRepository.
func (l *langEcoLinkRepository) GetByLinkAndType(link string, typ LangEcosystemType) (*LangEcosystem, error) {
	return sqlutil.QueryCommonFirst[LangEcosystem](l.appDb, LangEcosystemTableName,
		"WHERE git_link = $1 AND type = $2 ORDER BY id DESC", link, typ)
}

// InsertOrUpdate implements LangEcoLinkRepository.
func (l *langEcoLinkRepository) InsertOrUpdate(data *LangEcosystem) error {
	if data.GitLink == nil || *data.GitLink == "" || data.Type == nil {
		return ErrInvalidInput
	}

	oldData, err := l.GetByLinkAndType(*data.GitLink, *data.Type)

	if err != nil {
		sqlutil.MergeStruct(oldData, data)
	}

	data.UpdateTime = lo.ToPtr(time.Now())

	return sqlutil.Insert(l.appDb, LangEcosystemTableName, data)
}

// QueryByLink implements LangEcoLinkRepository.
func (l *langEcoLinkRepository) QueryByLink(link string) (iter.Seq[*LangEcosystem], error) {
	result, err := sqlutil.QueryCommon[LangEcosystem](l.appDb, LangEcosystemTableName, "WHERE git_link = $1", link)
	return result, err
}
