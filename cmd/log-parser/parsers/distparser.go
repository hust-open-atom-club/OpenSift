package parsers

import "fmt"

var ErrInvalidURL = fmt.Errorf("invalid url")
var ErrInvalidPackageName = fmt.Errorf("invalid package name")

type MatchInfo struct {
	Url string
	Ua  string
}
type Parser interface {
	IsMatch(info MatchInfo) bool
	ParseLine(url string) error
	Tag() string
	GetResult() map[string]int
}
