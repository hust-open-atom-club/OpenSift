package cargo

import (
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco/rust/lock"
)

func Parse(contents string) (*langeco.Package, *langeco.Dependencies, error) {
	return lock.Parse(contents)
}
