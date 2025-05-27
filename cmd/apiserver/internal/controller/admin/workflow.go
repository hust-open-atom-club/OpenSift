package admin

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/HUSTSecLab/criticality_score/cmd/apiserver/internal/model"
	"github.com/HUSTSecLab/criticality_score/cmd/workflow-runner/rpc"
	"github.com/HUSTSecLab/criticality_score/pkg/config"
	"github.com/gin-gonic/gin"
)

// getMaxWorkflowID godoc
// @Summary      获取最大 workflow 轮次 ID
// @Description  获取当前最大的 workflow 轮次 ID
// @Tags         workflow
// @Produce      json
// @Success      200  {object}  rpc.RoundResp
// @Failure      500  {object}  string
// @Router       /admin/workflows/maxRounds [get]
func getMaxWorkflowID(c *gin.Context) {
	rpcAddress := config.GetRpcWorkflowAddress()
	if rpcAddress == "" {
		c.JSON(500, "rpc address is not set")
		return
	}

	client, err := rpc.NewRpcServiceClient(rpcAddress)
	if err != nil {
		c.JSON(500, "Could not connect to rpc server: "+err.Error())
		return
	}
	defer client.Close()

	var resp rpc.RoundResp

	err = client.GetCurrentRoundID(struct{}{}, &resp)
	if err != nil {
		c.JSON(500, "Failed to get max workflow ID: "+err.Error())
		return
	}
	c.JSON(200, &resp)
}

// getWorkflowByID godoc
// @Summary      获取指定 workflow 轮次详情
// @Description  根据轮次 ID 获取 workflow 详情
// @Tags         workflow
// @Produce      json
// @Param        id   path      int  true  "轮次ID"
// @Success      200  {object}  rpc.RoundDTO
// @Failure      400  {object}  string
// @Failure      500  {object}  string
// @Router       /admin/workflows/rounds/{id} [get]
func getWorkflowByID(c *gin.Context) {
	type P struct {
		ID int `uri:"id" binding:"required"`
	}
	var p P

	if err := c.ShouldBindUri(&p); err != nil {
		c.JSON(400, "Invalid request: "+err.Error())
		return
	}

	rpcAddress := config.GetRpcWorkflowAddress()
	if rpcAddress == "" {
		c.JSON(500, "rpc address is not set")
		return
	}

	client, err := rpc.NewRpcServiceClient(rpcAddress)
	if err != nil {
		c.JSON(500, "Could not connect to rpc server: "+err.Error())
		return
	}
	defer client.Close()

	var resp rpc.RoundDTO
	err = client.GetRound(rpc.GetRoundReq{RoundID: p.ID}, &resp)
	if err != nil {
		c.JSON(500, "Failed to get workflow by ID: "+err.Error())
		return
	}
	c.JSON(200, &resp)
}

// getWorkflowLogs godoc
// @Summary      获取 workflow 日志
// @Description  获取指定轮次和名称的 workflow 日志文件
// @Tags         workflow
// @Produce      octet-stream
// @Param        id    path      int     true  "轮次ID"
// @Param        name  path      string  true  "日志名称"
// @Success      200   {file}    file
// @Failure      400   {object}  string
// @Failure      500   {object}  string
// @Router       /admin/workflows/{id}/logs/{name} [get]
func getWorkflowLogs(c *gin.Context) {
	type P struct {
		ID   int    `uri:"id" binding:"required"`
		Name string `uri:"name" binding:"required"`
	}
	var p P
	if err := c.ShouldBindUri(&p); err != nil {
		c.JSON(400, "Invalid request: "+err.Error())
		return
	}
	dir := config.GetWebToolHistoryDir()
	filename, err := filepath.Abs(filepath.Join(dir, fmt.Sprintf("round_%d", p.ID), p.Name+".log"))
	if err != nil {
		c.JSON(500, "Failed to get absolute path: "+err.Error())
		return
	}
	// check if exceed dir boundary
	if !strings.HasPrefix(filename, dir) {
		c.JSON(400, "Invalid request: log file is not in the history directory")
		return
	}

	c.File(filename)
}

// updateWorkflowStatus godoc
// @Summary      启动或停止 workflow
// @Description  启动或停止 workflow 运行状态
// @Tags         workflow
// @Accept       json
// @Produce      json
// @Param        data  body      model.UpdateWorkflowStatusReq  true  "workflow 状态参数"
// @Success      204   {object}  nil
// @Failure      400   {object}  string
// @Failure      500   {object}  string
// @Router       /admin/workflows/status [post]
func updateWorkflowStatus(c *gin.Context) {
	var req model.UpdateWorkflowStatusReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, "Invalid request: "+err.Error())
		return
	}

	rpcAddress := config.GetRpcWorkflowAddress()
	if rpcAddress == "" {
		c.JSON(500, "rpc address is not set")
	}
	client, err := rpc.NewRpcServiceClient(rpcAddress)
	if err != nil {
		c.JSON(500, "Could not connect to rpc server: "+err.Error())
	}
	defer client.Close()

	if req.Running {
		err = client.Start(struct{}{}, &struct{}{})
		if err != nil {
			c.JSON(500, "Failed to start workflow: "+err.Error())
			return
		}
	} else {
		err = client.Stop(struct{}{}, &struct{}{})
		if err != nil {
			c.JSON(500, "Failed to stop workflow: "+err.Error())
			return
		}
	}
	c.Status(204) // No Content
}

// killWorkflowJob godoc
// @Summary      杀死 workflow 任务
// @Description  杀死当前运行中的 workflow 任务
// @Tags         workflow
// @Accept       json
// @Produce      json
// @Param        data  body      model.KillWorkflowJobReq  true  "kill 参数"
// @Success      204   {object}  nil
// @Failure      400   {object}  string
// @Failure      500   {object}  string
// @Router       /admin/workflows/kill [post]
func killWorkflowJob(c *gin.Context) {
	var req model.KillWorkflowJobReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, "Invalid request: "+err.Error())
		return
	}

	rpcAddress := config.GetRpcWorkflowAddress()
	if rpcAddress == "" {
		c.JSON(500, "rpc address is not set")
		return
	}

	client, err := rpc.NewRpcServiceClient(rpcAddress)
	if err != nil {
		c.JSON(500, "Could not connect to rpc server: "+err.Error())
		return
	}
	defer client.Close()

	err = client.StopCurrentRunning(rpc.StopRunningReq{Type: req.Type}, &struct{}{})
	if err != nil {
		c.JSON(500, "Failed to kill workflow job: "+err.Error())
		return
	}
	c.Status(204) // No Content
}

func registWorkflow(g gin.IRoutes) {
	g.GET("/workflows/maxRounds", getMaxWorkflowID)
	// g.GET("/workflows/next", getNextWorkflow)
	g.GET("/workflows/rounds/:id", getWorkflowByID)
	g.GET("/workflows/:id/logs/:name", getWorkflowLogs)

	g.POST("/workflows/status", updateWorkflowStatus)
	g.POST("/workflows/kill", killWorkflowJob)
}
