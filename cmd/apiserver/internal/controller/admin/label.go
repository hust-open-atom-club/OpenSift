package admin

import (
	"slices"

	"github.com/HUSTSecLab/criticality_score/cmd/apiserver/internal/model"
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

func registLabel(g gin.IRoutes) {
	g.GET("/label/distributions/all", getDistributionPackagesPrefixes)
	g.PUT("/label/distributions/gitlink", updateDistributionGitLink)
	g.GET("/label/distributions", getDistributionPackages)
}
