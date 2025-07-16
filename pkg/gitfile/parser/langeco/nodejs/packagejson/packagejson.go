package packagejson

import (
	"encoding/json"
	"errors"

	"github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser"
	"github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser/langeco"
)

var (
	ErrDecodingFailed = errors.New("decoding package.json failed")
	ErrFieldMissed    = errors.New("field of package.json missed")
)

type packageJSON struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	//*	License              interface{}       `json:"license"`
	Dependencies         map[string]string `json:"dependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	//*	Workspaces           []string          `json:"workspaces"`
}

func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {
	var pkgJSON packageJSON
	if err := json.Unmarshal([]byte(content), &pkgJSON); err != nil {
		return nil, nil, ErrDecodingFailed
	}

	// Name and version fields are optional
	// https://docs.npmjs.com/cli/v9/configuring-npm/package-json#name
	if pkgJSON.Name == "" || pkgJSON.Version == "" {
		return nil, nil, ErrFieldMissed
	}

	pkg := langeco.Package{
		Name:    pkgJSON.Name,
		Version: pkgJSON.Version,
		Eco:     parser.NPM,
	}

	deps := make(langeco.Dependencies, 0)

	for name, version := range pkgJSON.Dependencies {
		deps = append(deps, langeco.Package{
			Name:    name,
			Version: version,
			Eco:     parser.NPM,
		})
	}

	for name, version := range pkgJSON.DevDependencies {
		deps = append(deps, langeco.Package{
			Name:    name,
			Version: version,
			Eco:     parser.NPM,
		})
	}

	for name, version := range pkgJSON.OptionalDependencies {
		deps = append(deps, langeco.Package{
			Name:    name,
			Version: version,
			Eco:     parser.NPM,
		})
	}

	return &pkg, &deps, nil
}
