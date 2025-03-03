package enumerator

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/HUSTSecLab/criticality_score/pkg/linkenumerator/api"
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

// Todo Use channel to receive and write data
func (c *gitlabEnumerator) Enumerate() error {
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
				logrus.Errorf("Gitlab fetch failed: %v", err)
				return
			}

			resp, err := api.FromGitlab(res)

			if err != nil {
				logrus.Errorf("Gitlab unmarshal failed: %v", err)
				return
			}

			for _, v := range *resp {
				// if ends with .git, remove it
				if strings.HasSuffix(v.HTTPURLToRepo, ".git") {
					v.HTTPURLToRepo = v.HTTPURLToRepo[:len(v.HTTPURLToRepo)-4]
				}
				c.writer.Write(v.HTTPURLToRepo)
			}

			func() {
				muCollected.Lock()
				defer muCollected.Unlock()
				collected += len(*resp)
				logrus.Infof("Enumerator has collected and written %d repositories", collected)
			}()

		})
	}
	wg.Wait()
	return nil
}
