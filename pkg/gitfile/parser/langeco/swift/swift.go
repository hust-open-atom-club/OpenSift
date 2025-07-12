package swift

import "github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser/langeco"

func Parse(contents string) (*langeco.Package, *langeco.Dependencies, error) {
	return &langeco.Package{}, &langeco.Dependencies{}, nil
}
