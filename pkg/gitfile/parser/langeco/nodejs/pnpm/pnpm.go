package pnpm

import (
	"errors"

	"github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser/langeco"
	"gopkg.in/yaml.v3"
)

var (
	ErrDecodingFailed = errors.New("decoding npm lockfile failed")
	ErrInvalidVersion = errors.New("")
)

type PackageResolution struct {
	Tarball string `yaml:"tarball,omitempty"`
}

type PackageInfo struct {
	Resolution      PackageResolution `yaml:"resolution"`
	Dependencies    map[string]string `yaml:"dependencies,omitempty"`
	DevDependencies map[string]string `yaml:"devDependencies,omitempty"`
	IsDev           bool              `yaml:"dev,omitempty"`
	Name            string            `yaml:"name,omitempty"`
	Version         string            `yaml:"version,omitempty"`
}

type LockFile struct {
	LockfileVersion any                    `yaml:"lockfileVersion"`
	Dependencies    map[string]any         `yaml:"dependencies,omitempty"`
	DevDependencies map[string]any         `yaml:"devDependencies,omitempty"`
	Packages        map[string]PackageInfo `yaml:"packages,omitempty"`
}

func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {
	var lockFile LockFile
	if err := yaml.Unmarshal([]byte(content), &lockFile); err != nil {
		return nil, nil, ErrDecodingFailed
	}

	// ToDo
	return &langeco.Package{}, &langeco.Dependencies{}, nil
}
