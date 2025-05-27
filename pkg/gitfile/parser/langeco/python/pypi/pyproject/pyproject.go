package pyproject

import (
	"errors"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
)

var (
	ErrDecodingFailed = errors.New("decoding pip toml failed")
)

type Project struct {
	Name    string `toml:"name"`
	Version string `toml:"version"`
	//*	Description string `toml:"description"`
	//*	License              License              `toml:"license"`
	//*	Maintainers          Maintainers          `toml:"maintainers"`
	Dependencies         Dependencies         `toml:"dependencies"`
	Url                  URL                  `toml:"urls"`
	OptionalDependencies OptionalDependencies `toml:"optional-dependencies"`
}

// * type License map[string]string
// * type Maintainers []map[string]string
type Dependencies []string
type URL map[string]string
type OptionalDependencies map[string]interface{}

type PyProject struct {
	Project Project `toml:"project"`
	Tool    Tool    `toml:"tool"`
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

// * Parse pyproject.toml defined in PEP518.
// https://peps.python.org/pep-0518/
func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {
	var proj PyProject
	if _, err := toml.Decode(content, &proj); err != nil {
		logger.Error(err)
		return nil, nil, ErrDecodingFailed
	}

	name := proj.Tool.Poetry.Name
	version := proj.Tool.Poetry.Version

	if proj.Project.Name != "" {
		name = proj.Project.Name
	}

	if proj.Project.Version != "" {
		version = proj.Project.Version
	}

	pkg := langeco.Package{
		Name:    name,
		Version: version,
		Eco:     parser.PYPI,
	}

	deps := make(langeco.Dependencies, 0)

	for _, dep := range proj.Project.Dependencies {
		deps = append(deps, parseDeps(dep))
	}
	for _, dep := range proj.Project.OptionalDependencies {
		switch value := dep.(type) {
		case string:
			deps = append(deps, parseDeps(value))
		case []string:
			deps = append(deps, parseDeps(value[0]))
		default:
			logger.Info(value)
		}
	}
	for name, element := range proj.Tool.Poetry.Dependencies {
		var version string
		switch value := element.(type) {
		case string:
			version = value
		//* case []string:
		case map[string]string:
			version = value["version"]
		case []map[string]string:
			version = value[0]["version"]
		default:
			continue
		}
		deps = append(deps, langeco.Package{
			Name:    name,
			Version: version,
			Eco:     parser.PYPI,
		})
	}

	return &pkg, &deps, nil
}

func parseDeps(dep string) langeco.Package {
	s := strings.Split(dep, ";")[0]
	var name, version string
	if strings.Contains(s, ">=") {
		d := strings.Split(s, ">=")
		name = d[0]
		version = ">=" + d[1]
	} else if strings.Contains(s, "<=") {
		d := strings.Split(s, "<=")
		name = d[0]
		version = "<=" + d[1]
	} else if strings.Contains(s, "=") {
		d := strings.Split(s, "=")
		name = d[0]
		version = d[1]
	} else {
		name = s
	}
	return langeco.Package{
		Name:    name,
		Version: version,
		Eco:     parser.PYPI,
	}
}
