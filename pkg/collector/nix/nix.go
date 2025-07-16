package nix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"unicode"

	collector "github.com/HUSTSecLab/OpenSift/pkg/collector/internal"
	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/HUSTSecLab/OpenSift/pkg/storage/repository"
)

type NixCollector struct {
	collector.CollecterInterface
}

func (nc *NixCollector) Collect(workerCount int, batchSize int, outputPath string) {
	adc := storage.GetDefaultAppDatabaseContext()
	err := nc.ParseInfo(workerCount)
	if err != nil {
		fmt.Printf("Error retrieving Nix packages: %v\n", err)
		return
	}
	nc.GetDep()
	nc.PageRank(0.85, 20)
	nc.GetDepCount()
	nc.UpdateDistRepoCount(adc)
	nc.CalculateDistImpact()
	nc.UpdateOrInsertDatabase(adc)
	nc.UpdateOrInsertDistDependencyDatabase(adc)
	if outputPath != "" {
		err = nc.GenerateDependencyGraph(outputPath)
		if err != nil {
			log.Printf("Error generating dependency graph: %v\n", err)
			return
		}
	}
}

func isValidNixIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	first := s[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}
	for _, c := range s[1:] {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

func attributePathToNixExpression(attributePath string) string {
	components := strings.Split(attributePath, ".")
	expr := "pkgs"
	for _, comp := range components {
		if isValidNixIdentifier(comp) {
			expr += "." + comp
		} else {
			expr += `."` + comp + `"`
		}
	}
	return expr
}

func (nc *NixCollector) ParseInfo(poolsize int) error {
	cmd := exec.Command("nix-env", "-qaP")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Error running nix-env command: %v", err)
	}

	lines := strings.Split(string(out), "\n")

	var mu sync.Mutex
	re := regexp.MustCompile(`^nixpkgs\.(.+?)\s+([^\s]+)$`)
	chunksize := (len(lines) + poolsize - 1) / poolsize
	linechunks := make([][]string, 0, poolsize)
	for i := 0; i < len(lines); i += chunksize {
		end := i + chunksize
		if end > len(lines) {
			end = len(lines)
		}
		linechunks = append(linechunks, lines[i:end])
	}

	wg := WorkerPool(poolsize, func(worker int) {
		if worker > len(linechunks) {
			return
		}
		chunk := linechunks[worker]
		for _, line := range chunk {
			if strings.TrimSpace(line) == "" || strings.Contains(line, "evaluation warning") {
				continue
			}

			matches := re.FindStringSubmatch(line)
			if len(matches) == 3 {
				attributePath := matches[1]
				packageFullName := matches[2]
				parts := strings.Split(packageFullName, "-")
				versionIndex := -1
				for i := 1; i < len(parts); i++ {
					if len(parts[i]) > 0 && unicode.IsDigit(rune(parts[i][0])) {
						versionIndex = i
						break
					}
				}

				var packageName, packageVersion string
				if versionIndex != -1 {
					packageName = strings.Join(parts[:versionIndex], "-")
					packageVersion = strings.Join(parts[versionIndex:], "-")
				} else {
					packageName = packageFullName
					packageVersion = ""
				}

				packageInfo, err := nc.GetNixPackageInfo(attributePath)
				if err != nil {
					fmt.Printf("Error getting info for %s: %v\n", attributePath, err)
					continue
				}

				pkgDepInfo := collector.PackageInfo{
					Name:        packageName,
					Version:     packageVersion,
					Homepage:    packageInfo.Homepage,
					Description: packageInfo.Description,
				}

				dependencies, err := nc.GetNixPackageDependencies(attributePath)
				if err != nil {
					fmt.Printf("Error getting dependencies for %s: %v\n", attributePath, err)
					continue
				}
				pkgDepInfo.DirectDepends = dependencies
				mu.Lock()
				nc.SetPkgInfo(packageName, &pkgDepInfo)
				mu.Unlock()
			}
		}
	})

	wg()
	return nil
}

func (nc *NixCollector) GetNixPackageInfo(attributePath string) (collector.PackageInfo, error) {
	nixPkgExpression := attributePathToNixExpression(attributePath)

	expr := fmt.Sprintf(`
let
  pkgs = import <nixpkgs> {};
  pkg = %s;
  pname = if pkg ? pname then pkg.pname else if pkg ? name then pkg.name else "";
  version = if pkg ? version then pkg.version else "unknown";
  meta = if pkg ? meta then pkg.meta else {};
  homepage = if meta ? homepage then meta.homepage else "";
  description = if meta ? description then meta.description else "";
  srcUrl = if pkg ? src then
    if pkg.src ? url then pkg.src.url else if pkg.src ? urls then builtins.elemAt pkg.src.urls 0 else ""
  else "";
  passthruUrl = if pkg ? passthru && pkg.passthru ? updateScript && pkg.passthru.updateScript ? url then
    pkg.passthru.updateScript.url
  else "";
in
{
  name = pname;
  version = version;
  homepage = homepage;
  description = description;
}
`, nixPkgExpression)
	data, err := nc.nixEval(expr)
	if err != nil {
		return collector.PackageInfo{}, fmt.Errorf("Error running nix-eval for package '%s': %v", attributePath, err)
	}
	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		return collector.PackageInfo{}, fmt.Errorf("Error parsing JSON for package '%s': %v", attributePath, err)
	}

	packageInfo := collector.PackageInfo{
		Name:        result["name"],
		Version:     result["version"],
		Homepage:    result["homepage"],
		Description: result["description"],
	}

	return packageInfo, nil
}

func (nc *NixCollector) GetNixPackageDependencies(attributePath string) ([]string, error) {
	nixPkgExpression := attributePathToNixExpression(attributePath)

	exprTemplate := `
	let
	  pkgs = import <nixpkgs> {};
	  pkg = %s;
	in {
		buildInputs = map (x: if x ? pname then x.pname else if x ? name then x.name else "") (pkg.buildInputs or []);
	}
	`
	evalExpr := fmt.Sprintf(exprTemplate, nixPkgExpression)
	data, err := nc.nixEval(evalExpr)
	if err != nil {
		return nil, fmt.Errorf("Error getting dependencies for %s: %v", attributePath, err)
	}

	var result map[string][]string
	if err := json.Unmarshal(data, &result); err != nil {
		return []string{}, fmt.Errorf("Error parsing JSON for package '%s': %v", attributePath, err)
	}

	finalInputs := result["buildInputs"]
	return finalInputs, nil
}

func (nc *NixCollector) nixEval(expr string) ([]byte, error) {
	cmd := exec.Command("nix", "eval", "--impure", "--expr", expr, "--extra-experimental-features", "nix-command", "--json")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("Error running nix eval: %v", err)
	}
	return out.Bytes(), nil
}

type WorkerFunc func(worker int)

func WorkerPool(n int, w WorkerFunc) (waitFunc func()) {
	wg := &sync.WaitGroup{}
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(worker int) {
			defer wg.Done()
			w(worker)
		}(i)
	}
	return wg.Wait
}

func NewNixCollector() *NixCollector {
	return &NixCollector{
		CollecterInterface: collector.NewCollector(repository.Nix, repository.DistPackageTablePrefix("nix")),
	}
}
