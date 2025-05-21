package toolimpl

import (
	"fmt"
	"io"

	"github.com/HUSTSecLab/criticality_score/cmd/apiserver/internal/tool"
)

var testTool = &tool.Tool{
	ID:          "862910f9-8ce0-4cd9-a26c-86f84a783967",
	Name:        "测试工具",
	Description: "该工具用于测试功能是否正常。",
	Args: []tool.ToolArg{
		{Name: "arg1", Type: tool.ToolArgTypeString,
			Description: "arg 1 desc", Default: "default"},
		{Name: "arg2", Type: tool.ToolArgTypeInt,
			Description: "arg 2 desc", Default: 1},
		{Name: "arg3", Type: tool.ToolArgTypeBool,
			Description: "arg 3 desc", Default: true},
	},
	Run: tool.CanioalizeWrapper(testImpl),
}

func testImpl(args map[string]any, in io.Reader, out io.Writer, err io.Writer, kill chan int) error {
	fmt.Fprintf(out, "test tool args: %v\n", args)
	fmt.Fprintf(out, "Please input two numbers:\n")
	var a, b int
	_, e := fmt.Fscanf(in, "%d %d", &a, &b)
	if e != nil {
		fmt.Fprintf(err, "Error reading input: %v\n", err)
	}
	fmt.Fprintf(out, "You input: %d %d\n", a, b)
	fmt.Fprintf(err, "Test tool error output\n")
	fmt.Fprintln(out, "Test tool output using fmt.Println")

	return nil
}

func init() {
	tool.RegistTool(testTool)
}
