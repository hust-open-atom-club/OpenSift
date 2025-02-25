package parsers

import (
	"strings"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
)

type ArchLinuxParser struct {
	db           *storage.AppDatabaseContext
	packageCount map[string]int
}

func (a *ArchLinuxParser) IsMatch(info MatchInfo) bool {
	if !strings.HasPrefix(info.Url, "/archlinux/") {
		return false
	}
	if !strings.HasSuffix(info.Url, ".pkg.tar.zst") {
		return false
	}
	return true
}

func (a *ArchLinuxParser) ParseLine(url string) error {
	urlSegs := strings.Split(url, "/")
	if len(urlSegs) != 6 {
		return ErrInvalidURL
	}
	fileName := urlSegs[5]
	// btrfs-progs-6.13-1-x86_64.pkg.tar.zst
	packageNameSegs := strings.Split(fileName, "-")
	if len(packageNameSegs) < 3 {
		return ErrInvalidPackageName
	}
	packageName := packageNameSegs[0]
	for i := 1; i < len(packageNameSegs)-2; i++ {
		maybeNamePart := packageNameSegs[i]
		// if maybenamepart starts with digit, it is version
		if maybeNamePart[0] >= '0' && maybeNamePart[0] <= '9' {
			break
		}
		packageName += "-" + maybeNamePart
	}
	if _, ok := a.packageCount[packageName]; ok {
		a.packageCount[packageName]++
	} else {
		a.packageCount[packageName] = 1
	}
	return nil
}
func (a *ArchLinuxParser) Tag() string {
	return "archlinux"
}
func (a *ArchLinuxParser) GetResult() map[string]int {
	return a.packageCount
}
func NewArchLinuxParser(db *storage.AppDatabaseContext) *ArchLinuxParser {
	return &ArchLinuxParser{
		packageCount: make(map[string]int),
		db:           db,
	}
}
