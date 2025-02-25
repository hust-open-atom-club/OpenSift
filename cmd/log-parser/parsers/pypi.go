package parsers

import (
	"fmt"
	"strings"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
)

type PypiParser struct {
	db           *storage.AppDatabaseContext
	packageCount map[string]int
}

func (p *PypiParser) IsMatch(info MatchInfo) bool {
	if !strings.HasPrefix(info.Ua, "pip/") {
		return false
	}
	if !strings.HasPrefix(info.Url, "/pypi/web/packages/") {
		return false
	}
	// fmt.Printf("pypi url: %s\n", info.Url)
	return true
}
func (p *PypiParser) ParseLine(url string) error {
	urlSegs := strings.Split(url, "/")
	if len(urlSegs) != 8 {
		fmt.Printf("invalid pypi url: %v\n", urlSegs)
		return ErrInvalidURL
	}
	packageNameSegs := strings.Split(urlSegs[7], "-")
	if len(packageNameSegs) < 2 {
		fmt.Printf("invalid package name: %s\n", urlSegs[7])
		return ErrInvalidPackageName
	}
	packageName := packageNameSegs[0]
	if _, ok := p.packageCount[packageName]; ok {
		p.packageCount[packageName]++
	} else {
		p.packageCount[packageName] = 1
	}
	return nil
}
func (p *PypiParser) Tag() string {
	return "pypi"
}
func (p *PypiParser) GetResult() map[string]int {
	return p.packageCount
}
func NewPypiParser(db *storage.AppDatabaseContext) *PypiParser {
	return &PypiParser{
		packageCount: make(map[string]int),
		db:           db,
	}
}
