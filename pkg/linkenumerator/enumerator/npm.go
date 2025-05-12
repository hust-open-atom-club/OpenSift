package enumerator

import (
	"encoding/json"

	"github.com/HUSTSecLab/criticality_score/pkg/logger"
)

type npmEnumerator struct {
	enumeratorBase
}

// Enumerate implements Enumerator.
func (n *npmEnumerator) Enumerate() error {
	req := n.client.R()
	req.SetURL("https://github.com/nice-registry/all-the-package-repos/raw/refs/heads/master/data/packages.json")

	logger.Info("Downloading npm data...")
	resp := req.Do()
	if resp.IsErrorState() {
		return resp.Err
	}

	logger.Info("Parsing npm data...")

	var data map[string]*string
	err := json.Unmarshal(resp.Bytes(), &data)

	if err != nil {
		logger.Error("Failed to parse npm data: ", err)
		return err
	}

	// remove duplicates
	distinctDataMap := make(map[string]struct{})
	for _, v := range data {
		if v == nil {
			continue
		}
		distinctDataMap[*v] = struct{}{}
	}
	// convert map to list
	var distinctData []string
	for k := range distinctDataMap {
		distinctData = append(distinctData, k)
	}

	n.writer.Open()
	defer n.writer.Close()
	for _, v := range distinctData {
		n.writer.Write(v)
	}
	return nil
}

var _ Enumerator = &npmEnumerator{}

func NewNpmEnumerator() Enumerator {
	return &npmEnumerator{
		enumeratorBase: newEnumeratorBase(),
	}
}
