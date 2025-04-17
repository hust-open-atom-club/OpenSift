package yarn

import (
	"errors"

	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
)

var (
	ErrParsingFailed = errors.New("parsing yarn lock file failed")
)

type LockFile struct {
	Dependencies map[string]Dependency
}

type Library struct {
	Patterns []string
	Name     string
	Version  string
}

type Dependency struct {
	Pattern string
	Name    string
}

func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {
	// ToDo
	return &langeco.Package{}, &langeco.Dependencies{}, nil
}
