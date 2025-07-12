package repository

import (
	"iter"
	"time"

	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/HUSTSecLab/OpenSift/pkg/storage/sqlutil"
	"github.com/samber/lo"
)

const DistDependencyTableName = "distribution_dependencies"

type DistDependencyRepository interface {
	/** QUERY **/

	Query() (iter.Seq[*DistDependency], error) // Query all distribution information.
	QueryByType(distType int) (iter.Seq[*DistDependency], error)
	GetByLink(packageName string, distType int) (*DistDependency, error)
	QueryDistCountByType(distType DistType) (int, error) // Get the total number of packages in a Distro.

	/** INSERT/UPDATE **/
	// update_time will be updated automatically
	InsertOrUpdate(packageInfo *DistDependency) error
}

type distLinkRepository struct {
	ctx storage.AppDatabaseContext
}

var _ DistDependencyRepository = (*distLinkRepository)(nil)

type DistType int

const (
	Debian DistType = iota
	Arch
	Homebrew
	Nix
	Alpine
	Centos
	Aur
	Deepin
	Fedora
	Gentoo
	Ubuntu
	Other
)

type DistDependency struct {
	ID           *int64 `generated:"true"`
	GitLink      *string
	Type         *DistType
	DepImpact    *float64
	DepCount     *int
	PageRank     *float64
	UpdateTime   *time.Time
	Downloads_3m *int
}

func NewDistDependencyRepository(appDb storage.AppDatabaseContext) DistDependencyRepository {
	return &distLinkRepository{ctx: appDb}
}

// Query implements DistributionDependencyRepository.
func (r *distLinkRepository) Query() (iter.Seq[*DistDependency], error) {
	return sqlutil.Query[DistDependency](r.ctx, `SELECT DISTINCT ON (git_link, "type") id, git_link, type, dep_impact, dep_count, page_rank, update_time FROM distribution_dependencies ORDER BY git_link, "type", id DESC`)
}

// QueryDistCountByType implements DistributionDependencyRepository.
func (r *distLinkRepository) QueryDistCountByType(distType DistType) (int, error) {
	var tableName string
	switch distType {
	case Debian:
		tableName = "debian_packages"
	case Arch:
		tableName = "arch_packages"
	case Homebrew:
		tableName = "homebrew_packages"
	case Nix:
		tableName = "nix_packages"
	case Alpine:
		tableName = "alpine_packages"
	case Centos:
		tableName = "centos_packages"
	case Aur:
		tableName = "aur_packages"
	case Deepin:
		tableName = "deepin_packages"
	case Fedora:
		tableName = "fedora_packages"
	case Gentoo:
		tableName = "gentoo_packages"
	case Ubuntu:
		tableName = "ubuntu_packages"
	default:
		return 0, ErrInvalidInput
	}

	var result int
	row := r.ctx.QueryRow(`SELECT COUNT(*) FROM ` + tableName)
	err := row.Scan(&result)
	return result, err
}

// GetByLink implements DistributionDependencyRepository.
func (r *distLinkRepository) GetByLink(packageName string, distType int) (*DistDependency, error) {
	return sqlutil.QueryCommonFirst[DistDependency](r.ctx, DistDependencyTableName,
		`WHERE git_link = $1 and type = $2 ORDER BY id DESC`, packageName, distType)
}

// InsertOrUpdate implements DistributionDependencyRepository.
func (r *distLinkRepository) InsertOrUpdate(packageInfo *DistDependency) error {
	if packageInfo.GitLink == nil || packageInfo.Type == nil {
		return ErrInvalidInput
	}

	packageInfo.UpdateTime = lo.ToPtr(time.Now())

	oldInfo, err := r.GetByLink(*packageInfo.GitLink, int(*packageInfo.Type))
	if err != nil {
		return err
	}

	if oldInfo == nil {
		return sqlutil.Insert(r.ctx, DistDependencyTableName, packageInfo)
	} else {
		sqlutil.MergeStruct(oldInfo, packageInfo)
		return sqlutil.Insert(r.ctx, DistDependencyTableName, packageInfo)
	}
}

// QueryByType implements DistributionDependencyRepository.
func (r *distLinkRepository) QueryByType(distType int) (iter.Seq[*DistDependency], error) {
	return sqlutil.QueryCommon[DistDependency](r.ctx, DistDependencyTableName,
		"where type = $1", distType)
}
