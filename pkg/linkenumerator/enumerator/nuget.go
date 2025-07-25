package enumerator

import (
	"fmt"
	"time"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api"
	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api/nuget"
	"github.com/sirupsen/logrus"
)

type NugetEnumerator struct {
	enumeratorBase
	take int
}

func NewNugetEnumerator(take int) *NugetEnumerator {
	return &NugetEnumerator{
		enumeratorBase: newEnumeratorBase(),
		take:           take,
	}
}

func (c *NugetEnumerator) Enumerate() error {
	err := c.writer.Open()
	defer c.writer.Close()
	if err != nil {
		logrus.Panic("Open writer", err)
	}

	u := api.NUGET_INDEX_URL
	collected := 0
	page := 0

	for {
		time.Sleep(time.Second * api.TIME_INTERVAL)
		page++
		url := fmt.Sprintf("%s?take=%d&skip=%d", u, api.PER_PAGE, (page-1)*api.PER_PAGE)
		res, err := c.fetch(url)
		if err != nil {
			logrus.Panic("NuGet fetch error", err)
		}
		resp := nuget.Response{}
		if err = res.UnmarshalJson(&resp); err != nil {
			logrus.Panic("NuGet unmarshal error", err)
		}

		for _, pkg := range resp.Data {
			c.writer.Write("Name: " + pkg.Title)
			c.writer.Write("Version: " + pkg.Version)
			c.writer.Write("URL: " + pkg.ProjectURL)
			c.writer.Write("TotalDownloads: " + fmt.Sprintf("%d", pkg.TotalDownloads))
			c.writer.Write("\n")
		}

		collected += len(resp.Data)

		if collected >= c.take || len(resp.Data) == 0 {
			break
		}
	}
	logrus.Infof("Enumerator has collected and written %d packages", collected)
	return nil
}
