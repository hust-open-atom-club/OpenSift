package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/HUSTSecLab/OpenSift/pkg/function-track/common"
)

type GoParser struct{}

// NewGoParser 创建新的Go语言解析器
func NewGoParser() Parser {
	return &GoParser{}
}

// Parse 解析Go代码
func (p *GoParser) Parse(filePath string, content []byte) (ParseResult, error) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("解析Go文件失败: %w", err)
	}

	result := &BaseParseResult{
		Functions:    make([]Function, 0),
		CommentLines: 0,
		TotalLines:   strings.Count(string(content), "\n") + 1,
		Language:     common.Go,
	}

	// 计算注释行数
	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			result.CommentLines += strings.Count(comment.Text, "\n") + 1
		}
	}

	// 分析函数
	ast.Inspect(file, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			startPos := fileSet.Position(funcDecl.Pos())
			endPos := fileSet.Position(funcDecl.End())

			function := Function{
				Name:      funcDecl.Name.Name,
				StartLine: startPos.Line,
				EndLine:   endPos.Line,
				//Complexity: calculateComplexity(funcDecl),
				Parameters: calculateParameters(funcDecl),
			}

			result.Functions = append(result.Functions, function)
		}
		return true
	})

	return result, nil
}

// SupportedLanguages 返回支持的语言类型
func (p *GoParser) SupportedLanguages() []common.LanguageType {
	return []common.LanguageType{common.Go}
}

// // calculateComplexity 计算函数的循环复杂度
// func calculateComplexity(funcDecl *ast.FuncDecl) int {
// 	complexity := 1

// 	ast.Inspect(funcDecl, func(n ast.Node) bool {
// 		switch node := n.(type) {
// 		case *ast.IfStmt:
// 			complexity++
// 		case *ast.ForStmt, *ast.RangeStmt:
// 			complexity++
// 		case *ast.CaseClause:
// 			complexity++
// 		case *ast.BinaryExpr:
// 			if node.Op == token.LAND || node.Op == token.LOR {
// 				complexity++
// 			}
// 		}
// 		return true
// 	})

// 	return complexity
// }

// calculateParameters 计算函数参数数量
func calculateParameters(funcDecl *ast.FuncDecl) int {
	if funcDecl.Type.Params == nil {
		return 0
	}

	count := 0
	for _, field := range funcDecl.Type.Params.List {
		count += len(field.Names)
	}

	return count
}
