package tool

import (
	"io"
)

type ToolRunner func(args map[string]any, in io.Reader, out io.Writer, kill chan int, resize chan ResizeArg) (int, error)

type Tool struct {
	ID          string
	Name        string
	Description string
	Args        []ToolArg
	Run         ToolRunner
}

type ToolArgType string

const (
	ToolArgTypeString ToolArgType = "string"
	ToolArgTypeInt    ToolArgType = "int"
	ToolArgTypeFloat  ToolArgType = "float"
	ToolArgTypeBool   ToolArgType = "bool"
)

type ToolArg struct {
	Name        string
	Type        ToolArgType
	Description string
	Default     any
}

var toolset = make(map[string]*Tool)

func RegistTool(tool *Tool) {
	toolset[tool.ID] = tool
}

func GetToolList() []*Tool {
	tools := make([]*Tool, 0, len(toolset))
	for _, tool := range toolset {
		tools = append(tools, tool)
	}
	return tools
}

func GetTool(id string) (*Tool, error) {
	tool, ok := toolset[id]
	if !ok {
		return nil, nil
	}
	return tool, nil
}
