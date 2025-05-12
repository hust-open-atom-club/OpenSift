package requirements

import (
	"strings"
	"unicode"

	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
)

const (
	commentMarker string = "#"
	endColon      string = ";"
	hashMarker    string = "--"
	startExtras   string = "["
	endExtras     string = "]"
)

func splitLine(line string) []string {
	separators := []string{"~=", ">=", "=="}
	// Without useMinVersion check only `==`
	for _, sep := range separators {
		if result := strings.Split(line, sep); len(result) == 2 {
			return result
		}
	}
	return nil
}

func rStripByKey(line, key string) string {
	if pos := strings.Index(line, key); pos >= 0 {
		line = strings.TrimRightFunc((line)[:pos], unicode.IsSpace)
	}
	return line
}

func removeExtras(line string) string {
	startIndex := strings.Index(line, startExtras)
	endIndex := strings.Index(line, endExtras) + 1
	if startIndex != -1 && endIndex != -1 {
		line = line[:startIndex] + line[endIndex:]
	}
	return line
}

func isValidName(name string) bool {
	for _, r := range name {
		// only characters [A-Z0-9._-] are allowed (case insensitive)
		// cf. https://peps.python.org/pep-0508/#names
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '.' && r != '_' && r != '-' {
			return false
		}
	}
	return true
}

func isValidVersion(ver string) bool {
	return true
}

func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {
	// `requirements.txt` can use byte order marks (BOM)
	// e.g. on Windows `requirements.txt` can use UTF-16LE with BOM
	// We need to override them to avoid the file being read incorrectly
	deps := make(langeco.Dependencies, 0)
	for _, line := range strings.Split(content, "\n") {
		line := strings.ReplaceAll(line, " ", "")
		line = strings.ReplaceAll(line, `\`, "")
		line = removeExtras(line)
		line = rStripByKey(line, commentMarker)
		line = rStripByKey(line, endColon)
		line = rStripByKey(line, hashMarker)

		s := splitLine(line)
		if len(s) != 2 {
			continue
		}
		if strings.HasSuffix(s[1], ".*") {
			s[1] = strings.TrimSuffix(s[1], "*") + "0"
		}

		if !isValidName(s[0]) || !isValidVersion(s[1]) {
			logger.Debug(s)
			continue
		}

		deps = append(deps, langeco.Package{
			Name:    s[0],
			Version: s[1],
		})
	}

	return &langeco.Package{
		Eco: parser.PYPI,
	}, &deps, nil
}
