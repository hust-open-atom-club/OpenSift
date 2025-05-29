package admin

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/HUSTSecLab/criticality_score/cmd/apiserver/internal/model"
	"github.com/HUSTSecLab/criticality_score/pkg/llm"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/HUSTSecLab/criticality_score/pkg/storage/repository"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

var allowedTables = []repository.DistPackageTablePrefix{
	repository.DistLinkTablePrefixAlpine,
	repository.DistLinkTablePrefixArchlinux,
	repository.DistLinkTablePrefixAur,
	repository.DistLinkTablePrefixCentos,
	repository.DistLinkTablePrefixDebian,
	repository.DistLinkTablePrefixDeepin,
	repository.DistLinkTablePrefixFedora,
	repository.DistLinkTablePrefixGentoo,
	repository.DistLinkTablePrefixHomebrew,
	repository.DistLinkTablePrefixNix,
	repository.DistLinkTablePrefixUbuntu,
}

// updateDistributionGitLink godoc
// @Summary      更新发行版包的 Git 链接
// @Description  更新指定发行版包的 Git 仓库链接和置信度
// @Tags         label
// @Accept       json
// @Produce      json
// @Param        data  body      model.UpdateDistributionGitLinkReq  true  "Git 链接参数"
// @Success      204   {object}  nil
// @Failure      400   {object}  string
// @Failure      500   {object}  string
// @Router       /admin/label/distributions/gitlink [put]
func updateDistributionGitLink(c *gin.Context) {
	var req model.UpdateDistributionGitLinkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	if req.Distribution == "" {
		c.JSON(400, gin.H{"error": "Distribution is required"})
		return
	}
	if lo.IndexOf(allowedTables, repository.DistPackageTablePrefix(req.Distribution)) == -1 {
		c.JSON(400, gin.H{"error": "Invalid distribution: " + req.Distribution})
		return
	}

	repo := repository.NewDistPackageRepository(storage.GetDefaultAppDatabaseContext(), repository.DistPackageTablePrefix(req.Distribution))

	if err := repo.UpdateGitLink(req.PackageName, req.Link, req.Confidence); err != nil {
		c.JSON(500, gin.H{"error": "Failed to update git link: " + err.Error()})
		return
	}
	c.Status(204) // No Content
}

// getDistributionPackages godoc
// @Summary      查询发行版包列表
// @Description  根据发行版、链接、置信度等条件分页查询包列表
// @Tags         label
// @Produce      json
// @Param        distribution     query     string  true   "发行版名称"
// @Param        skip             query     int     false  "跳过数量"
// @Param        take             query     int     false  "返回数量"
// @Param        search           query     string  false  "链接过滤"
// @Param        confidence       query     int     false  "置信度过滤（0:全部, 1:已标注, 2:未标注）"
// @Success      200  {object}  model.PageDTO[model.DistributionPackageDTO]
// @Failure      400  {object}  string
// @Failure      500  {object}  string
// @Router       /admin/label/distributions [get]
func getDistributionPackages(c *gin.Context) {
	type Q struct {
		Skip             int    `form:"skip"`
		Take             int    `form:"take"`
		Distribution     string `form:"distribution" required:"true"`
		LinkFilter       string `form:"search"`
		ConfidenceFilter int    `form:"confidence"` // 0: all, 1: set to 1, 2: not set to 1
	}
	var q = Q{
		Skip: 0,
		Take: 100,
	}
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(400, gin.H{"error": "Invalid query parameters: " + err.Error()})
		return
	}

	if q.Distribution == "" {
		c.JSON(400, gin.H{"error": "Distribution is required"})
		return
	}
	if lo.IndexOf(allowedTables, repository.DistPackageTablePrefix(q.Distribution)) == -1 {
		c.JSON(400, gin.H{"error": "Invalid distribution: " + q.Distribution})
	}

	repo := repository.NewDistPackageRepository(storage.GetDefaultAppDatabaseContext(), repository.DistPackageTablePrefix(q.Distribution))

	items, cnt, err := repo.QueryWithFilter(q.ConfidenceFilter, q.LinkFilter, q.Skip, q.Take)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to query distribution packages: " + err.Error()})
		return
	}
	packages := lo.Map(slices.Collect(items), func(i *repository.DistPackage, _ int) *model.DistributionPackageDTO {
		return model.ToDistributionPackageDTO(i)
	})

	c.JSON(200, model.NewPageDTO(cnt, q.Skip, q.Take, packages))
}

// getDistributionPackagesPrefixes godoc
// @Summary      获取所有发行版包的前缀
// @Description  获取所有支持的发行版包前缀列表
// @Tags         label
// @Produce      json
// @Success      200  {object}  []string
// @Failure      500  {object}  string
// @Router       /admin/label/distributions/all [get]
func getDistributionPackagesPrefixes(c *gin.Context) {
	prefixes := lo.Map(allowedTables, func(i repository.DistPackageTablePrefix, _ int) string {
		return string(i)
	})

	c.JSON(200, prefixes)
}

// getDistributionAICompletion godoc
// @Summary      AI 补全发行版包 Git 链接
// @Description  使用 AI 补全指定发行版包的 Git 仓库链接，返回 JSON 流
// @Tags         label
// @Accept       json
// @Produce      json
// @Param        data  body      model.GitLinkAICompletionReq  true  "AI 补全参数"
// @Success      200   {object}  object  "JSON 流，每行为一个结果"
// @Failure      400   {object}  string
// @Failure      500   {object}  string
// @Router       /admin/label/distributions/ai-completion [post]
func getDistributionAICompletion(c *gin.Context) {
	var req model.GitLinkAICompletionReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	if req.Distribution == "" || req.PackageName == "" {
		c.JSON(400, gin.H{"error": "Distribution and package name are required"})
		return
	}

	if lo.IndexOf(allowedTables, repository.DistPackageTablePrefix(req.Distribution)) == -1 {
		c.JSON(400, gin.H{"error": "Invalid distribution: " + req.Distribution})
		return
	}

	res, err := llm.AskGitLinkPrompt(req.Distribution, req.PackageName, req.Description, req.HomePage)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get AI completion: " + err.Error()})
		return
	}

	// Disable gzipping for SSE
	c.Writer.Header().Del("Content-Encoding")

	// et proper headers for SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering if using nginx
	c.Status(200)

	// Ensure the client buffer is flushed immediately
	c.Writer.Flush()

	// Set client-gone detection
	clientGone := c.Writer.CloseNotify()

	for r, err := range res {
		// Check if client disconnected
		select {
		case <-clientGone:
			return
		default:
		}

		if err != nil {
			msg, _ := json.Marshal(gin.H{"error": "Failed to get AI completion: " + err.Error()})
			fmt.Fprintf(c.Writer, "data: %s\n\n", string(msg))
			c.Writer.Flush()
			continue
		}

		p, _ := r.MarshalJSON()
		fmt.Fprintf(c.Writer, "data: %s\n\n", string(p))
		c.Writer.Flush()
	}
}

func registLabel(g gin.IRoutes) {
	g.GET("/label/distributions/all", getDistributionPackagesPrefixes)
	g.PUT("/label/distributions/gitlink", updateDistributionGitLink)
	g.GET("/label/distributions", getDistributionPackages)
	g.POST("/label/distributions/ai-completion", getDistributionAICompletion)
}
