package admin

import (
	"slices"

	"github.com/HUSTSecLab/criticality_score/cmd/apiserver/internal/model"
	"github.com/HUSTSecLab/criticality_score/cmd/git-metadata-collector/rpc"
	"github.com/HUSTSecLab/criticality_score/pkg/config"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/HUSTSecLab/criticality_score/pkg/storage/repository"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

// @Summary	Get git files statistics and collector status
// @Router			/admin/gitfiles/status	[get]
// @Success			200	{object}	model.GitFileStatusResp
// @Failure			500	{string}	string
func getGitFilesStatus(c *gin.Context) {
	rpcAddress := config.GetRpcCollectorAddress()
	if rpcAddress == "" {
		c.JSON(500, "rpc address is not set")
		return
	}

	r, err := rpc.NewRpcServiceClient(rpcAddress)

	var collectorRet *rpc.StatusResp = nil
	if err == nil {
		defer r.Close()
		var ret rpc.StatusResp
		collectorRet = &ret
		err := r.QueryCurrent(struct{}{}, &ret)
		if err != nil {
			logger.Warnf("Could not fetch data from rpc server: %v", err)
			collectorRet = nil
		}
	} else {
		logger.Warnf("Could not connect to rpc: %v", err)
	}

	repo := repository.NewGitMetricsRepository(storage.GetDefaultAppDatabaseContext())
	gfs, err := repo.GetGitFilesStatistics()
	if err != nil {
		c.JSON(500, "Could not fetch git files from database")
		return
	}

	c.JSON(200, model.GitFileStatusResp{
		GitFile:   model.GitFileStatisticsResultDOToDTO(gfs),
		Collector: collectorRet,
	})
}

// @Summary			Get Git File List
// @Router			/admin/gitfiles	[get]
// @Param			link	query	string	false "Git link"
// @Param           filter  query   integer false "Filter, 0: no filter, 1: success, 2: fail, 3: never success"
// @Param			skip	query	integer	false "Skip count"
// @Param			take	query	integer	false "Take count"
// @Success			200	{object} model.PageDTO[model.GitFileDTO]
func getGitFilesList(c *gin.Context) {
	type query struct {
		Link   string `form:"link"`
		Filter int    `form:"filter"`
		Skip   int    `form:"skip"`
		Take   int    `form:"take"`
	}

	var q query = query{
		Link:   "",
		Filter: 0,
		Skip:   0,
		Take:   100,
	}
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(400, "Invalid query parameters")
		return
	}

	r := repository.NewGitMetricsRepository(storage.GetDefaultAppDatabaseContext())
	ret, cnt, err := r.QueryGitFiles(q.Link, q.Filter, q.Skip, q.Take)
	if err != nil {
		c.JSON(500, "fetch database error")
		return
	}
	items := lo.Map(slices.Collect(ret), func(i *repository.GitFile, _ int) *model.GitFileDTO {
		return model.GitFileDOToDTO(i)
	})

	c.JSON(200, model.NewPageDTO(cnt, q.Skip, q.Take, items))
}

// @Summary			Append to collector manual list
// @Router			/admin/gitfiles/manual	[post]
// @Success			204 {string} string
// @Failure			400 {string} string
// @Failure			500 {string} string
// @Accept			json
// @Param           req  body    model.GitFileAppendManualReq  true "Append manual request"
func appendGitFilesManualList(c *gin.Context) {
	rpcAddress := config.GetRpcCollectorAddress()
	if rpcAddress == "" {
		c.JSON(500, "rpc address is not set")
		return
	}

	r, err := rpc.NewRpcServiceClient(rpcAddress)
	if err != nil {
		c.JSON(500, "")
		return
	}

	defer r.Close()

	var req model.GitFileAppendManualReq

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(400, "git link is not valid")
	}

	err = r.AddManualTask(struct{ GitLink string }{
		GitLink: req.GitLink,
	}, &struct{}{})
	if err != nil {
		c.JSON(500, "rpc call failed: "+err.Error())
		return
	}
	c.Status(204)
}

// @Summary      Start Git File Collector
// @Description  Starts the Git File collection process
// @Router       /admin/gitfiles/start [post]
// @Success      204 {string} string "No Content"
// @Failure      500 {string} string "Internal Server Error"
func startGitFileCollector(c *gin.Context) {
	rpcAddress := config.GetRpcCollectorAddress()
	if rpcAddress == "" {
		c.JSON(500, "rpc address is not set")
		return
	}

	r, err := rpc.NewRpcServiceClient(rpcAddress)
	if err != nil {
		c.JSON(500, "")
		return
	}

	defer r.Close()

	r.Start(struct{}{}, nil)
	c.JSON(204, "")
}

// @Summary      Stop Git File Collector
// @Description  Stops the Git File collection process
// @Router       /admin/gitfiles/stop [post]
// @Success      204 {string} string "No Content"
// @Failure      500 {string} string "Internal Server Error"
func stopGitFileCollector(c *gin.Context) {
	rpcAddress := config.GetRpcCollectorAddress()
	if rpcAddress == "" {
		c.JSON(500, "rpc address is not set")
		return
	}
	r, err := rpc.NewRpcServiceClient(rpcAddress)
	if err != nil {
		c.JSON(500, "")
		return
	}

	defer r.Close()

	r.Stop(struct{}{}, nil)
	c.JSON(204, "")
}

func registGitFile(r gin.IRoutes) {
	r.GET("/gitfiles", getGitFilesList)
	r.GET("/gitfiles/status", getGitFilesStatus)
	r.POST("/gitfiles/manual", appendGitFilesManualList)
	r.POST("/gitfiles/start", startGitFileCollector)
	r.POST("/gitfiles/stop", stopGitFileCollector)
	// TODO: Delete Repo, update stastics only
}
