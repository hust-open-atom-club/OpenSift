package git

import (
	"fmt"
	"path/filepath"

	parser "github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
	dotnet "github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco/dornet"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco/dornet/nuget"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco/go/mod"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco/go/sum"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco/java/maven"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco/nodejs/npm"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco/nodejs/packagejson"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco/python/pypi/pyproject"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco/python/pypi/requirements"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco/python/pypi/setup"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco/rust/cargo"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco/rust/lock"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type LangEcoConfig struct {
	defaultName    string
	defaultVersion string
	eco            map[string]bool
}

type LangEcoDeps struct {
	languages    map[string]int64
	ecosystems   map[string]int64
	dependencies map[*langeco.Package]*langeco.Dependencies
	config       LangEcoConfig
}

func NewLangEcoDeps(r *Repo) LangEcoDeps {
	return LangEcoDeps{
		languages:    map[string]int64{},
		ecosystems:   map[string]int64{},
		dependencies: map[*langeco.Package]*langeco.Dependencies{},
		config: LangEcoConfig{
			defaultName:    fmt.Sprintf("%s/%s/%s", r.Source, r.Owner, r.Name),
			defaultVersion: " ",
			eco:            langeco.SUPPORTED_ECOS,
		},
	}
}

func (led *LangEcoDeps) Parse(f *object.File) error {
	filename := filepath.Base(f.Name)
	filesize := f.Size

	//* Get language
	if v, ok := parser.LANGUAGE_FILENAMES[filename]; ok {
		led.languages[v] += filesize
	} else {
		ex := filepath.Ext(filename)
		v, ok = parser.LANGUAGE_EXTENSIONS[ex]
		if ok {
			led.languages[v] += filesize
		}
	}

	//* Get Ecosystem and Dependency
	if v, ok := parser.ECOSYSTEM_MAP[filename]; ok {
		led.ecosystems[v] += filesize
		if t, ok := led.config.eco[v]; ok && t {
			led.getDependencies(f)
		}
	}

	return nil
}

func (led *LangEcoDeps) getDependencies(file *object.File) {
	filename := filepath.Base(file.Name)

	eco, ok := parser.ECOSYSTEM_MAP[filename]
	if ok {
		if v, ok := langeco.SUPPORTED_ECOS[eco]; !ok || !v {
			return
		}
	} else {
		return
	}

	content, err := file.Contents()
	if err != nil {
		logger.Error(err)
		return
	}

	pkg := &langeco.Package{}
	deps := &langeco.Dependencies{}

	switch filename {
	case langeco.PY_SETUP:
		pkg, deps, err = setup.Parse(content)
	case langeco.NODEJS_PACKAGE_JSON:
		pkg, deps, err = packagejson.Parse(content)
	case langeco.GO_MOD:
		pkg, deps, err = mod.Parse(content)
	case langeco.GO_SUM:
		pkg, deps, err = sum.Parse(content)
	case langeco.NPM_PACKAGE_LOCK:
		pkg, deps, err = npm.Parse(content)
	case langeco.CARGO_TOML:
		pkg, deps, err = cargo.Parse(content)
	case langeco.CARGO_LOCK:
		pkg, deps, err = lock.Parse(content)
	case langeco.PY_PROJECT:
		pkg, deps, err = pyproject.Parse(content)
	case langeco.MAVEN_POM:
		pkg, deps, err = maven.Parse(content)
	case langeco.PY_REQUIREMENTS:
		pkg, deps, err = requirements.Parse(content)
	case langeco.DOT_NET:
		pkg, deps, err = dotnet.Parse(content)
	case langeco.NUGET_CONFIG:
		pkg, deps, err = nuget.Parse(content)
	default:
		return
	}

	if err != nil {
		logger.Error(err)
		return
	}

	if pkg != nil {
		if pkg.Name == "" {
			pkg.Name = led.config.defaultName
		}
		if pkg.Version == "" {
			pkg.Version = led.config.defaultVersion
		}
		led.dependencies[pkg] = deps
		/*
			if v, ok := langeco.TRUSTED_FILES[eco]; ok && filename == v {
				led.config.eco[eco] = false
			}
		*/
	}
}

func (led *LangEcoDeps) Merge(r *Repo) {
	depsMap := make(map[langeco.Package]langeco.Dependencies)
	for pkg, deps := range r.EcoDeps {
		if existingDeps, exists := depsMap[*pkg]; exists {
			depsMap[*pkg] = led.mergeDependencyLists(existingDeps, *deps)
		} else {
			depsMap[*pkg] = *deps
		}
	}

	result := make(map[*langeco.Package]*langeco.Dependencies)

	for pkgVal, deps := range depsMap {
		result[&pkgVal] = &deps
	}

	r.EcoDeps = result

}

func (led *LangEcoDeps) mergeDependencyLists(a, b langeco.Dependencies) langeco.Dependencies {
	seen := make(map[langeco.Package]bool)
	merged := make(langeco.Dependencies, 0)

	for _, dep := range a {
		if !seen[dep] {
			seen[dep] = true
			merged = append(merged, dep)
		}
	}

	for _, dep := range b {
		if !seen[dep] {
			seen[dep] = true
			merged = append(merged, dep)
		}
	}

	return merged
}
