package admin

import (
	"encoding/binary"
	"net/http"
	"sync"

	"github.com/HUSTSecLab/criticality_score/cmd/apiserver/internal/model"
	"github.com/HUSTSecLab/criticality_score/cmd/apiserver/internal/tool"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// listTools godoc
// @Summary      获取工具列表
// @Description  获取所有可用工具的信息
// @Tags         toolset
// @Produce      json
// @Success      200  {array}   model.ToolDTO
// @Router       /admin/toolset/list [get]
func listTools(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, lo.Map(tool.GetToolList(), func(t *tool.Tool, _ int) model.ToolDTO {
		return *model.ToToolDTO(t)
	}))
}

// createInstance godoc
// @Summary      创建工具实例
// @Description  根据工具ID和参数创建并运行工具实例
// @Tags         toolset
// @Accept       json
// @Produce      json
// @Param        data  body      model.ToolCreateInstanceReq  true  "工具实例创建参数"
// @Success      200   {object}  model.ToolInstanceHistoryDTO
// @Failure      400   {object}  gin.H
// @Failure      500   {object}  gin.H
// @Router       /admin/toolset/instances [post]
func createInstance(ctx *gin.Context) {
	var req model.ToolCreateInstanceReq
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t, err := tool.GetTool(req.ToolID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	username, _, _ := getUser(ctx)
	inst, err := tool.CreateAndRun(t, req.Args, username)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.ToToolInstanceHistoryDTO(
		tool.RunningInstanceToHistory(inst)))
}

type websocketWriter struct {
	conn *websocket.Conn
	mu   *sync.Mutex
	t    byte
}

func (w *websocketWriter) Write(p []byte) (n int, err error) {
	send := []byte{w.t}
	send = append(send, p...)
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.conn.WriteMessage(websocket.BinaryMessage, send); err != nil {
		return 0, err
	}
	return len(p), nil
}

// WebSocket attach
// attachInstance godoc
// @Summary      WebSocket 连接工具实例
// @Description  通过 WebSocket 方式 attach 到指定工具实例
// @Tags         toolset
// @Produce      json
// @Param        id   path      string  true  "实例ID"
// @Success      101  {string}  string  "WebSocket 连接已建立"
// @Failure      400  {object}  gin.H
// @Failure      500  {object}  gin.H
// @Router       /admin/toolset/instances/{id}/attach [get]

// Web socket 报文格式
// Binary Message
// 第一个字节是消息类型
// 0: 输入
// 1：输出
// 2：错误
// 6：程序终止，1-4 字节状态码，后面为可选的错误信息
// 7: resize， 1-4 字节宽度，5-8 字节高度
// 9: 已经结束的实例的输出
func attachInstance(ctx *gin.Context) {
	type P struct {
		ID string `uri:"id" binding:"required"`
	}
	var p P
	if err := ctx.ShouldBindUri(&p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inst, err := tool.GetRunningInstance(p.ID)
	if inst == nil || err != nil {
		if log, err := tool.GetLog(p.ID, false); err == nil {
			// attach to dead instance
			conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			toSend := []byte{9}
			toSend = append(toSend, log...)
			conn.WriteMessage(websocket.BinaryMessage, toSend)
			conn.Close()
			return
		}
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	muWebsocket := &sync.Mutex{}

	outputWriter := &websocketWriter{conn: conn, t: 1, mu: muWebsocket}
	// errWriter := &websocketWriter{conn: conn, t: 2}
	inst.Output.AddWriter(outputWriter)
	// inst.Error.AddWriter(errWriter)
	defer func() {
		inst.Output.RemoveWriter(outputWriter)
		// inst.Error.RemoveWriter(errWriter)
		conn.Close()
	}()

	go func() {
		select {
		case c := <-inst.Result:
			// 发送程序终止消息
			send := []byte{6}
			send = binary.BigEndian.AppendUint32(send, uint32(c.Ret))
			if c.Err != nil {
				send = append(send, []byte(c.Err.Error())...)
			}

			muWebsocket.Lock()
			conn.WriteMessage(websocket.BinaryMessage, send)
			muWebsocket.Unlock()
		case <-ctx.Done():
		}
	}()

	log, _ := tool.GetLog(inst.ID, false)
	outputWriter.Write([]byte(log))

	// wait for disconnect
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			// 关闭连接
			conn.Close()
			return
		}

		if msgType == websocket.BinaryMessage {
			if len(msg) < 1 {
				continue
			}
			msgType := msg[0]
			if msgType == 0 {
				inst.Input.WriteBuffer(msg[1:])
			} else if msgType == 7 {
				select {
				case inst.Resize <- tool.ResizeArg{
					Width:  binary.BigEndian.Uint16(msg[1:3]),
					Height: binary.BigEndian.Uint16(msg[3:5]),
				}:
				default:
					// 如果没有人接收，丢弃
				}
			}
		}
	}
}

// getLog godoc
// @Summary      获取工具实例日志
// @Description  获取指定工具实例的日志
// @Tags         toolset
// @Produce      json
// @Param        id   path      string  true  "实例ID"
// @Param        all  query     bool    false "是否获取所有日志，默认只获取最后1MB"
// @Success      200  {object}  []byte  "日志内容"
// @Failure      400  {object}  gin.H
// @Failure      500  {object}  gin.H
// @Router       /admin/toolset/instances/{id}/log [get]
func getLog(ctx *gin.Context) {
	type P struct {
		ID string `uri:"id" binding:"required"`
	}
	type Q struct {
		all bool `form:"all"`
	}

	var p P
	var q Q
	if err := ctx.ShouldBindUri(&p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ctx.ShouldBindQuery(&q); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log, err := tool.GetLog(p.ID, q.all)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Data(http.StatusOK, "application/octet-stream", log)
}

// getInstances godoc
// @Summary      获取运行中的工具实例列表
// @Description  获取所有运行中的工具实例的信息
// @Param        all  query     bool    false "是否获取所有实例，默认只获取运行中的实例"
// @Param        skip query     int     false "跳过的实例数量，默认0"
// @Param        take query     int     false "获取的实例数量，默认10"
// @Tags         toolset
// @Produce      json
// @Success      200  {object}  model.PageDTO[model.ToolInstanceHistoryDTO]
// @Router       /admin/toolset/instances [get]
func getInstances(ctx *gin.Context) {
	type Q struct {
		All  bool `form:"all"`
		Skip int  `form:"skip"`
		Take int  `form:"take"`
	}

	var q = Q{
		Skip: 0,
		Take: 10,
	}
	if err := ctx.ShouldBindQuery(&q); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var cnt int
	if q.All {
		var err error
		cnt, err = tool.CountToolInstanceHistories()

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		v := tool.GetRunningInstances()
		cnt = len(v)
	}

	instances, err := tool.GetInstanceHistories(!q.All, q.Skip, q.Take)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ret := lo.Map(instances, func(i *tool.ToolInstanceHistory, _ int) model.ToolInstanceHistoryDTO {
		return *model.ToToolInstanceHistoryDTO(i)
	})
	ctx.JSON(http.StatusOK, model.NewPageDTO(cnt, q.Skip, q.Take, ret))
}

// killInstance godoc
// @Summary      杀死工具实例
// @Description  杀死指定工具实例
// @Tags         toolset
// @Param        id   path      string  true  "实例ID"
// @Param        data  body     model.KillToolInstanceReq  true  "工具实例杀死参数"
// @Success      204  {object}  gin.H
// @Failure      400  {object}  gin.H
// @Failure      500  {object}  gin.H
// @Router       /admin/toolset/instances/{id}/kill [post]
func killInstance(ctx *gin.Context) {
	type P struct {
		ID string `uri:"id" binding:"required"`
	}
	var p P
	if err := ctx.ShouldBindUri(&p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inst, err := tool.GetRunningInstance(p.ID)
	if inst == nil || err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var req model.KillToolInstanceReq

	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ok := false

	for _, sig := range inst.Tool.AllowSignals {
		if sig.Value == req.Signal {
			ok = true
			break
		}
	}
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid signal"})
		return
	}

	inst.Kill <- req.Signal
	ctx.JSON(http.StatusNoContent, nil)
}

// getInstance godoc
// @Summary      获取单个工具实例详情
// @Description  获取指定ID的工具实例的详细信息
// @Tags         toolset
// @Produce      json
// @Param        id   path      string  true  "实例ID"
// @Success      200  {object}  model.ToolInstanceHistoryDTO
// @Failure      400  {object}  gin.H
// @Failure      404  {object}  gin.H
// @Router       /admin/toolset/instances/{id} [get]
func getInstance(ctx *gin.Context) {
	type P struct {
		ID string `uri:"id" binding:"required"`
	}
	var p P
	if err := ctx.ShouldBindUri(&p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	history, err := tool.GetInstanceHistory(p.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
		return
	}

	ctx.JSON(http.StatusOK, model.ToToolInstanceHistoryDTO(history))
}

func registToolset(e gin.IRoutes) {
	e.GET("/toolset/list", listTools)
	e.GET("/toolset/instances", getInstances)
	e.POST("/toolset/instances", createInstance)
	e.GET("/toolset/instances/:id", getInstance)
	e.GET("/toolset/instances/:id/attach", attachInstance)
	e.GET("/toolset/instances/:id/log", getLog)
	e.POST("/toolset/instances/:id/kill", killInstance)
}
