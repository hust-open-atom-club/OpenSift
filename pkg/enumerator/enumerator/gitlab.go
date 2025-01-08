package enumerator

import (
	"fmt"
	"sync"
	"time"

	"github.com/HUSTSecLab/criticality_score/pkg/enumerator/internal/api"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/bytedance/gopkg/util/gopool"
)

type GitlabEnumerator struct {
	enumeratorBase
	take int
	jobs int
}

func NewGitlabEnumerator(take int, jobs int) *GitlabEnumerator {
	return &GitlabEnumerator{
		enumeratorBase: newEnumeratorBase(),
		take:           take,
		jobs:           jobs,
	}
}

// Todo Use channel to receive and write data
func (c *GitlabEnumerator) Enumerate() error {
	if err := c.writer.Open(); err != nil {
		return err
	}
	defer c.writer.Close()

	api_url := api.GITLAB_ENUMERATE_API_URL
	var wg sync.WaitGroup

	collected := 0
	var muCollected sync.Mutex

	pool := gopool.NewPool("gitlab_enumerator", int32(c.jobs), &gopool.Config{})

	for page := 1; page <= c.take/api.PER_PAGE; page++ {
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
				logger.Errorf("Gitlab fetch failed: %v", err)
				return
			}

			resp, err := api.FromGitlab(res)

			if err != nil {
				logger.Errorf("Gitlab unmarshal failed: %v", err)
				return
			}

			for _, v := range *resp {
				c.writer.Write(v.HTTPURLToRepo)
			}

			func() {
				muCollected.Lock()
				defer muCollected.Unlock()
				collected += len(*resp)
				logger.Infof("Enumerator has collected and written %d repositories", collected)
			}()

		})
	}
	wg.Wait()
	return nil
}
