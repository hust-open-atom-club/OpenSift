package pyproject

import (
	"errors"

	"github.com/BurntSushi/toml"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
)

var (
	ErrDecodingFailed = errors.New("decoding pip toml failed")
)

type PyProject struct {
	Tool Tool `toml:"tool"`
}

type Tool struct {
	Poetry Poetry `toml:"poetry"`
}

type Poetry struct {
	Name            string                 `toml:"name"`
	Version         string                 `toml:"version"`
	DevDependencies map[string]interface{} `toml:"dev-dependencies"`
	Dependencies    map[string]interface{} `toml:"dependencies"`
}

// Parser parses pyproject.toml defined in PEP518.
// https://peps.python.org/pep-0518/
func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {
	var conf PyProject
	if _, err := toml.Decode(content, &conf); err != nil {
		return nil, nil, ErrDecodingFailed
	}

	pkg := langeco.Package{
		Name:    conf.Tool.Poetry.Name,
		Version: conf.Tool.Poetry.Version,
	}

	deps := make(langeco.Dependencies, 0)

	for name, element := range conf.Tool.Poetry.Dependencies {
		var version string
		switch value := element.(type) {
		case string:
			version = value
		case map[string]string:
			version = value["version"]
		case []map[string]string:
			version = value[0]["version"]
		}
		deps = append(deps, langeco.Package{
			Name:    name,
			Version: version,
		})
	}

	return &pkg, &deps, nil
}
