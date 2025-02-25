package parsers

import (
	"fmt"
	"os"
	"strings"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
)

type DebianParser struct {
	db           *storage.AppDatabaseContext
	packageCount map[string]int
}

func (d *DebianParser) IsMatch(info MatchInfo) bool {
	if !strings.HasPrefix(info.Url, "/debian/pool/") {
		return false
	}
	if !strings.HasSuffix(info.Url, ".deb") {
		return false
	}
	return true
}
func (d *DebianParser) ParseLine(url string) error {
	urlSegs := strings.Split(url, "/")
	if len(urlSegs) < 7 {
		fmt.Fprintf(os.Stderr, "invalid url: %s\n", url)
		return ErrInvalidURL
	}
	// fmt.Println(urlSegs[5], urlSegs[6])
	packageNameSegs := strings.Split(urlSegs[6], "_")
	if len(packageNameSegs) != 3 {
		fmt.Fprintf(os.Stderr, "invalid package name: %s\n", urlSegs[6])
		return ErrInvalidPackageName
	}
	packageName := packageNameSegs[0]
	// packageVersion := packageNameSegs[1]
	if _, ok := d.packageCount[packageName]; ok {
		d.packageCount[packageName]++
	} else {
		d.packageCount[packageName] = 1
	}
	return nil
}

func (d *DebianParser) Tag() string {
	return "debian"
}
func (d *DebianParser) GetResult() map[string]int {
	return d.packageCount
}
func NewDebianParser(db *storage.AppDatabaseContext) *DebianParser {
	return &DebianParser{
		packageCount: make(map[string]int),
		db:           db,
	}
}
