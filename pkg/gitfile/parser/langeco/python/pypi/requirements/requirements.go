package requirements

import (
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
)

func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {
	// `requirements.txt` can use byte order marks (BOM)
	// e.g. on Windows `requirements.txt` can use UTF-16LE with BOM
	// We need to override them to avoid the file being read incorrectly
	// ToDo
	return nil, nil, nil
}
