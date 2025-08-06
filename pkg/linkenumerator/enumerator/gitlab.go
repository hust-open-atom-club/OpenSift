// GitLab enumerator for popular repositories
package enumerator

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api"
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/sirupsen/logrus"
)

type gitlabEnumerator struct {
	enumeratorBase
	take int
	jobs int
}

func NewGitlabEnumerator(take int, jobs int) Enumerator {
	return &gitlabEnumerator{
		enumeratorBase: newEnumeratorBase(),
		take:           take,
		jobs:           jobs,
	}
}

// Main enumerate logic with concurrency and pagination
func (c *gitlabEnumerator) Enumerate() error {
	// Open writer and initialize variables
	if err := c.writer.Open(); err != nil {
		return err
	}
	defer c.writer.Close()

	api_url := api.GITLAB_ENUMERATE_API_URL
	var wg sync.WaitGroup

	collected := 0
	var muCollected sync.Mutex

	pool := gopool.NewPool("gitlab_enumerator", int32(c.jobs), &gopool.Config{})

	repoCount := 0
	// Loop through pages and fetch repositories concurrently
	for page := 1; repoCount < c.take; page++ {
		time.Sleep(api.TIME_INTERVAL * time.Second)
		wg.Add(1)
		pool.Go(func() {
			defer wg.Done()
			u := fmt.Sprintf(
				"%s?%s=%s&%s=%s&%s=%d&%s=%d",
				api_url,
				"order_by", "star_count",
				"sort", "desc",
				"per_page", api.PER_PAGE,
				"page", page,
			)
			res, err := c.fetch(u)
			if err != nil {
				logrus.Errorf("Gitlab fetch failed: %v", err)
				return
			}

			resp, err := api.FromGitlab(res)

			if err != nil {
				logrus.Errorf("Gitlab unmarshal failed: %v", err)
				return
			}

			// Write repository info
			for _, v := range *resp {
				if repoCount >= c.take {
					break
				}
				if strings.HasSuffix(v.HTTPURLToRepo, ".git") {
					v.HTTPURLToRepo = v.HTTPURLToRepo[:len(v.HTTPURLToRepo)-4]
				}
				c.writer.Write(v.Name)
				c.writer.Write(v.HTTPURLToRepo)
				c.writer.Write(fmt.Sprintf("%d", v.StarCount))
				repoCount++
			}

			func() {
				muCollected.Lock()
				defer muCollected.Unlock()
				collected += len(*resp)
			}()

		})
	}
	wg.Wait()
	// Log final result
	logrus.Infof("Enumerator has collected and written %d repositories", collected)
	return nil
}
