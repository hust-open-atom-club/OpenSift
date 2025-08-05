package openharmony

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	collector "github.com/HUSTSecLab/OpenSift/pkg/collector/internal"
	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/HUSTSecLab/OpenSift/pkg/storage/repository"
)

type OpenHarmonyCollector struct {
	collector.CollecterInterface
	httpClient *http.Client
	semaphore  chan struct{}
	remoteMap  map[string]string // 存储remote名称到URL的映射
}

func NewOpenHarmonyCollector() *OpenHarmonyCollector {
	return &OpenHarmonyCollector{
		CollecterInterface: collector.NewCollector(repository.OpenHarmony, repository.DistPackageTablePrefix("openharmony")),
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        20,
				IdleConnTimeout:     60 * time.Second,
				DisableCompression:  true,
				MaxConnsPerHost:     10,
				MaxIdleConnsPerHost: 10,
			},
		},
		semaphore: make(chan struct{}, 10),
		remoteMap: make(map[string]string),
	}
}

func (hc *OpenHarmonyCollector) Collect(outputPath, downloadDir string) {
	adc := storage.GetDefaultAppDatabaseContext()
	data := hc.GetPackageInfo(collector.OpenHarmonyURL)
	hc.ParseInfo(data, downloadDir)
	hc.GetDep()
	hc.PageRank(0.85, 20)
	hc.GetDepCount()
	hc.UpdateRelationships(adc)
	hc.UpdateDistRepoCount(adc)
	hc.CalculateDistImpact()
	hc.UpdateOrInsertDatabase(adc)
	hc.UpdateOrInsertDistDependencyDatabase(adc)
	if outputPath != "" {
		err := hc.GenerateDependencyGraph(outputPath)
		if err != nil {
			log.Printf("生成依赖图错误: %v\n", err)
		}
	}
}

func (hc *OpenHarmonyCollector) ParseInfo(data, downloadDir string) {
	log.Println("解析OpenHarmony清单文件...")

	// 解析XML清单文件
	manifest, err := hc.parseManifest(data)
	if err != nil {
		log.Printf("清单解析失败: %v", err)
		return
	}

	log.Printf("发现 %d 个OpenHarmony组件", len(manifest.Projects))

	// 创建缓存目录
	cacheDir := filepath.Join(downloadDir, "openharmony_cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Printf("创建缓存目录失败: %v", err)
	}

	// 处理所有项目
	var wg sync.WaitGroup
	results := make(chan *collector.PackageInfo, len(manifest.Projects))

	for _, project := range manifest.Projects {
		wg.Add(1)
		go func(p Project) {
			defer wg.Done()
			hc.semaphore <- struct{}{}
			defer func() { <-hc.semaphore }()

			pkgInfo := hc.processProject(p, cacheDir)
			results <- pkgInfo
		}(project)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	count := 0
	for pkgInfo := range results {
		if pkgInfo != nil {
			hc.SetPkgInfo(pkgInfo.Name, pkgInfo)
			count++
		}
	}
	log.Printf("成功处理 %d/%d 个组件", count, len(manifest.Projects))
}

func (hc *OpenHarmonyCollector) parseManifest(data string) (*Manifest, error) {
	decoder := xml.NewDecoder(strings.NewReader(data))
	manifest := &Manifest{
		Remotes:  make([]Remote, 0),
		Projects: make([]Project, 0),
	}

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("XML解析错误: %v", err)
		}

		switch se := token.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "remote":
				var remote Remote
				if err := decoder.DecodeElement(&remote, &se); err != nil {
					log.Printf("解析remote失败: %v, 跳过", err)
					continue
				}
				manifest.Remotes = append(manifest.Remotes, remote)
				hc.remoteMap[remote.Name] = remote.Fetch

			case "project":
				var project Project
				if err := decoder.DecodeElement(&project, &se); err != nil {
					log.Printf("解析project失败: %v, 跳过", err)
					continue
				}
				manifest.Projects = append(manifest.Projects, project)
			}
		}
	}
	return manifest, nil
}

func (hc *OpenHarmonyCollector) processProject(project Project, cacheDir string) *collector.PackageInfo {
	// 获取远程仓库URL
	remoteURL, ok := hc.remoteMap[project.Remote]
	if !ok {
		log.Printf("项目 %s 的remote未定义: %s", project.Name, project.Remote)
		return nil
	}

	pkgInfo := &collector.PackageInfo{
		Name:        project.Name,
		Description: fmt.Sprintf("OpenHarmony组件: %s", project.Name),
		Version:     project.Revision,
		Homepage:    fmt.Sprintf("%s/%s", remoteURL, project.Path),
	}

	// 尝试获取依赖
	deps, err := hc.getProjectDependencies(remoteURL, project, cacheDir)
	if err != nil {
		log.Printf("依赖解析失败 %s: %v", project.Name, err)
	} else if len(deps) > 0 {
		pkgInfo.DirectDepends = deps
	}

	return pkgInfo
}

func (hc *OpenHarmonyCollector) getProjectDependencies(baseURL string, project Project, cacheDir string) ([]string, error) {
	// 尝试多种元数据文件
	metadataFiles := []string{
		"bundle.json",
		"ohos.build",
		"BUILD.gn",
		"package.json",
	}

	for _, file := range metadataFiles {
		deps, err := hc.parseMetadataFile(baseURL, project, file, cacheDir)
		if err == nil && len(deps) > 0 {
			return deps, nil
		}
	}
	return nil, fmt.Errorf("未找到有效的元数据文件")
}

func (hc *OpenHarmonyCollector) parseMetadataFile(baseURL string, project Project, filename, cacheDir string) ([]string, error) {
	filePath := filepath.Join(project.Path, filename)
	localPath := filepath.Join(cacheDir, filePath)

	// 尝试从缓存读取
	if data, err := os.ReadFile(localPath); err == nil {
		return parseMetadata(data, filename)
	}

	// 从远程获取
	url := fmt.Sprintf("%s/%s/raw/%s/%s",
		baseURL,
		project.Path,
		project.Revision,
		filename)

	log.Printf("获取 %s 的元数据: %s", project.Name, url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("User-Agent", "OpenSift/1.0")

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP错误: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 保存缓存
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err == nil {
		os.WriteFile(localPath, data, 0644)
	}

	return parseMetadata(data, filename)
}

func parseMetadata(data []byte, filename string) ([]string, error) {
	switch filename {
	case "bundle.json":
		return parseBundleJSON(data)
	case "ohos.build", "BUILD.gn":
		return parseBuildFile(data)
	case "package.json":
		return parsePackageJSON(data)
	default:
		return nil, fmt.Errorf("不支持的元数据类型")
	}
}

func parseBundleJSON(data []byte) ([]string, error) {
	var bundle struct {
		Component struct {
			Dependencies []struct {
				Bundle string `json:"bundle"`
			} `json:"dependencies"`
		} `json:"component"`
	}

	if err := json.Unmarshal(data, &bundle); err != nil {
		return nil, err
	}

	deps := []string{}
	for _, dep := range bundle.Component.Dependencies {
		if dep.Bundle != "" {
			deps = append(deps, dep.Bundle)
		}
	}
	return deps, nil
}

func parseBuildFile(data []byte) ([]string, error) {
	// 简单解析依赖关系
	content := string(data)
	deps := []string{}

	// 查找依赖声明
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(line, "deps") || strings.Contains(line, "dependencies") {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "\"") && strings.HasSuffix(part, "\"") {
					dep := strings.Trim(part, "\"")
					if !strings.Contains(dep, ":") && !strings.Contains(dep, "//") {
						deps = append(deps, dep)
					}
				}
			}
		}
	}
	return deps, nil
}

func parsePackageJSON(data []byte) ([]string, error) {
	var pkg struct {
		Dependencies map[string]string `json:"dependencies"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	deps := []string{}
	for dep := range pkg.Dependencies {
		deps = append(deps, dep)
	}
	return deps, nil
}

func (hc *OpenHarmonyCollector) GetPackageInfo(urls collector.PackageURL) string {
	log.Println("下载OpenHarmony清单文件...")

	for _, url := range urls {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("创建请求失败: %v", err)
			continue
		}
		req.Header.Set("User-Agent", "OpenSift/1.0")

		resp, err := hc.httpClient.Do(req)
		if err != nil {
			log.Printf("下载失败: %v", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("HTTP错误: %d", resp.StatusCode)
			continue
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("读取失败: %v", err)
			continue
		}

		log.Printf("成功下载清单文件 (%d bytes)", len(data))
		return string(data)
	}

	return ""
}

// Manifest 结构定义
type Manifest struct {
	Remotes  []Remote  `xml:"remote"`
	Projects []Project `xml:"project"`
}

type Remote struct {
	Name  string `xml:"name,attr"`
	Fetch string `xml:"fetch,attr"`
}

type Project struct {
	Name     string `xml:"name,attr"`
	Path     string `xml:"path,attr"`
	Revision string `xml:"revision,attr"`
	Remote   string `xml:"remote,attr"`
	Groups   string `xml:"groups,attr"`
}
