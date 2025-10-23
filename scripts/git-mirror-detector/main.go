package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/tealeg/xlsx/v3"
)

// GitHubRepo 结构体用于存储GitHub仓库信息
type GitHubRepo struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Desc     string `json:"description"`
	HTMLURL  string `json:"html_url"`
	CloneURL string `json:"clone_url"`
	Fork     bool   `json:"fork"`
	Parent   *struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		HTMLURL  string `json:"html_url"`
	} `json:"parent"`
	Source *struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		HTMLURL  string `json:"html_url"`
	} `json:"source"`
}

// RepoInfo 用于存储处理后的仓库信息
type RepoInfo struct {
	OriginalURL        string
	RepoName           string
	IsValid            bool
	IsFork             bool
	IsMirror           bool
	ForkOriginalRepo   string
	MirrorOriginalRepo string
	Description        string
	ErrorMessage       string
}

func main() {
	inputFile := "input.xlsx"
	outputFile := "output.xlsx"

	// 读取输入Excel文件
	repos, err := readExcelFile(inputFile)
	if err != nil {
		fmt.Printf("读取Excel文件失败: %v\n", err)
		return
	}

	// 处理每个GitHub链接
	var results []RepoInfo
	for i, repoURL := range repos {
		fmt.Printf("处理第 %d 个链接: %s\n", i+1, repoURL)

		info := processGitHubURL(repoURL)
		results = append(results, info)

		// 添加延迟避免API限制
		time.Sleep(1 * time.Second)
	}

	// 写入输出Excel文件
	err = writeExcelFile(outputFile, results)
	if err != nil {
		fmt.Printf("写入Excel文件失败: %v\n", err)
		return
	}

	fmt.Printf("处理完成！结果已保存到 %s\n", outputFile)
}

// readExcelFile 读取Excel文件中的GitHub链接
func readExcelFile(filename string) ([]string, error) {
	file, err := xlsx.OpenFile(filename)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}

	var urls []string
	if len(file.Sheets) == 0 {
		return nil, fmt.Errorf("Excel文件中没有工作表")
	}

	sheet := file.Sheets[0]

	// 找到 git_link 列的索引
	gitLinkColumnIndex := -1

	// 检查第一行（标题行）找到 git_link 列
	err = sheet.ForEachRow(func(r *xlsx.Row) error {
		// 只处理第一行来找到列索引
		if gitLinkColumnIndex == -1 {
			for i := 0; i < 10; i++ { // 最多检查前10列
				cell := r.GetCell(i)
				if cell != nil {
					cellValue := strings.TrimSpace(strings.ToLower(cell.String()))
					if cellValue == "git_link" || cellValue == "gitlink" {
						gitLinkColumnIndex = i
						break
					}
				}
			}
		} else {
			// 从数据行读取 git_link 列的值
			cell := r.GetCell(gitLinkColumnIndex)
			if cell != nil {
				cellValue := strings.TrimSpace(cell.String())
				if cellValue != "" && strings.Contains(cellValue, "github.com") {
					urls = append(urls, cellValue)
				}
			}
		}
		return nil
	})

	if gitLinkColumnIndex == -1 {
		return nil, fmt.Errorf("未找到 git_link 列")
	}

	return urls, err
}

// writeExcelFile 将结果写入Excel文件
func writeExcelFile(filename string, results []RepoInfo) error {
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Results")
	if err != nil {
		return fmt.Errorf("创建工作表失败: %v", err)
	}

	// 添加标题行
	headerRow := sheet.AddRow()
	headers := []string{"原始链接", "是否有效", "是否Fork", "是否镜像", "Fork的原始仓库", "镜像的原始仓库"}
	for _, header := range headers {
		cell := headerRow.AddCell()
		cell.Value = header
	}

	// 添加数据行
	for _, result := range results {
		row := sheet.AddRow()

		cells := []*xlsx.Cell{
			row.AddCell(), // 原始链接
			row.AddCell(), // 是否有效
			row.AddCell(), // 是否Fork
			row.AddCell(), // 是否镜像
			row.AddCell(), // Fork的原始仓库
			row.AddCell(), // 镜像的原始仓库
		}

		cells[0].Value = result.OriginalURL
		cells[1].Value = fmt.Sprintf("%t", result.IsValid)
		cells[2].Value = fmt.Sprintf("%t", result.IsFork)
		cells[3].Value = fmt.Sprintf("%t", result.IsMirror)
		cells[4].Value = result.ForkOriginalRepo
		cells[5].Value = result.MirrorOriginalRepo
	}

	return file.Save(filename)
}

// processGitHubURL 处理GitHub URL并获取仓库信息
func processGitHubURL(githubURL string) RepoInfo {
	info := RepoInfo{
		OriginalURL: githubURL,
		IsValid:     false,
	}

	// 解析GitHub URL
	owner, repo, err := parseGitHubURL(githubURL)
	if err != nil {
		info.ErrorMessage = err.Error()
		return info
	}

	info.RepoName = fmt.Sprintf("%s/%s", owner, repo)

	// 获取仓库信息
	repoData, err := getGitHubRepoInfo(owner, repo)
	if err != nil {
		info.ErrorMessage = err.Error()
		return info
	}

	info.IsValid = true
	info.Description = repoData.Desc
	info.IsFork = repoData.Fork

	// 如果是fork，获取原始仓库信息
	if repoData.Fork && repoData.Source != nil {
		info.ForkOriginalRepo = repoData.Source.HTMLURL
	}

	// 检查是否为镜像仓库并获取镜像的原始仓库
	info.IsMirror, info.MirrorOriginalRepo = checkIfMirrorRepo(repoData)

	return info
}

// parseGitHubURL 解析GitHub URL获取owner和repo
func parseGitHubURL(githubURL string) (string, string, error) {
	// 正则表达式匹配GitHub URL
	re := regexp.MustCompile(`github\.com/([^/]+)/([^/]+)`)
	matches := re.FindStringSubmatch(githubURL)

	if len(matches) < 3 {
		return "", "", fmt.Errorf("无效的GitHub URL格式")
	}

	owner := matches[1]
	repo := strings.TrimSuffix(matches[2], ".git")

	return owner, repo, nil
}

// getGitHubRepoInfo 通过GitHub API获取仓库信息
func getGitHubRepoInfo(owner, repo string) (*GitHubRepo, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("请求GitHub API失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API返回错误状态: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var repoData GitHubRepo
	err = json.Unmarshal(body, &repoData)
	if err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return &repoData, nil
}

// checkIfMirrorRepo 检查是否为镜像仓库并返回是否为镜像和原始仓库链接
func checkIfMirrorRepo(repoData *GitHubRepo) (bool, string) {
	if repoData.Desc == "" {
		return false, ""
	}

	desc := strings.ToLower(repoData.Desc)

	// 检查描述中是否包含镜像相关关键词
	if strings.Contains(desc, "mirror") ||
		strings.Contains(desc, "镜像") ||
		strings.Contains(desc, "clone") ||
		strings.Contains(desc, "copy") ||
		strings.Contains(desc, "backup") ||
		strings.Contains(desc, "同步") ||
		strings.Contains(desc, "复制") {

		// 尝试从描述中提取原始仓库链接
		// 优先查找"mirror of"后面的链接（支持任何代码托管平台）
		mirrorOfRe := regexp.MustCompile(`(?i)mirror\s+of\s+(https?://[^\s]+)`)
		matches := mirrorOfRe.FindStringSubmatch(repoData.Desc)

		if len(matches) >= 2 {
			// 返回"mirror of"后面找到的完整链接
			return true, matches[1]
		}

		// 如果没有找到"mirror of"格式，尝试查找任何代码托管平台的链接
		generalRe := regexp.MustCompile(`https?://(?:github\.com|gitlab\.com|gitee\.com|bitbucket\.org|codeberg\.org|git\.sr\.ht)/[^\s]+`)
		matches = generalRe.FindStringSubmatch(repoData.Desc)

		if len(matches) >= 1 {
			// 返回找到的代码托管平台链接
			return true, matches[0]
		}

		// 如果没有找到具体链接，但确实是镜像，返回true和空字符串
		return true, ""
	}

	return false, ""
}
