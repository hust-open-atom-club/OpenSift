package cargo

import (
	"errors"

	"github.com/BurntSushi/toml"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
)

var (
	ErrDecodingFailed  = errors.New("decoding cargo lock file failed")
	ErrMultiRootPkg    = errors.New("multiple root cargo package found")
	ErrRootPkgNotFound = errors.New("root cargo package not found")
)

type cargoPkg struct {
	Name          string      `toml:"name"`
	Version       string      `toml:"version"`
	Authors       interface{} `toml:"authors,omitempty"`
	Edition       interface{} `toml:"edition,omitempty"`
	Documentation interface{} `toml:"documentation,omitempty"`
	Homepage      interface{} `toml:"homepage,omitempty"`
	Repository    interface{} `toml:"repository,omitempty"`
	License       interface{} `toml:"license,omitempty"`
	LicenseFile   interface{} `toml:"license-file,omitempty"`
}

type cargoFile struct {
	Package           cargoPkg               `toml:"package"`
	Dependencies      map[string]interface{} `toml:"dependencies"`
	DevDependencies   map[string]interface{} `toml:"dev-dependencies,omitempty"`
	BuildDependencies map[string]interface{} `toml:"build-dependencies,omitempty"`
}

// * Parse Cargo.toml File
func Parse(contents string) (*langeco.Package, *langeco.Dependencies, error) {
	var cargoFile cargoFile

	if _, err := toml.Decode(contents, &cargoFile); err != nil {
		return nil, nil, ErrDecodingFailed
	}

	pkg := langeco.Package{
		Name:    cargoFile.Package.Name,
		Version: cargoFile.Package.Version,
		Eco:     parser.CARGO,
	}
	deps := make(langeco.Dependencies, 0)
	deps = append(deps, *exactDependencies(cargoFile.Dependencies)...)
	deps = append(deps, *exactDependencies(cargoFile.DevDependencies)...)
	deps = append(deps, *exactDependencies(cargoFile.BuildDependencies)...)
	return &pkg, &deps, nil
}

func exactDependencies(dependencies map[string]interface{}) *langeco.Dependencies {
	deps := make(langeco.Dependencies, 0)
	for name, v := range dependencies {
		var version string
		switch value := v.(type) {
		case string:
			version = value
		case map[string]interface{}:
			if ver, ok := value["version"].(string); ok {
				version = ver
			}
		}
		if version != "" {
			deps = append(deps, langeco.Package{
				Name:    name,
				Version: version,
				Eco:     parser.CARGO,
			})
		}
	}
	return &deps
}
