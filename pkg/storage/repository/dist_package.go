package repository

import (
	"iter"
	"strconv"
	"strings"

	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/HUSTSecLab/OpenSift/pkg/storage/sqlutil"
)

type DistPackageRepository interface {
	/** QUERY **/

	Query() (iter.Seq[*DistPackage], error)
	QueryWithFilter(confidenceFilter int, linkFilter string, skip, take int) (iter.Seq[*DistPackage], int, error)
	GetByName(name string) (*DistPackage, error)
	GetByGitLink(gitLink string) (iter.Seq[*DistPackage], error)

	/** INSERT/UPDATE **/

	InsertOrUpdate(packageInfo *DistPackage) error
	// NOTE: git_link will be ignored
	Insert(packageInfo *DistPackage) error
	// NOTE: git_link will be ignored
	BatchInsert(packageInfos []*DistPackage) error
	// NOTE: git_link will be ignored
	Update(packageInfos *DistPackage) error

	UpdateGitLink(name string, gitLink *string, confidence *float32) error
	/** DELETE **/
	Delete(name string) error
	DeleteAll() error
}

const DistPackageTableNameAppendix = "_packages"

type DistPackageTablePrefix string

const (
	DistLinkTablePrefixAlpine    DistPackageTablePrefix = "alpine"
	DistLinkTablePrefixArchlinux                        = "arch"
	DistLinkTablePrefixAur                              = "aur"
	DistLinkTablePrefixCentos                           = "centos"
	DistLinkTablePrefixDebian                           = "debian"
	DistLinkTablePrefixDeepin                           = "deepin"
	DistLinkTablePrefixFedora                           = "fedora"
	DistLinkTablePrefixGentoo                           = "gentoo"
	DistLinkTablePrefixHomebrew                         = "homebrew"
	DistLinkTablePrefixNix                              = "nix"
	DistLinkTablePrefixUbuntu                           = "ubuntu"
)

type DistPackage struct {
	Downloads_3m   *int
	Package        *string `pk:"true"`
	HomePage       *string `column:"homepage"`
	Description    *string
	Version        *string
	GitLink        *string
	LinkConfidence **float32
}

type distPackageRepository struct {
	ctx    storage.AppDatabaseContext
	prefix DistPackageTablePrefix
}

var _ DistPackageRepository = (*distPackageRepository)(nil)

// NewDistPackageRepository creates a new DistPackageRepository.
func NewDistPackageRepository(appDb storage.AppDatabaseContext, prefix DistPackageTablePrefix) DistPackageRepository {
	return &distPackageRepository{ctx: appDb, prefix: prefix}
}

// InsertOrUpdate implements DistPackageRepository.
func (d *distPackageRepository) InsertOrUpdate(packageInfo *DistPackage) error {
	return sqlutil.Upsert(d.ctx, string(d.prefix)+DistPackageTableNameAppendix, packageInfo)
}

// BatchInsert implements DistPackageRepository.
func (d *distPackageRepository) BatchInsert(packageInfos []*DistPackage) error {
	return sqlutil.BatchInsert(d.ctx, string(d.prefix)+DistPackageTableNameAppendix, packageInfos)
}

// Delete implements DistPackageRepository.
func (d *distPackageRepository) Delete(name string) error {
	return sqlutil.Delete(d.ctx, string(d.prefix)+DistPackageTableNameAppendix, &DistPackage{Package: &name})
}

// DeleteAll implements DistPackageRepository.
func (d *distPackageRepository) DeleteAll() error {
	_, err := d.ctx.Exec("DELETE FROM " + string(d.prefix) + DistPackageTableNameAppendix)
	return err
}

// GetByGitLink implements DistPackageRepository.
func (d *distPackageRepository) GetByGitLink(gitLink string) (iter.Seq[*DistPackage], error) {
	return sqlutil.QueryCommon[DistPackage](d.ctx, string(d.prefix)+DistPackageTableNameAppendix, "WHERE git_link = $1", gitLink)
}

// GetByName implements DistPackageRepository.
func (d *distPackageRepository) GetByName(name string) (*DistPackage, error) {
	return sqlutil.QueryCommonFirst[DistPackage](d.ctx, string(d.prefix)+DistPackageTableNameAppendix, "WHERE package = $1", name)
}

// Insert implements DistPackageRepository.
func (d *distPackageRepository) Insert(packageInfo *DistPackage) error {
	if packageInfo.Package == nil || *packageInfo.Package == "" {
		return ErrInvalidInput
	}
	packageInfo.GitLink = nil

	return sqlutil.Insert(d.ctx, string(d.prefix)+DistPackageTableNameAppendix, packageInfo)
}

// Query implements DistPackageRepository.
func (d *distPackageRepository) Query() (iter.Seq[*DistPackage], error) {
	return sqlutil.QueryCommon[DistPackage](d.ctx, string(d.prefix)+DistPackageTableNameAppendix, "")
}

// Update implements DistPackageRepository.
func (d *distPackageRepository) Update(packageInfos *DistPackage) error {
	if packageInfos.Package == nil || *packageInfos.Package == "" {
		return ErrInvalidInput
	}
	packageInfos.GitLink = nil

	return sqlutil.Update(d.ctx, string(d.prefix)+DistPackageTableNameAppendix, packageInfos)
}

// UpdateGitLink implements DistPackageRepository.
func (d *distPackageRepository) UpdateGitLink(name string, gitLink *string, confidence *float32) error {
	_, err := d.ctx.Exec("UPDATE "+string(d.prefix)+DistPackageTableNameAppendix+" SET git_link = $1, link_confidence = $2 WHERE package = $3", gitLink, confidence, name)
	return err
}

func (d *distPackageRepository) QueryWithFilter(confidenceFilter int, linkFilter string, skip, take int) (iter.Seq[*DistPackage], int, error) {
	query := "SELECT * FROM " + string(d.prefix) + DistPackageTableNameAppendix
	cntQuery := "SELECT COUNT(*) FROM " + string(d.prefix) + DistPackageTableNameAppendix
	whereClauses := []string{}
	args := []any{}
	placeHolder := 1

	if confidenceFilter == 1 {
		whereClauses = append(whereClauses, "link_confidence <> 1")
	} else if confidenceFilter == 2 {
		whereClauses = append(whereClauses, "link_confidence = 1")
	}

	if linkFilter != "" {
		whereClauses = append(whereClauses, "git_link LIKE '%' || $"+strconv.Itoa(placeHolder)+" || '%'")
		placeHolder++
		args = append(args, linkFilter)
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
		cntQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	var cnt int
	err := d.ctx.QueryRow(cntQuery, args...).Scan(&cnt)
	if err != nil {
		return nil, 0, err
	}

	query += " ORDER BY package LIMIT $" + strconv.Itoa(placeHolder) + " OFFSET $" + strconv.Itoa(placeHolder+1)
	args = append(args, take, skip)
	placeHolder += 2

	res, err := sqlutil.Query[DistPackage](d.ctx, query, args...)

	return res, cnt, err
}
