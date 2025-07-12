package lock

import (
	"errors"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser"
	"github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser/langeco"
	"github.com/HUSTSecLab/OpenSift/pkg/logger"
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
		if pkg.Source == "" && pkg.Dependencies != nil {
			if rootPkg != nil {
				logger.Debug(ErrMultiRootPkg)
				//* return nil, nil, ErrMultiRootPkg
				rootPkg.Dependencies = append(rootPkg.Dependencies, pkg.Dependencies...)
				if len(rootPkg.Name) > len(pkg.Name) {
					rootPkg.Name = pkg.Name
					rootPkg.Version = pkg.Version
				}
			} else {
				rootPkg = &pkg
			}
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
						Eco:     parser.CARGO,
					})
					break
				}
			}
			continue
		}
		deps = append(deps, langeco.Package{
			Name:    parts[0],
			Version: parts[1],
			Eco:     parser.CARGO,
		})
	}

	pkg := langeco.Package{
		Name:    rootPkg.Name,
		Version: rootPkg.Version,
		Eco:     parser.CARGO,
	}

	return &pkg, &deps, nil
}
