/*
Git链接检查器
---------------------------------
批量检查 Excel 文件中的 Git 仓库链接可访问性，支持 HTTP 和 Git 两种方式。
结果实时输出，并将失败链接保存到新的 Excel 文件。

主要功能：
1. 读取 Excel 文件 git_link 列的 URL
2. 并发检查 HTTP 和 Git 连通性
3. 输出检查结果和失败报告

作者：zjy zhuyanhuazhuyanhua
日期：2025年10月21日
*/

package main

import (
	"bytes"    // 字节缓冲区
	"context"  // 控制超时和取消
	"fmt"      // 格式化输出
	"log"      // 日志
	"net/http" // HTTP 客户端
	"os"       // 系统接口
	"os/exec"  // 外部命令
	"strings"  // 字符串处理
	"sync"     // 并发同步
	"time"     // 时间相关

	"github.com/xuri/excelize/v2" // Excel库
)

// 单个URL的检查结果，包含HTTP和Git两种方式
type URLCheckResult struct {
	URL           string // 检查的URL
	HTTPSuccess   bool   // HTTP是否成功
	HTTPError     string // HTTP错误或成功信息
	GitSuccess    bool   // Git是否成功
	GitError      string // Git错误或成功信息
	OverallStatus string // 综合状态
}

// 检查HTTP连通性，优先HEAD，失败再GET，3秒超时
func checkHTTPConnectivity(ctx context.Context, url string) (bool, string) {
	client := &http.Client{
		Timeout: 3 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("重定向次数过多")
			}
			return nil
		},
	}
	methods := []string{"HEAD", "GET"}
	for i, method := range methods {
		req, err := http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			if i == len(methods)-1 {
				return false, fmt.Sprintf("创建请求失败: %v", err)
			}
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		if method == "GET" {
			req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		}
		resp, err := client.Do(req)
		if err != nil {
			if i == len(methods)-1 {
				return false, fmt.Sprintf("请求失败 (%s): %v", method, err)
			}
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			return true, fmt.Sprintf("状态码 %d (%s)", resp.StatusCode, method)
		}
		if i == len(methods)-1 {
			return false, fmt.Sprintf("状态码 %d (%s)", resp.StatusCode, method)
		}
	}
	return false, "所有HTTP方法都失败"
}

// 使用 git ls-remote 检查Git仓库可访问性，3秒超时
func checkGitRemote(ctx context.Context, repoURL string) (bool, string) {
	if repoURL == "" {
		return false, "URL 不能为空"
	}
	// 自动补.git后缀
	if (strings.HasPrefix(repoURL, "https://github.com/") ||
		strings.HasPrefix(repoURL, "https://gitlab.") ||
		strings.HasPrefix(repoURL, "https://sourceware.org/") ||
		strings.HasPrefix(repoURL, "https://gcc.gnu.org/")) &&
		!strings.HasSuffix(repoURL, ".git") {
		repoURL += ".git"
	}
	gitCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(gitCtx, "git", "ls-remote", repoURL)
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_SSH_COMMAND=ssh -o ConnectTimeout=5",
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	errorMsg := stderr.String()
	if err != nil {
		if strings.Contains(errorMsg, "warning: redirecting to") {
			return true, "Git仓库可访问（重定向）"
		}
		if !strings.Contains(errorMsg, "not found") && !strings.Contains(errorMsg, "fatal") {
			return true, "Git仓库可访问（exit status 非0但无致命错误）"
		}
		if errorMsg == "" {
			return false, fmt.Sprintf("命令执行失败: %v", err)
		}
		return false, fmt.Sprintf("命令执行失败: %v, 错误详情: %s", err, strings.TrimSpace(errorMsg))
	}
	return true, "Git仓库可访问"
}

// 并发检查单个URL的HTTP和Git连通性，3秒超时
func checkURL(url string) URLCheckResult {
	result := URLCheckResult{URL: url}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		result.HTTPSuccess, result.HTTPError = checkHTTPConnectivity(ctx, url)
	}()
	go func() {
		defer wg.Done()
		result.GitSuccess, result.GitError = checkGitRemote(ctx, url)
	}()
	wg.Wait()
	if result.HTTPSuccess && result.GitSuccess {
		result.OverallStatus = "完全成功"
	} else if !result.HTTPSuccess && !result.GitSuccess {
		result.OverallStatus = "完全失败"
	} else if !result.HTTPSuccess {
		result.OverallStatus = "HTTP失败"
	} else {
		result.OverallStatus = "Git失败"
	}
	return result
}

// 主流程：读取Excel，批量检查URL，输出结果和失败报告
func main() {
	// 检查Git命令
	if _, err := exec.LookPath("git"); err != nil {
		log.Fatalf("错误: 'git' 命令未找到或不在系统的 PATH 中。请先安装 Git。")
	}
	// 检查参数
	if len(os.Args) < 2 {
		fmt.Println("请提供XLSX文件路径作为参数")
		fmt.Println("示例: go run gitlink_check.go input.xlsx")
		return
	}
	filePath := os.Args[1]
	// 打开Excel
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		fmt.Printf("打开文件失败: %v\n", err)
		return
	}
	defer f.Close()
	// 获取第一个工作表
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		fmt.Println("未找到工作表")
		return
	}
	// 读取所有行
	rows, err := f.GetRows(sheetName)
	if err != nil {
		fmt.Printf("读取工作表失败: %v\n", err)
		return
	}
	// 查找git_link列
	colIndex := -1
	if len(rows) > 0 {
		for i, cell := range rows[0] {
			if cell == "git_link" {
				colIndex = i
				break
			}
		}
	}
	if colIndex == -1 {
		fmt.Println("在工作表中未找到 'git_link' 列")
		return
	}
	// 收集所有URL
	var urls []string
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if colIndex < len(row) {
			url := strings.TrimSpace(row[colIndex])
			if url != "" {
				urls = append(urls, url)
			}
		}
	}
	if len(urls) == 0 {
		fmt.Println("在 'git_link' 列中未找到任何有效的 URL")
		return
	}
	fmt.Printf("共发现 %d 个Git链接，开始检查HTTP连通性和Git可访问性...\n\n", len(urls))
	// 并发检查
	var wg sync.WaitGroup
	results := make(chan URLCheckResult, len(urls))
	semaphore := make(chan struct{}, 10) // 限制并发数为10
	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			result := checkURL(u)
			results <- result
		}(url)
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	// 输出结果
	var allResults []URLCheckResult
	var failedResults []URLCheckResult
	for result := range results {
		allResults = append(allResults, result)
		if result.OverallStatus == "完全成功" {
			fmt.Printf("[成功 ✅] %s: HTTP和Git都可访问\n", result.URL)
		} else {
			fmt.Printf("[失败 ❌] %s: %s\n", result.URL, result.OverallStatus)
			fmt.Printf("    HTTP: %s, Git: %s\n",
				getStatusText(result.HTTPSuccess, result.HTTPError),
				getStatusText(result.GitSuccess, result.GitError))
			failedResults = append(failedResults, result)
		}
	}
	fmt.Printf("\n所有链接检查完成。成功: %d, 失败: %d\n",
		len(allResults)-len(failedResults), len(failedResults))
	// 保存失败报告
	if len(failedResults) > 0 {
		outF := excelize.NewFile()
		sheet := outF.GetSheetName(0)
		outF.SetSheetName(sheet, "FailedLinks")
		outF.SetCellValue("FailedLinks", "A1", "failed_git_link")
		outF.SetCellValue("FailedLinks", "B1", "failure_type")
		outF.SetCellValue("FailedLinks", "C1", "http_status")
		outF.SetCellValue("FailedLinks", "D1", "git_status")
		for i, result := range failedResults {
			row := i + 2
			outF.SetCellValue("FailedLinks", fmt.Sprintf("A%d", row), result.URL)
			outF.SetCellValue("FailedLinks", fmt.Sprintf("B%d", row), result.OverallStatus)
			outF.SetCellValue("FailedLinks", fmt.Sprintf("C%d", row),
				getStatusText(result.HTTPSuccess, result.HTTPError))
			outF.SetCellValue("FailedLinks", fmt.Sprintf("D%d", row),
				getStatusText(result.GitSuccess, result.GitError))
		}
		outputFilename := "failed_links_output.xlsx"
		if err := outF.SaveAs(outputFilename); err != nil {
			fmt.Printf("保存 %s 失败: %v\n", outputFilename, err)
		} else {
			fmt.Printf("已将 %d 个失败的链接保存到 %s\n", len(failedResults), outputFilename)
		}
	} else {
		fmt.Println("所有链接均通过HTTP和Git检查，未生成失败报告文件。")
	}
}

// 格式化输出检查状态
func getStatusText(success bool, errorMsg string) string {
	if success {
		return "成功: " + errorMsg
	}
	return "失败: " + errorMsg
}
