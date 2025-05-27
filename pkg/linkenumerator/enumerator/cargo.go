package enumerator

import (
	"fmt"
	"sync"
	"time"

	"github.com/HUSTSecLab/criticality_score/pkg/linkenumerator/api"
	"github.com/HUSTSecLab/criticality_score/pkg/linkenumerator/api/cargo"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/bytedance/gopkg/util/gopool"
)

// Todo Use channel to receive and write data
func (c *enumeratorBase) EnumerateCargo() {
	api_url := api.CRATES_IO_ENUMERATE_API_URL
	var wg sync.WaitGroup
	ch := make(chan []cargo.Crate)
	crates := []cargo.Crate{}
	// ToDo Set Wait Group
	wg.Add(1)
	gopool.Go(func() {
		defer wg.Done()
		for crate := range ch {
			crates = append(crates, crate...)
		}
	})
	for page := 1; page <= 1; page++ {
		time.Sleep(api.TIME_INTERVAL * time.Second)
		gopool.Go(func() {
			defer wg.Done()
			u := fmt.Sprintf(
				"%s?%s=%s&%s=%d&%s=%d",
				api_url,
				"sort", "downloads",
				"per_page", api.PER_PAGE,
				"page", page,
			)
			res, err := c.fetch(u)
			if err != nil {
				logger.Panic("Cargo", err)
			}
			resp := cargo.Response{}
			if err = res.UnmarshalJson(&resp); err != nil {
				logger.Panic("Cargo", err)
			}
			ch <- resp.Crates
		})
	}
	wg.Wait()
}
