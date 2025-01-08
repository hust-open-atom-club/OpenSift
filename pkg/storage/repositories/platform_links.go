package repositories

import (
	"fmt"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
)

type PlatformLinkRepository interface {
	IsLinkInPlatform(link string) (bool, error)
	ClearLinks() error
	BatchInsertLinks(links []string) error
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

func (r *platformLinkRepository) ClearLinks() error {
	query := fmt.Sprintf(`DELETE FROM %s`, getPlatformTableName(r.Platform))
	_, err := r.AppDb.Exec(query)
	return err
}

func (r *platformLinkRepository) BatchInsertLinks(links []string) error {
	if len(links) == 0 {
		return nil
	}

	query := fmt.Sprintf(`INSERT INTO %s (link) VALUES`, getPlatformTableName(r.Platform))
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
