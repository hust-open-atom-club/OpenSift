package poetry

import (
	"errors"

	"github.com/BurntSushi/toml"
	"github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser/langeco"
)

var (
	ErrDecodingFailed = errors.New("decoding poetry lock file failed")
)

type Lockfile struct {
	Packages []struct {
		Category       string                 `toml:"category"`
		Description    string                 `toml:"description"`
		Marker         string                 `toml:"marker,omitempty"`
		Name           string                 `toml:"name"`
		Optional       bool                   `toml:"optional"`
		PythonVersions string                 `toml:"python-versions"`
		Version        string                 `toml:"version"`
		Dependencies   map[string]interface{} `toml:"dependencies"`
		Metadata       interface{}
	} `toml:"package"`
}

func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {
	var lockfile Lockfile
	if _, err := toml.Decode(content, &lockfile); err != nil {
		return nil, nil, ErrDecodingFailed
	}

	// ToDo
	return nil, nil, nil
}
