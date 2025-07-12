package setup

import (
	"strings"

	"github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser"
	"github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser/langeco"
)

func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {
	lines := strings.Split(content, "\n")
	var name, version string
	flag := false
	for _, line := range lines {
		if strings.Contains(line, "setup(") {
			flag = true
			line = strings.TrimLeft(line, "setup(")
		}
		if flag {
			if strings.Contains(line, "name=\"") {
				name = strings.ReplaceAll(line, "\"", "")
				name = strings.ReplaceAll(name, "(", "")
				name = strings.ReplaceAll(name, ")", "")
				name = strings.ReplaceAll(name, " ", "")
				name = strings.ReplaceAll(name, ",", "")
				name = strings.ReplaceAll(name, "\r", "")
				name = strings.ReplaceAll(name, "\n", "")
			}
			if strings.Contains(line, "version=\"") {
				version = strings.ReplaceAll(line, "\"", "")
				version = strings.ReplaceAll(version, "(", "")
				version = strings.ReplaceAll(version, ")", "")
				version = strings.ReplaceAll(version, " ", "")
				version = strings.ReplaceAll(version, ",", "")
				version = strings.ReplaceAll(version, "\r", "")
				version = strings.ReplaceAll(version, "\n", "")
			}
			if name != "" && version != "" {
				break
			}
		}
	}
	return &langeco.Package{
		Name:    name,
		Version: version,
		Eco:     parser.PYPI,
	}, nil, nil
}
