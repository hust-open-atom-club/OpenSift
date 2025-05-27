package mod

import (
	"errors"

	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
	"golang.org/x/mod/modfile"
)

var (
	ErrParsingFailed = errors.New("parsing go.mod failed")
)

// * Parse go.mod
func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {
	f, err := modfile.Parse("go.mod", []byte(content), nil)
	if err != nil {
		return nil, nil, ErrParsingFailed
	}

	var pkg langeco.Package
	if f.Module != nil {
		pkg = langeco.Package{
			Name:    f.Module.Mod.Path,
			Version: f.Module.Mod.Version,
			Eco:     parser.GO,
		}
	}

	deps := make(langeco.Dependencies, 0)
	for _, req := range f.Require {
		if !req.Indirect {
			deps = append(deps, langeco.Package{
				Name:    req.Mod.Path,
				Version: req.Mod.Version,
				Eco:     parser.GO,
			})
		}
	}

	return &pkg, &deps, nil
}
