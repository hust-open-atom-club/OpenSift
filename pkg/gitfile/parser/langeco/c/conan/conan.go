package conan

import (
	"encoding/json"
	"strings"

	"github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser/langeco"
	"github.com/liamg/jfather"
	"golang.org/x/xerrors"
)

type LockFile struct {
	GraphLock GraphLock `json:"graph_lock"`
	Requires  Requires  `json:"requires"`
}

type GraphLock struct {
	Nodes map[string]Node `json:"nodes"`
}

type Node struct {
	Ref       string   `json:"ref"`
	Requires  []string `json:"requires"`
	StartLine int
	EndLine   int
}

type Require struct {
	Dependency string
	StartLine  int
	EndLine    int
}

type Requires []Require

// ToDo
func parseV1(lock LockFile) (*langeco.Package, *langeco.Dependencies, error) {
	return &langeco.Package{}, &langeco.Dependencies{}, nil
}

// ToDo
func parseV2(lock LockFile) (*langeco.Package, *langeco.Dependencies, error) {
	return &langeco.Package{}, &langeco.Dependencies{}, nil
}

func ParsePackage(text string) (string, string, error) {
	// full ref format: package/version@user/channel#rrev:package_id#prev
	// various examples:
	// 'pkga/0.1@user/testing'
	// 'pkgb/0.1.0'
	// 'pkgc/system'
	// 'pkgd/0.1.0#7dcb50c43a5a50d984c2e8fa5898bf18'
	ss := strings.Split(strings.Split(strings.Split(text, "@")[0], "#")[0], "/")
	if len(ss) != 2 {
		return "", "", xerrors.Errorf("Unable to determine conan dependency: %q", text)
	}
	return ss[0], ss[1], nil
}

// UnmarshalJSONWithMetadata needed to detect start and end lines of deps
func (n *Node) UnmarshalJSONWithMetadata(node jfather.Node) error {
	if err := node.Decode(&n); err != nil {
		return err
	}
	// Decode func will overwrite line numbers if we save them first
	n.StartLine = node.Range().Start.Line
	n.EndLine = node.Range().End.Line
	return nil
}

func (r *Require) UnmarshalJSONWithMetadata(node jfather.Node) error {
	var dep string
	if err := node.Decode(&dep); err != nil {
		return err
	}
	r.Dependency = dep
	r.StartLine = node.Range().Start.Line
	r.EndLine = node.Range().End.Line
	return nil
}

func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {
	var lock LockFile
	if err := json.Unmarshal([]byte(content), &lock); err != nil {
		return nil, nil, err
	}

	// try to parse requirements as conan v1.x
	if lock.GraphLock.Nodes != nil {
		return parseV1(lock)
	} else {
		// try to parse requirements as conan v2.x
		return parseV2(lock)
	}
}
