package npm

import (
	"errors"
	"strings"

	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"

	"github.com/liamg/jfather"
)

const nodeModulesDir = "node_modules"

var ErrDecodingFailed = errors.New("decoding npm lockfile failed")

type LockFile struct {
	Name            string                `json:"name"`
	Version         string                `json:"version"`
	Dependencies    map[string]Dependency `json:"dependencies"`
	Packages        map[string]Package    `json:"packages"`
	LockfileVersion int                   `json:"lockfileVersion"`
}

type Dependency struct {
	Version string `json:"version"`
	Dev     bool   `json:"dev"`
	//*	Dependencies map[string]Dependency `json:"dependencies"`
	//* Integrity string `json:"integrity"`
	Requires map[string]string `json:"requires"`
	Resolved string            `json:"resolved"`
}

type Package struct {
	Name                 string            `json:"name"`
	Version              string            `json:"version"`
	Dependencies         map[string]string `json:"dependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	Resolved             string            `json:"resolved"`
	Dev                  bool              `json:"dev"`
	Link                 bool              `json:"link"`
	Workspaces           []string          `json:"workspaces"`
}

func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {
	var lockFile LockFile
	if err := jfather.Unmarshal([]byte(content), &lockFile); err != nil {
		return nil, nil, ErrDecodingFailed
	}

	pkg := langeco.Package{
		Name:    lockFile.Name,
		Version: lockFile.Version,
		Eco:     parser.NPM,
	}

	deps := make(langeco.Dependencies, 0)
	if lockFile.LockfileVersion == 1 {
		for k, v := range lockFile.Dependencies {
			deps = append(deps, langeco.Package{
				Name:    k,
				Version: v.Version,
				Eco:     parser.NPM,
			})
		}
	} else {
		for k, v := range lockFile.Packages {
			var name string
			if v.Name != "" {
				name = v.Name
			} else {
				name = pkgNameFromPath(k)
			}
			deps = append(deps, langeco.Package{
				Name:    name,
				Version: v.Version,
				Eco:     parser.NPM,
			})
		}
	}

	return &pkg, &deps, nil
}

func pkgNameFromPath(path string) string {
	if index := strings.LastIndex(path, nodeModulesDir); index != -1 {
		return path[index+len(nodeModulesDir)+1:]
	}
	logger.Warnf("npm %q package path doesn't have `node_modules` prefix", path)
	return path
}
