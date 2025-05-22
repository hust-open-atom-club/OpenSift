package tool

import (
	"io"
)

type ToolRunner func(args map[string]any, in io.Reader, out io.Writer, kill chan int, resize chan ResizeArg) (int, error)

type Tool struct {
	ID           string
	Name         string
	Description  string
	Group        string
	Args         []ToolArg
	AllowSignals []ToolSignal
	Run          ToolRunner
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

var (
	ToolSignalTemplateKill = &ToolSignal{
		Value:       9,
		Name:        "SIGKILL",
		Description: "强制杀死进程",
	}
	ToolSignalTemplateInt = &ToolSignal{
		Value:       2,
		Name:        "SIGINT",
		Description: "CTRL+C 优雅退出",
	}
	ToolSignalTemplateTerm = &ToolSignal{
		Value:       15,
		Name:        "SIGTERM",
		Description: "请求终止进程",
	}
)

type ToolSignal struct {
	Value       int
	Name        string
	Description string
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
