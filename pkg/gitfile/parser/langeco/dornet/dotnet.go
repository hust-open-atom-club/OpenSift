package dotnet

import (
	"encoding/json"
	"strings"

	"github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser"
	"github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser/langeco"
	"github.com/samber/lo"
	"golang.org/x/xerrors"
)

type dotNetDependencies struct {
	Libraries     map[string]dotNetLibrary        `json:"libraries"`
	RuntimeTarget RuntimeTarget                   `json:"runtimeTarget"`
	Targets       map[string]map[string]TargetLib `json:"targets"`
}

type dotNetLibrary struct {
	Type string `json:"type"`
}

type RuntimeTarget struct {
	Name string `json:"name"`
}

type TargetLib struct {
	Runtime        any `json:"runtime"`
	RuntimeTargets any `json:"runtimeTargets"`
	Native         any `json:"native"`
}

// isRuntimeLibrary returns true if library contains `runtime`, `runtimeTarget` or `native` sections, or if the library is missing from `targetLibs`.
// See https://github.com/aquasecurity/trivy/discussions/4282#discussioncomment-8830365 for more details.
func isRuntimeLibrary(targetLibs map[string]TargetLib, library string) bool {
	lib, ok := targetLibs[library]
	// Selected target doesn't contain library
	// Mark these libraries as runtime to avoid mistaken omission
	if !ok {
		return true
	}
	// Check that `runtime`, `runtimeTarget` and `native` sections are not empty
	return !lo.IsEmpty(lib)
}

func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {
	var depsFile dotNetDependencies

	if err := json.Unmarshal([]byte(content), &depsFile); err != nil {
		return nil, nil, xerrors.Errorf("failed to decode .deps.json file: %w", err)
	}

	var deps langeco.Dependencies
	for nameVer, lib := range depsFile.Libraries {
		if !strings.EqualFold(lib.Type, "package") {
			continue
		}

		split := strings.Split(nameVer, "/")
		if len(split) != 2 {
			// Invalid name
			continue
		}

		if targetLibs, ok := depsFile.Targets[depsFile.RuntimeTarget.Name]; !ok {
		} else if !isRuntimeLibrary(targetLibs, nameVer) {
			continue
		}

		deps = append(deps, langeco.Package{
			Name:    split[0],
			Version: split[1],
			Eco:     parser.DOTNET,
		})
	}

	return &langeco.Package{
		Eco: parser.DOTNET,
	}, &deps, nil
}
