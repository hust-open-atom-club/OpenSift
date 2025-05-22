# 工具开发指南

## 工具的基本结构

每个工具都需要实现一个 `tool.Tool` 结构体，并通过 `tool.RegistTool` 注册。主要字段如下：

- `ID`：工具唯一标识符（UUID）。
- `Name`：工具名称。
- `Description`：工具描述。
- `Group`：工具分组。
- `Args`：参数列表（可选）。
- `AllowSignals`：允许的信号类型（可选）。
- `Run`：工具的执行函数。


## 常见工具开发

常见工具分为两类：外部调用型工具和内部调用型工具。
- **外部调用型工具**：通过调用系统命令实现，适合需要执行 shell 命令或其他外部程序的场景。
- **内部调用型工具**：直接用 Go 代码实现业务逻辑，适合纯 Go 逻辑的场景。

### 外部调用型工具

外部调用型工具通过调用系统命令（如 shell、python 等）实现。推荐使用 `tool.RunExternalCommand` 作为 `Run` 实现。

#### 步骤

1. **定义 Tool 结构体**

   ```go
   var shellTool = tool.Tool{
       ID:           "ad614ff4-66bf-427a-b880-fed96fc380c2",
       Name:         "调试 Shell",
       Description:  "该工具可以获得一个 shell 环境，您可以在其中执行任意命令。",
       Group:        "调试工具",
       AllowSignals: tool.ExternalCommandToolSignals,
       Args:         nil,
       Run:          shellImpl,
   }
   ```

2. **实现 Run 函数**

   推荐直接调用 `tool.RunExternalCommand`，传入命令、环境变量、输入输出流等参数。例如：

   ```go
   func shellImpl(args map[string]any, in io.Reader, out io.Writer, kill chan int, resize chan tool.ResizeArg) (int, error) {
       return tool.RunExternalCommand([]string{"/bin/bash"}, os.Environ(), in, out, kill, resize)
   }
   ```

3. **注册工具**

   ```go
   func init() {
       tool.RegistTool(&shellTool)
   }
   ```


### 内部调用型工具

内部调用型工具直接用 Go 代码实现业务逻辑。推荐使用 `tool.CanioalizeWrapper` 包装你的实现函数。

#### 步骤

1. **定义 Tool 结构体**

   ```go
   var testTool = &tool.Tool{
       ID:          "862910f9-8ce0-4cd9-a26c-86f84a783967",
       Name:        "测试工具",
       Description: "该工具用于测试功能是否正常。",
       Group:       "示例工具",
       Args: []tool.ToolArg{
           {Name: "arg1", Type: tool.ToolArgTypeString, Description: "arg 1 desc", Default: "default"},
           // ...更多参数...
       },
       Run: tool.CanioalizeWrapper(testImpl),
   }
   ```

2. **实现业务逻辑函数**

   你的实现函数签名通常为：

   ```go
   func testImpl(args map[string]any, in io.Reader, out io.Writer, err io.Writer, kill chan int) error {
       // 业务逻辑
   }
   ```

   例如：

   ```go
   func testImpl(args map[string]any, in io.Reader, out io.Writer, err io.Writer, kill chan int) error {
       fmt.Fprintf(out, "test tool args: %v\n", args)
       // ...更多逻辑...
       return nil
   }
   ```

3. **注册工具**

   ```go
   func init() {
       tool.RegistTool(testTool)
   }
   ```

## 如何手动实现 Run 函数

虽然推荐使用 `tool.RunExternalCommand` 或 `tool.CanioalizeWrapper`，但你也可以手动实现 `Run` 函数，只需保证签名和参数约定一致。例如：

```go
func myCustomRun(args map[string]any, in io.Reader, out io.Writer, kill chan int, resize chan tool.ResizeArg) (int, error) {
    // 你可以完全自定义命令执行、输入输出处理、信号响应等
    // 返回值为 exit code 和 error
}
```

注意，其中 in 是来自终端的输入流，out 是发往终端的输出流，这可能需要你处理终端的控制序列和输入输出格式，kill 是一个信号通道，用于接收终止信号，resize 是一个通道，用于处理终端大小变化。

## 关于终止信号

每个工具首先需要定义允许的终止信号，外部只可能向 kill 通道发送允许的终止信号，工具应该在任何可以终止的地方检查 kill 通道，收到信号后应该立即返回。

可以采用以下方式检查 kill 通道：

```go
select {
    case sig := <-kill:
        // 处理终止信号
        ...
    default:
        // 正常执行逻辑
        ...
}
```

或者调用 `tool.CheckKill` 函数来检查 kill 通道。

## 注意事项

如有疑问，建议参考 `cmd/apiserver/internal/toolimpl/` 下的其他工具实现，或查看 `tool` 包的文档和示例代码。