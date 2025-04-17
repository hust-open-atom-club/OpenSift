package php

import (
	"errors"

	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
	"github.com/liamg/jfather"
)

var (
	ErrDecodingFailed = errors.New("decoding composer lock file failed")
)

type lockFile struct {
	Packages []packageInfo `json:"packages"`
}

type packageInfo struct {
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Require   map[string]string `json:"require"`
	License   []string          `json:"license"`
	StartLine int
	EndLine   int
}

func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {
	var lockFile lockFile
	if err := jfather.Unmarshal([]byte(content), &lockFile); err != nil {
		return nil, nil, ErrDecodingFailed
	}

	// ToDo
	return &langeco.Package{}, &langeco.Dependencies{}, nil
}

// UnmarshalJSONWithMetadata needed to detect start and end lines of deps
func (t *packageInfo) UnmarshalJSONWithMetadata(node jfather.Node) error {
	if err := node.Decode(&t); err != nil {
		return err
	}
	// Decode func will overwrite line numbers if we save them first
	t.StartLine = node.Range().Start.Line
	t.EndLine = node.Range().End.Line
	return nil
}
