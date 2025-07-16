package repository

import (
	"fmt"

	"github.com/HUSTSecLab/OpenSift/pkg/storage"
)

type PlatformLinkRepository interface {
	IsLinkInPlatform(link string) (bool, error)
	BeginTemp() error
	BatchInsertTemp(links []string) error
	CommitTemp() error
}

type PlatformLinkTablePrefix string

const (
	PlatformLinkTablePrefixGithub    PlatformLinkTablePrefix = "github"
	PlatformLinkTablePrefixGitlab                            = "gitlab"
	PlatformLinkTablePrefixBitbucket                         = "bitbucket"
	PlatformLinkTablePrefixGitee                             = "gitee"
)

type platformLinkRepository struct {
	AppDb    storage.AppDatabaseContext
	Platform PlatformLinkTablePrefix
}

var _ PlatformLinkRepository = (*platformLinkRepository)(nil)

func NewPlatformLinkRepository(appDb storage.AppDatabaseContext, platform PlatformLinkTablePrefix) PlatformLinkRepository {
	return &platformLinkRepository{
		AppDb:    appDb,
		Platform: platform,
	}
}

func getPlatformTableName(platform PlatformLinkTablePrefix) string {
	return fmt.Sprintf("%s_links", platform)
}

func (r *platformLinkRepository) IsLinkInPlatform(link string) (bool, error) {
	query := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s WHERE link = $1)`, getPlatformTableName(r.Platform))
	row := r.AppDb.QueryRow(query, link)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *platformLinkRepository) BeginTemp() error {
	tn := getPlatformTableName(r.Platform)
	query := fmt.Sprintf(`
		DROP TABLE IF EXISTS %s_tmp;
		CREATE TEMPORARY TABLE %s_tmp AS TABLE %s WITH NO DATA;
	`, tn, tn, tn)
	_, err := r.AppDb.Exec(query)
	return err
}

// BatchInsertTemp implements PlatformLinkRepository.
func (r *platformLinkRepository) BatchInsertTemp(links []string) error {
	if len(links) == 0 {
		return nil
	}
	query := fmt.Sprintf(`INSERT INTO %s_tmp (git_link) VALUES`, getPlatformTableName(r.Platform))
	args := make([]interface{}, 0, len(links))
	for i, link := range links {
		if i == 0 {
			query += fmt.Sprintf(`($%d)`, i+1)
		} else {
			query += fmt.Sprintf(`, ($%d)`, i+1)
		}
		args = append(args, link)
	}
	_, err := r.AppDb.Exec(query, args...)
	return err
}

// CommitTemp implements PlatformLinkRepository.
func (r *platformLinkRepository) CommitTemp() error {
	tn := getPlatformTableName(r.Platform)
	query := fmt.Sprintf(`
		DELETE FROM %s;
		INSERT INTO %s (SELECT DISTINCT * FROM %s_tmp);
		DROP TABLE %s_tmp;
	`, tn, tn, tn, tn)
	_, err := r.AppDb.Exec(query)
	return err

}
