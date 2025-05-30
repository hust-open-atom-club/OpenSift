package langeco

import (
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser"
)

var (
	SUPPORTED_ECOS = map[string]bool{
		parser.NPM:   true,
		parser.GO:    true,
		parser.MAVEN: true,
		parser.CARGO: true,
		parser.PYPI:  true,
		parser.NUGET: true,
	}

/*
	TRUSTED_FILES = map[string]string{
		parser.NPM:   NPM_PACKAGE_LOCK,
		parser.GO:    GO_MOD,
		parser.MAVEN: MAVEN_POM,
		parser.CARGO: CARGO_TOML,
		parser.PYPI:  PY_PROJECT,
		parser.NUGET: NUGET_CONFIG, //* NUGET_PROPS
	}
*/
)

const (
	NUGET_CONFIG        = "packages.config"
	NUGET_PROPS         = "packages.props"
	NPM_PACKAGE_LOCK    = "package-lock.json"
	NODEJS_PACKAGE_JSON = "package.json"
	GO_MOD              = "go.mod"
	GO_SUM              = "go.sum"
	MAVEN_POM           = "pom.xml"
	CARGO_TOML          = "Cargo.toml"
	CARGO_LOCK          = "Cargo.lock"
	DOT_NET             = "deps.json"
	PY_PROJECT          = "pyproject.toml"
	PY_REQUIREMENTS     = "requirements.txt"
	PY_SETUP            = "setup.py"
)

type Package struct {
	Name    string
	Version string
	Eco     string
}

type Dependencies []Package
