// Cargo enumerator for crates.io packages
package enumerator

import (
	"fmt"
	"strings"
	"sync"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api"
	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api/cargo"
	"github.com/sirupsen/logrus"
)

type CargoEnumerator struct {
	enumeratorBase
	take int
}

func NewCargoEnumerator(take int) *CargoEnumerator {
	return &CargoEnumerator{
		enumeratorBase: newEnumeratorBase(),
		take:           take,
	}
}

// Get the best URL for a crate
func getBestCargoUrl(crate *cargo.Crate) string {
	if crate.Homepage != nil && *crate.Homepage != "" {
		return *crate.Homepage
	}
	if crate.Repository != "" {
		return crate.Repository
	}
	return ""
}

// Main enumerate logic with concurrency and pagination
func (c *CargoEnumerator) Enumerate() error {
	// Open writer and initialize variables
	if err := c.writer.Open(); err != nil {
		logrus.Panic("Open writer", err)
		return err
	}
	defer c.writer.Close()

	u := api.CRATES_IO_ENUMERATE_API_URL + "?sort=downloads&per_page=100&page=1"
	collected := 0
	maxConcurrency := 8
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	stop := false

	// Loop through pages with concurrency
	for !stop {
		sem <- struct{}{}
		wg.Add(1)
		pageURL := u
		go func(url string) {
			defer wg.Done()
			defer func() { <-sem }()
			res, err := c.fetch(url)
			if err != nil {
				logrus.Warnf("Cargo fetch error: %v", err)
				return
			}
			resp, err := api.FromCargo(res)
			if err != nil {
				logrus.Warnf("Cargo unmarshal error: %v", err)
				return
			}

			mu.Lock()
			defer mu.Unlock()
			// Write crate info and check stop condition
			for _, crate := range resp.Crates {
				if collected >= c.take {
					stop = true
					break
				}
				url := strings.TrimSuffix(getBestCargoUrl(&crate), ".git")
				c.writer.Write(crate.Name)
				c.writer.Write(url)
				c.writer.Write(crate.MaxVersion)
				c.writer.Write(fmt.Sprintf("%d", crate.Downloads))
				c.writer.Write(fmt.Sprintf("%d", crate.RecentDownloads))
				c.writer.Write("\n")
				collected++
			}
			if collected >= c.take || resp.Meta.NextPage == "" || len(resp.Crates) == 0 {
				stop = true
			} else {
				u = api.CRATES_IO_ENUMERATE_API_URL + resp.Meta.NextPage
			}
		}(pageURL)

		if len(sem) == maxConcurrency {
			wg.Wait()
		}
		if stop {
			break
		}
	}
	wg.Wait()
	// Log final result
	logrus.Infof("Enumerator has collected and written %d repositories", collected)
	return nil
}
