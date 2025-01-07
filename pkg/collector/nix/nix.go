package nix

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"unicode"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/lib/pq"
)

// DepInfo struct to store package information
type DepInfo struct {
	Name        string
	Version     string
	Homepage    string
	Description string
	GitLink     string
	DepCount    int
	PageRank    float64
}

func storeDependenciesInDatabase(pkgName string, dependencies []DepInfo) error {
	db, err := storage.GetDefaultDatabaseConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	for _, dep := range dependencies {
		_, err := db.Exec("INSERT INTO nix_relationships (frompackage, topackage) VALUES ($1, $2)", pkgName, dep.Name)
		if err != nil {
			return err
		}
	}
	return nil
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

// attributePathToNixExpression converts an attribute path to a valid Nix expression
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

// getAllNixPackages retrieves all Nix packages as DepInfo and their dependencies
func GetAllNixPackages(poolsize int) (map[DepInfo][]DepInfo, error) {
	cmd := exec.Command("nix-env", "-qaP")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Error running nix-env command: %v", err)
	}

	packages := make(map[DepInfo][]DepInfo)
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

				packageInfo, err := GetNixPackageInfo(attributePath)
				if err != nil {
					fmt.Printf("Error getting info for %s: %v\n", attributePath, err)
					continue
				}

				pkgDepInfo := DepInfo{
					Name:        packageName,
					Version:     packageVersion,
					Homepage:    packageInfo.Homepage,
					Description: packageInfo.Description,
					GitLink:     packageInfo.GitLink,
				}

				dependencies, err := GetNixPackageDependencies(attributePath)
				if err != nil {
					fmt.Printf("Error getting dependencies for %s: %v\n", attributePath, err)
					continue
				}
				mu.Lock()
				packages[pkgDepInfo] = dependencies
				mu.Unlock()
			}
		}
	})

	wg()
	return packages, nil
}

// getNixPackageInfo retrieves the package information using attribute path
func GetNixPackageInfo(attributePath string) (DepInfo, error) {
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
  gitLink = if srcUrl != "" then srcUrl else passthruUrl;
in
{
  name = pname;
  version = version;
  homepage = homepage;
  description = description;
  gitLink = gitLink;
}
`, nixPkgExpression)

	cmd := exec.Command("nix", "eval", "--impure", "--expr", expr, "--extra-experimental-features", "nix-command", "--json")
	var out bytes.Buffer
	var outErr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &outErr
	err := cmd.Run()
	if err != nil {
		return DepInfo{}, fmt.Errorf("Error running nix eval for package '%s': %v\nNix error output:\n%s", attributePath, err, outErr.String())
	}

	var result map[string]string
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return DepInfo{}, fmt.Errorf("Error parsing JSON for package '%s': %v", attributePath, err)
	}

	depInfo := DepInfo{
		Name:        result["name"],
		Version:     result["version"],
		Homepage:    result["homepage"],
		Description: result["description"],
		GitLink:     result["gitLink"],
	}

	depInfo.GitLink = processGitLink(depInfo.GitLink)

	return depInfo, nil
}

func GetNixPackageDependencies(attributePath string) ([]DepInfo, error) {
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
	results, err := nixEval(evalExpr)
	if err != nil {
		return nil, fmt.Errorf("Error getting dependencies for %s: %v", attributePath, err)
	}

	buildInputNames := results["buildInputs"]
	finalInputs := []DepInfo{}
	for _, name := range buildInputNames {
		finalInputs = append(finalInputs, DepInfo{Name: name.Name})
	}

	return finalInputs, nil
}

// nixEval executes a Nix expression and parses the JSON output into []DepInfo
func nixEval(expr string) (map[string][]DepInfo, error) {
	cmd := exec.Command("nix", "eval", "--impure", "--expr", expr, "--extra-experimental-features", "nix-command", "--json")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("Error running nix eval: %v", err)
	}

	// 修改为 map[string][]string 以匹配 JSON 结构
	var result map[string][]string
	// fmt.Println(string(out.Bytes())) // 打印输出以便调试
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("Error parsing JSON: %v", err)
	}

	depsMap := make(map[string][]DepInfo)

	// 构建依赖映射
	for key, depList := range result {
		for _, depName := range depList {
			depInfo := DepInfo{
				Name: depName, // 只存储名称
			}
			depsMap[key] = append(depsMap[key], depInfo)
		}
	}

	return depsMap, nil
}

// processGitLink processes the gitLink to ensure it points to a git repository
func processGitLink(gitLink string) string {
	if gitLink == "" {
		return ""
	}

	parsedURL, err := url.Parse(gitLink)
	if err != nil {
		return ""
	}

	codeHostingDomains := []string{
		"github.com",
		"gitlab.com",
		"bitbucket.org",
	}

	for _, domain := range codeHostingDomains {
		if strings.Contains(parsedURL.Host, domain) {
			return gitLink
		}
	}

	return ""
}

func mergeDependencies(packages map[DepInfo][]DepInfo) map[DepInfo][]DepInfo {
	mergedPackages := make(map[DepInfo][]DepInfo)

	for pkg, deps := range packages {
		merged := false

		// 遍历已合并的包，查找相同包名
		for existingPkg := range mergedPackages {
			if existingPkg.Name == pkg.Name {
				// 合并版本信息
				versionSet := make(map[string]struct{})
				versionSet[existingPkg.Version] = struct{}{}
				versionSet[pkg.Version] = struct{}{}
				pkg.Version = strings.Join(getKeys(versionSet), ",") // 更新版本信息

				// 合并依赖
				mergedDeps := make(map[string]DepInfo)
				for _, dep := range mergedPackages[existingPkg] {
					mergedDeps[dep.Name] = dep
				}
				for _, dep := range deps {
					mergedDeps[dep.Name] = dep // 保留最新的依赖信息
				}

				// 更新合并后的依赖列表
				mergedDepsList := make([]DepInfo, 0, len(mergedDeps))
				for _, dep := range mergedDeps {
					mergedDepsList = append(mergedDepsList, dep)
				}
				mergedPackages[existingPkg] = mergedDepsList
				merged = true
				break
			}
		}

		// 如果没有找到相同的包名，则直接添加
		if !merged {
			mergedPackages[pkg] = deps
		}
	}

	return mergedPackages
}

func getAllDep(packages map[string][]string, pkgName string, visited map[string]bool, deps []string) []string {
	if visited[pkgName] {
		return deps
	}

	visited[pkgName] = true
	deps = append(deps, pkgName)

	for _, dep := range packages[pkgName] {
		if !visited[dep] {
			deps = getAllDep(packages, dep, visited, deps)
		}
	}
	return deps
}

func calculatePageRank(packages map[DepInfo][]DepInfo, iterations int, dampingFactor float64) map[string]float64 {
	ranks := make(map[string]float64)
	numPackages := float64(len(packages))

	for pkg := range packages {
		ranks[pkg.Name] = 1.0 / numPackages
	}

	for i := 0; i < iterations; i++ {
		newRanks := make(map[string]float64)
		for pkg := range packages {
			newRanks[pkg.Name] = (1.0 - dampingFactor) / numPackages
		}

		for pkg, deps := range packages {
			depCount := 0
			for _, dep := range deps {
				if _, exists := packages[dep]; exists {
					depCount++
				}
			}
			contribution := (dampingFactor * ranks[pkg.Name]) / float64(depCount)
			for _, dep := range deps {
				if _, exists := packages[dep]; exists {
					newRanks[dep.Name] += contribution
				}
			}
		}

		ranks = newRanks
	}

	return ranks
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getKeys(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	return keys
}

func SavePackage(packages map[DepInfo][]DepInfo) error {
	file, err := os.Create("packages.gob")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	return encoder.Encode(packages)
}

func LoadPackage() (map[DepInfo][]DepInfo, error) {
	file, err := os.Open("packages.gob")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var packages map[DepInfo][]DepInfo
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&packages)
	if err != nil {
		return nil, err
	}

	return packages, nil
}

func GetNixPackageList() ([]DepInfo, error) {
	cmd := exec.Command("nix-env", "-qaP")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Error running nix-env command: %v", err)
	}

	var packages []DepInfo
	lines := strings.Split(string(out), "\n")

	re := regexp.MustCompile(`^nixpkgs\.(.+?)\s+([^\s]+)$`)
	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.Contains(line, "evaluation warning") {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			packages = append(packages, DepInfo{Name: matches[1], Version: matches[2]})
		}
	}
	return packages, nil
}

func reverseDependencies(deps map[DepInfo][]DepInfo) map[DepInfo][]DepInfo {
	reversed := make(map[DepInfo][]DepInfo)
	for key, values := range deps {
		for _, dep := range values {
			simplifiedDep := DepInfo{Name: dep.Name}
			reversed[simplifiedDep] = append(reversed[simplifiedDep], key)
		}
	}
	return reversed
}

func Nix(workerCount int, batchSize int) {
	// packages, err := LoadPackage()
	// if err != nil {
	//     fmt.Printf("Error loading package list: %v\n", err)
	//     return
	// }

	// if packages == nil {
	packages, err := GetAllNixPackages(workerCount)
	if err != nil {
		fmt.Printf("Error retrieving Nix packages: %v\n", err)
		return
	}

	//     if err := SavePackage(packages); err != nil {
	//         fmt.Printf("Error saving package list: %v\n", err)
	//         return
	//     }
	// }
	fmt.Println("Nix package information retrieved successfully")
	countDependencies(packages)

	pageRanks := calculatePageRank(packages, 20, 0.85)

	for pkgInfo := range packages {
		pkgInfo.PageRank = pageRanks[pkgInfo.Name]
		packages[pkgInfo] = packages[pkgInfo]
	}

	fmt.Println("Nix package information updated successfully")

	if err := batchupdateOrInsertNixPackages(packages, batchSize); err != nil {
		fmt.Printf("Error updating or inserting Nix packages into database: %v\n", err)
		return
	}

	for pkg, pkgInfo := range packages {
		if err := storeDependenciesInDatabase(pkg.Name, pkgInfo); err != nil {
			if isUniqueViolation(err) {
				continue
			}
			fmt.Printf("Error storing dependencies for package %s: %v\n", pkg.Name, err)
		}
	}

	fmt.Println("Successfully updated package information in the database")
}

func batchupdateOrInsertNixPackages(packages map[DepInfo][]DepInfo, batchSize int) error {
	db, err := storage.GetDefaultDatabaseConnection()
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}
	defer db.Close()

	var packageList []DepInfo
	seen := make(map[string]bool)

	for pkg := range packages {
		if !seen[pkg.Name] {
			packageList = append(packageList, pkg)
			seen[pkg.Name] = true
		}
	}

	for i := 0; i < len(packageList); i += batchSize {
		end := i + batchSize
		if end > len(packageList) {
			end = len(packageList)
		}
		batch := packageList[i:end]

		if err := updateOrInsertBatch(db, batch); err != nil {
			return fmt.Errorf("error processing batch: %w", err)
		}
	}

	return nil
}

func updateOrInsertBatch(db *sql.DB, batch []DepInfo) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO nix_packages (package, version, homepage, description, depends_count, page_rank)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (package) DO UPDATE
		SET version = EXCLUDED.version,
			homepage = EXCLUDED.homepage,
			description = EXCLUDED.description,
			depends_count = EXCLUDED.depends_count,
			page_rank = EXCLUDED.page_rank
	`

	for _, pkg := range batch {
		_, err := tx.Exec(query, pkg.Name, pkg.Version, pkg.Homepage, pkg.Description, pkg.DepCount, pkg.PageRank)
		if err != nil {
			return fmt.Errorf("error inserting or updating package %s: %w", pkg.Name, err)
		}
	}

	return tx.Commit()
}

func findDepInfoByName(packages map[DepInfo][]DepInfo, name string) (DepInfo, bool) {
	for dep := range packages {
		if dep.Name == name {
			return dep, true
		}
	}
	return DepInfo{}, false
}

func normalizeGitLink(link string) string {
	var orgName, repoName string

	if strings.HasSuffix(link, ".git") {
		link = link[:len(link)-4]
	}

	if strings.HasPrefix(link, "https://github.com/") {
		parts := strings.Split(link, "/")
		if len(parts) >= 5 {
			orgName = parts[3]
			repoName = parts[4]
			return fmt.Sprintf("https://github.com/%s/%s.git", orgName, repoName)
		}
	} else if strings.HasPrefix(link, "http://github.com/") {
		parts := strings.Split(link, "/")
		if len(parts) >= 5 {
			orgName = parts[3]
			repoName = parts[4]
			return fmt.Sprintf("http://github.com/%s/%s.git", orgName, repoName)
		}
	} else if strings.HasPrefix(link, "https://gitlab.com/") {
		parts := strings.Split(link, "/")
		if len(parts) >= 5 {
			orgName = parts[3]
			repoName = parts[4]
			return fmt.Sprintf("https://gitlab.com/%s/%s.git", orgName, repoName)
		}
	} else if strings.HasPrefix(link, "http://gitlab.com/") {
		parts := strings.Split(link, "/")
		if len(parts) >= 5 {
			orgName = parts[3]
			repoName = parts[4]
			return fmt.Sprintf("http://gitlab.com/%s/%s.git", orgName, repoName)
		}
	} else if strings.HasPrefix(link, "https://gitee.com/") {
		parts := strings.Split(link, "/")
		if len(parts) >= 5 {
			orgName = parts[3]
			repoName = parts[4]
			return fmt.Sprintf("https://gitee.com/%s/%s.git", orgName, repoName)
		}
	} else if strings.HasPrefix(link, "http://gitee.com/") {
		parts := strings.Split(link, "/")
		if len(parts) >= 5 {
			orgName = parts[3]
			repoName = parts[4]
			return fmt.Sprintf("http://gitee.com/%s/%s.git", orgName, repoName)
		}
	} else if strings.HasPrefix(link, "https://bitbucket.org/") {
		parts := strings.Split(link, "/")
		if len(parts) >= 5 {
			orgName = parts[3]
			repoName = parts[4]
			return fmt.Sprintf("https://bitbucket.org/%s/%s", orgName, repoName)
		}
	} else if strings.HasPrefix(link, "http://bitbucket.org/") {
		parts := strings.Split(link, "/")
		if len(parts) >= 5 {
			orgName = parts[3]
			repoName = parts[4]
			return fmt.Sprintf("http://bitbucket.org/%s/%s", orgName, repoName)
		}
	}

	return "" // 如果不符合任何协议，返回空字符串
}

func countDependencies(packages map[DepInfo][]DepInfo) {
	countMap := make(map[string]int)
	depMap := make(map[string][]string)

	deporigMap := make(map[string][]string)

	for key, list := range packages {
		for _, value := range list {
			deporigMap[key.Name] = append(deporigMap[key.Name], value.Name)
		}
	}

	for pkgInfo := range packages {
		visited := make(map[string]bool)
		deps := getAllDep(deporigMap, pkgInfo.Name, visited, []string{})
		depMap[pkgInfo.Name] = deps
	}

	for _, deps := range depMap {
		for _, dep := range deps {
			countMap[dep]++
		}
	}

	for pkgInfo := range packages {
		depCount := countMap[pkgInfo.Name]
		pkgInfo.DepCount = depCount
		packages[pkgInfo] = packages[pkgInfo]
	}
}

func isUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
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
