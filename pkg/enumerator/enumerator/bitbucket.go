package enumerator

import (
	"github.com/HUSTSecLab/criticality_score/internal/logger"
	"github.com/HUSTSecLab/criticality_score/pkg/enumerator/internal/api"
	"github.com/HUSTSecLab/criticality_score/pkg/enumerator/internal/api/bitbucket"
)

type BitBucketEnumerator struct {
	enumeratorBase
	take int
}

func NewBitBucketEnumerator(take int) *BitBucketEnumerator {
	return &BitBucketEnumerator{
		enumeratorBase: newEnumeratorBase(),
		take:           take,
	}
}

func getBestBitBucketGitURL(val *bitbucket.Value) string {
	for _, v := range val.Links.Clone {
		if v.Name == "https" || v.Name == "http" {
			return v.Href
		}
	}
	if len(val.Links.Clone) > 0 {
		return val.Links.Clone[0].Href
	}
	return ""
}

func (c *BitBucketEnumerator) Enumerate() error {
	err := c.writer.Open()
	defer c.writer.Close()
	if err != nil {
		logger.Panic("Open writer", err)
	}

	u := api.BITBUCKET_ENUMERATE_API_URL
	collected := 0
	for {
		res, err := c.fetch(u)
		if err != nil {
			logger.Panic("Bitbucket", err)
		}
		resp, err := api.FromBitbucket(res)
		if err != nil {
			logger.Panic("Bitbucket", err)
		}

		for _, v := range resp.Values {
			url := getBestBitBucketGitURL(&v)
			c.writer.Write(url)
		}

		collected += len(resp.Values)

		logger.Infof("Enumerator has collected and written %d repositories", collected)

		if collected >= c.take || resp.Next == "" || len(resp.Values) == 0 {
			break
		}

		u = resp.Next
	}
	return nil
}

// TODO: implement the following functions

// // ToDo
// func (c *Enumerator) enumeratePyPI() {

// }

// // ToDo
// func (c *Enumerator) enumerateNPM() {

// }

// // ToDo
// func (c *Enumerator) enumerateGo() {

// }

// // ToDo
// func (c *Enumerator) enumeratePHP() {

// }

// // ToDo
// func (c *Enumerator) enumerateHaskell() {

// }

// // ToDo
// func (c *Enumerator) enumerateRubyGems() {

// }
