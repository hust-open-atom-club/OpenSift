package common

import (
	"os"
	"path/filepath"
	"strings"
)

// FindSourceFiles 在指定目录中查找源代码文件
// 参数:
//   - rootDir: 根目录路径
//   - includePatterns: 包含模式列表 (支持glob模式，如 "*.go")
//   - excludePatterns: 排除模式列表 (支持glob模式，如 "vendor/**")
//   - progressCallback: 进度回调函数，用于报告已找到的文件数量
//
// 返回值:
//   - []string: 符合条件的文件路径列表
//   - error: 可能的错误
func FindSourceFiles(
	rootDir string,
	includePatterns []string,
	excludePatterns []string,
	progressCallback func(found int),
) ([]string, error) {
	var files []string
	detector := NewLanguageDetector()

	// 处理根目录为当前目录的情况
	if rootDir == "." {
		absPath, err := filepath.Abs(rootDir)
		if err == nil {
			rootDir = absPath
		}
	}

	// 遍历目录查找文件
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			// 检查是否应该跳过此目录
			if shouldSkipDir(path, rootDir, excludePatterns) {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查是否为支持的文件类型
		if !detector.IsSupportedFile(path) {
			return nil
		}

		// 检查是否符合包含/排除模式
		if shouldIncludeFile(path, rootDir, includePatterns, excludePatterns) {
			files = append(files, path)

			// 报告进度
			if progressCallback != nil {
				progressCallback(len(files))
			}
		}

		return nil
	})

	return files, err
}

// shouldSkipDir 判断是否应该跳过目录
func shouldSkipDir(path, rootDir string, excludePatterns []string) bool {
	// 跳过隐藏目录
	if isHiddenDir(path) {
		return true
	}

	// 检查排除模式
	return matchesAnyPattern(path, rootDir, excludePatterns)
}

// shouldIncludeFile 判断是否应该包含文件
func shouldIncludeFile(path, rootDir string, includePatterns, excludePatterns []string) bool {
	if matchesAnyPattern(path, rootDir, excludePatterns) {
		return false
	}

	if len(includePatterns) == 0 {
		return true
	}

	return matchesAnyPattern(path, rootDir, includePatterns)
}

// matchesAnyPattern 检查路径是否匹配任一模式
func matchesAnyPattern(path, rootDir string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}

	// 获取相对路径，用于匹配
	relPath, err := filepath.Rel(rootDir, path)
	if err != nil {
		relPath = path
	}

	// 标准化路径分隔符
	relPath = filepath.ToSlash(relPath)

	for _, pattern := range patterns {
		// 标准化模式分隔符
		pattern = filepath.ToSlash(pattern)

		// 尝试直接匹配
		if matched, _ := filepath.Match(pattern, relPath); matched {
			return true
		}

		// 处理 **/ 模式
		if strings.Contains(pattern, "**/") {
			parts := strings.Split(pattern, "**/")
			if len(parts) == 2 {
				prefix := parts[0]
				suffix := parts[1]

				// 检查前缀匹配
				if prefix == "" || strings.HasPrefix(relPath, prefix) {
					// 检查后缀匹配
					if suffix == "" || strings.HasSuffix(relPath, suffix) || matchesPattern(relPath, suffix) {
						return true
					}
				}
			}
		}
	}

	return false
}

// matchesPattern 检查路径是否匹配模式
func matchesPattern(path, pattern string) bool {
	matched, _ := filepath.Match(pattern, filepath.Base(path))
	return matched
}

// isHiddenDir 判断是否为隐藏目录
func isHiddenDir(path string) bool {
	base := filepath.Base(path)
	return strings.HasPrefix(base, ".") && base != "."
}
