package gem

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
)

const specNewStr = "Gem::Specification.new"

var (
	// Capture the variable name
	// e.g. Gem::Specification.new do |s|
	//      => s
	newVarRegexp = regexp.MustCompile(`\|(?P<var>.*)\|`)

	// Capture the value of "name"
	// e.g. s.name = "async".freeze
	//      => "async".freeze
	nameRegexp = regexp.MustCompile(`\.name\s*=\s*(?P<name>\S+)`)

	// Capture the value of "version"
	// e.g. s.version = "1.2.3"
	//      => "1.2.3"
	versionRegexp = regexp.MustCompile(`\.version\s*=\s*(?P<version>\S+)`)
)

func findSubString(re *regexp.Regexp, line, name string) string {
	m := re.FindStringSubmatch(line)
	if m == nil {
		return ""
	}
	return m[re.SubexpIndex(name)]
}

// Trim single quotes, double quotes and ".freeze"
// e.g. "async".freeze => async
func trim(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, ".freeze")
	return strings.Trim(s, `'"`)
}

func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {
	var newVar, name, version, license string

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line := strings.TrimSpace(line)
		if strings.Contains(line, specNewStr) {
			newVar = findSubString(newVarRegexp, line, "var")
		}

		if newVar == "" {
			continue
		}

		// Capture name, version, license, and licenses
		switch {
		case strings.HasPrefix(line, fmt.Sprintf("%s.name", newVar)):
			// https://guides.rubygems.org/specification-reference/#name
			name = findSubString(nameRegexp, line, "name")
			name = trim(name)
		case strings.HasPrefix(line, fmt.Sprintf("%s.version", newVar)):
			// https://guides.rubygems.org/specification-reference/#version
			version = findSubString(versionRegexp, line, "version")
			version = trim(version)
		}

		// No need to iterate the loop anymore
		if name != "" && version != "" && license != "" {
			break
		}
	}

	return &langeco.Package{
		Name:    name,
		Version: version,
		Eco:     parser.GEMS,
	}, nil, nil
}
