package enumerator

import (
	"fmt"
	"sync"
	"time"

	"github.com/HUSTSecLab/criticality_score/pkg/linkenumerator/api"
	"github.com/HUSTSecLab/criticality_score/pkg/linkenumerator/api/nuget"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/bytedance/gopkg/util/gopool"
)

// Todo Use channel to receive and write data
func (c *enumeratorBase) EnumerateNuget() {
	api_url := api.NUGET_INDEX_URL
	var wg sync.WaitGroup
	ch := make(chan []nuget.Datum)
	pkgs := []nuget.Datum{}
	// ToDo Set Wait Group
	wg.Add(1)
	gopool.Go(func() {
		defer wg.Done()
		for pkg := range ch {
			pkgs = append(pkgs, pkg...)
		}
	})
	for page := 1; page <= 1; page++ {
		time.Sleep(api.TIME_INTERVAL * time.Second)
		gopool.Go(func() {
			defer wg.Done()
			u := fmt.Sprintf(
				"%s?%s=%d&%s=%d",
				api_url,
				"take", api.PER_PAGE,
				"skip", page*api.PER_PAGE,
			)
			res, err := c.fetch(u)
			if err != nil {
				logger.Panic("NuGet", err)
			}
			resp := nuget.Response{}
			if err = res.UnmarshalJson(&resp); err != nil {
				logger.Panic("NuGet", err)
			}
			ch <- resp.Data
		})
	}
	wg.Wait()
}
