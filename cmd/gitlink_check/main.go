package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/xuri/excelize/v2"
)

func main() {
	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("请提供XLSX文件路径作为参数")
		fmt.Println("示例: ./url-checker input.xlsx")
		return
	}
	filePath := os.Args[1]

	// 打开Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		fmt.Printf("打开文件失败: %v\n", err)
		return
	}
	defer f.Close()

	// 获取第一个工作表名
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		fmt.Println("未找到工作表")
		return
	}

	// 读取git_link列的所有单元格
	rows, err := f.GetRows(sheetName)
	if err != nil {
		fmt.Printf("读取工作表失败: %v\n", err)
		return
	}

	// 查找git_link列的索引
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
		fmt.Println("未找到git_link列")
		return
	}

	// 收集URL（跳过标题行）
	var urls []string
	for i, row := range rows {
		if i == 0 { // 跳过标题行
			continue
		}
		if colIndex < len(row) {
			url := row[colIndex]
			if url != "" {
				urls = append(urls, url)
			}
		}
	}

	if len(urls) == 0 {
		fmt.Println("未找到有效URL")
		return
	}

	fmt.Printf("共发现 %d 个URL，开始检查连通性...\n", len(urls))

	// 创建带超时的HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 并发控制
	var wg sync.WaitGroup
	results := make(chan string, len(urls))
	failedUrls := make(chan string, len(urls)) // 新增：用于收集失败的URL
	semaphore := make(chan struct{}, 10)       // 限制并发数为10

	// 启动goroutine检查每个URL
	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()

			// 创建请求
			req, err := http.NewRequestWithContext(ctx, "HEAD", u, nil)
			if err != nil {
				results <- fmt.Sprintf("[错误] %s: 创建请求失败 - %v", u, err)
				failedUrls <- u // 记录失败
				return
			}

			// 发送请求
			resp, err := client.Do(req)
			if err != nil {
				results <- fmt.Sprintf("[失败] %s: 请求失败 - %v", u, err)
				failedUrls <- u // 记录失败
				return
			}
			defer resp.Body.Close()

			// 检查状态码 (2xx/3xx 视为成功)
			if resp.StatusCode >= 200 && resp.StatusCode < 400 {
				results <- fmt.Sprintf("[成功] %s: 状态码 %d", u, resp.StatusCode)
			} else {
				results <- fmt.Sprintf("[失败] %s: 状态码 %d", u, resp.StatusCode)
				failedUrls <- u // 记录失败
			}
		}(url)
	}

	// 等待所有任务完成
	go func() {
		wg.Wait()
		close(results)
		close(failedUrls) // 新增：关闭失败通道
	}()

	// 打印结果
	for res := range results {
		fmt.Println(res)
	}
	fmt.Println("所有URL检查完成")

	// 只收集所有失败的URL并写入 output.xlsx
	var failedList []string
	for u := range failedUrls {
		failedList = append(failedList, u)
	}

	if len(failedList) > 0 {
		outF := excelize.NewFile()
		sheet := outF.GetSheetName(0)
		outF.SetCellValue(sheet, "A1", "failed_git_link")
		for i, u := range failedList {
			outF.SetCellValue(sheet, fmt.Sprintf("A%d", i+2), u)
		}
		if err := outF.SaveAs("output.xlsx"); err != nil {
			fmt.Printf("保存 output.xlsx 失败: %v\n", err)
		} else {
			fmt.Printf("已将 %d 个失败URL保存到 output.xlsx\n", len(failedList))
		}
	} else {
		fmt.Println("没有失败的URL，无需生成 output.xlsx")
	}
}
