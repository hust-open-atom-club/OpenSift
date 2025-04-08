package cargo

import (
	"errors"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
)

var (
	ErrDecodingFailed  = errors.New("decoding cargo lock file failed")
	ErrMultiRootPkg    = errors.New("multiple root cargo package found")
	ErrRootPkgNotFound = errors.New("root cargo package not found")
)

type cargoPkg struct {
	Name         string   `toml:"name"`
	Version      string   `toml:"version"`
	Source       string   `toml:"source,omitempty"`
	Dependencies []string `toml:"dependencies,omitempty"`
}

type lockFile struct {
	Packages []cargoPkg `toml:"package"`
}

// * Parse Cargo.lock File
func Parse(contents string) (*langeco.Package, *langeco.Dependencies, error) {
	var lockfile lockFile

	if _, err := toml.Decode(contents, &lockfile); err != nil {
		return nil, nil, ErrDecodingFailed
	}

	var rootPkg *cargoPkg

	for _, pkg := range lockfile.Packages {
		if pkg.Source == "" {
			if rootPkg != nil {
				return nil, nil, ErrMultiRootPkg
			}
			rootPkg = &pkg
		}
	}

	if rootPkg == nil {
		return nil, nil, ErrRootPkgNotFound
	}

	deps := make(langeco.Dependencies, 0)
	for _, depStr := range rootPkg.Dependencies {
		parts := strings.Fields(depStr)
		if len(parts) < 2 {
			for _, depPkg := range lockfile.Packages {
				if depPkg.Name == parts[0] {
					deps = append(deps, langeco.Package{
						Name:    depPkg.Name,
						Version: depPkg.Version,
					})
					break
				}
			}
			continue
		}
		deps = append(deps, langeco.Package{
			Name:    parts[0],
			Version: parts[1],
		})
	}

	pkg := langeco.Package{
		Name:    rootPkg.Name,
		Version: rootPkg.Version,
	}

	return &pkg, &deps, nil
}

// ToDo Parsing Indirect dependencies is available but not necessary now
