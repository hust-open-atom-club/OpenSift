package dotnet

import "github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"

func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {
	return &langeco.Package{}, &langeco.Dependencies{}, nil
}
