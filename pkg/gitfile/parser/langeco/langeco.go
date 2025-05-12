package langeco

import (
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser"
)

// ToDo Merge results from different files instead of choose a trusted one as final

var (
	SUPPORTED_ECOS = map[string]bool{
		parser.NPM:   true,
		parser.GO:    true,
		parser.MAVEN: true,
		parser.CARGO: true,
		parser.PYPI:  true,
	}

	TRUSTED_FILES = map[string]string{
		parser.NPM:   NPM_PACKAGE_LOCK,
		parser.GO:    GO_MOD,
		parser.MAVEN: MAVEN_POM,
		parser.CARGO: CARGO_TOML,
		parser.PYPI:  PY_PROJECT,
	}
)

const (
	NPM_PACKAGE_LOCK    = "package-lock.json"
	NODEJS_PACKAGE_JSON = "package.json"
	GO_MOD              = "go.mod"
	GO_SUM              = "go.sum"
	MAVEN_POM           = "pom.xml"
	CARGO_TOML          = "Cargo.toml"
	CARGO_LOCK          = "Cargo.lock"
	NUGET               = ""
	PY_PROJECT          = "pyproject.toml"
	PY_REQUIREMENTS     = "requirements.txt"
)

type Package struct {
	Name    string
	Version string
	Eco     string
}

type Dependencies []Package

/*
func ExactUniqueDependencies(ecoDeps *map[*Package]*Dependencies) {
	eco := map[string]bool{}
	pkg := map[string]bool{}
	version := map[string]bool{}
	for k, v := range *ecoDeps {
		_, eok := eco[k.Eco]
		_, pok := pkg[k.Name]
		_, vok := version[k.Version]
		if !eok {
			eco[k.Eco] = true
		}
		if !pok {
			eco[k.Name] = true
		}
		if !vok {
			version[k.Version] = true
		}

	}
}
*/
