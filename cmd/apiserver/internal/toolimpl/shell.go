package toolimpl

import (
	"io"
	"os"

	"github.com/HUSTSecLab/criticality_score/cmd/apiserver/internal/tool"
)

var shellTool = tool.Tool{
	ID:          "ad614ff4-66bf-427a-b880-fed96fc380c2",
	Name:        "调试 Shell",
	Description: "该工具可以获得一个 shell 环境，您可以在其中执行任意命令。注意：该工具相当危险，可能会导致数据丢失或泄露。请谨慎使用。",
	Args:        nil,
	Run:         shellImpl,
}

func shellImpl(args map[string]any, in io.Reader, out io.Writer, kill chan int, resize chan tool.ResizeArg) (int, error) {
	return tool.RunExternalCommand([]string{"/bin/bash"}, os.Environ(), in, out, kill, resize)
}

func init() {
	tool.RegistTool(&shellTool)
}
