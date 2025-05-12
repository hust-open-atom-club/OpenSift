package sum

import (
	"strings"

	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
)

func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {
	var deps langeco.Dependencies
	uniquePkgs := make(map[string]string)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line := strings.TrimSpace(line)
		s := strings.Fields(line)
		if len(s) < 2 {
			continue
		}
		// go.sum records and sorts all non-major versions
		// with the latest version as last entry
		uniquePkgs[s[0]] = strings.TrimSuffix(s[1], "/go.mod")
	}

	for k, v := range uniquePkgs {
		deps = append(deps, langeco.Package{
			Name:    k,
			Version: v,
		})
	}

	return &langeco.Package{
		Eco: parser.GO,
	}, &deps, nil
}
