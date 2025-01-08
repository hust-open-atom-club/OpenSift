package enumerator

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/HUSTSecLab/criticality_score/internal/logger"
	"github.com/HUSTSecLab/criticality_score/pkg/enumerator/config"
	"github.com/HUSTSecLab/criticality_score/pkg/enumerator/internal/api"
	"github.com/bytedance/gopkg/util/gopool"
)

// Todo Use channel to receive and write data
func (c *enumeratorBase) EnumerateCargo() {
	api_url := api.CRATES_IO_ENUMERATE_API_URL
	var wg sync.WaitGroup
	wg.Add(1)
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
			data := res.Bytes()
			err = os.WriteFile(config.OUTPUT_DIR+config.CRATES_IO_OUTPUT_FILEPATH, data, 0644)
			if err != nil {
				logger.Panic("Cargo", err)
			}
		})
	}
	wg.Wait()
}
