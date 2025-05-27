package pipenv

import (
	"errors"

	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
	"github.com/liamg/jfather"
)

var (
	ErrDecodingFailed = errors.New("decoding pip lock file failed")
)

type lockFile struct {
	Meta    map[string]([]source) `json:"_meta"`
	Default map[string]dependency `json:"default"`
}

type source struct {
	Name string `json:"name"`
}

type dependency struct {
	Version string `json:"version"`
}

func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {
	var lockFile lockFile
	if err := jfather.Unmarshal([]byte(content), &lockFile); err != nil {
		return nil, nil, ErrDecodingFailed
	}

	pkg := langeco.Package{
		Name: lockFile.Meta["sources"][0].Name,
	}
	// ToDo
	return &pkg, nil, nil
}
