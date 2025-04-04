package conda

import (
	"encoding/json"
	"errors"

	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
)

var (
	ErrDecodingFailed = errors.New("decoding conda json file failed")
	ErrParsingFailed  = errors.New("parsing conda json file failed")
)

type packageJSON struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	License string `json:"license"`
}

// Parse parses Anaconda (a.k.a. conda) environment metadata.
// e.g. <conda-root>/envs/<env>/conda-meta/<package>.json
// For details see https://conda.io/projects/conda/en/latest/user-guide/concepts/environments.html
func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {
	var data packageJSON
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return nil, nil, ErrDecodingFailed
	}

	if data.Name == "" || data.Version == "" {
		return nil, nil, ErrParsingFailed
	}

	return &langeco.Package{
		Name:    data.Name,
		Version: data.Version,
	}, nil, nil
}
