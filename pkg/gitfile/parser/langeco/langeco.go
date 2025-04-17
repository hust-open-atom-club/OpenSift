package langeco

var (
	SUPPORTED_ECOS map[string]bool = map[string]bool{
		NPM:    true,
		GO_MOD: true,
		GO_SUM: true,
		MAVEN:  true,
		CARGO:  true,
		PYPI:   true,
	}
)

const (
	NPM    = "package-lock.json"
	GO_MOD = "go.mod"
	GO_SUM = "go.sum"
	MAVEN  = "pom.xml"
	CARGO  = "Cargo.lock"
	NUGET  = ""
	PYPI   = "pyproject.toml"
)

type Package struct {
	Name    string
	Version string
	Eco     string
}

type Dependencies []Package
