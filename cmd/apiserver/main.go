package main

import (
	"fmt"

	"github.com/HUSTSecLab/criticality_score/cmd/apiserver/docs"
	"github.com/HUSTSecLab/criticality_score/cmd/apiserver/internal/controller"
	"github.com/HUSTSecLab/criticality_score/cmd/apiserver/internal/server"
	"github.com/HUSTSecLab/criticality_score/cmd/apiserver/internal/tool"

	// regist all tool implementations
	_ "github.com/HUSTSecLab/criticality_score/cmd/apiserver/internal/toolimpl"
	"github.com/HUSTSecLab/criticality_score/pkg/config"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/spf13/pflag"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	var (
		flagHost = pflag.StringP("host", "H", "0.0.0.0", "apiserver host")
		flagPort = pflag.IntP("port", "p", 5000, "apiserver port")
	)

	config.RegistCommonFlags(pflag.CommandLine)
	config.RegistRpcFlags(pflag.CommandLine, true, true)
	config.ParseFlags(pflag.CommandLine)

	logger.SetContext("apiserver")
	logger.Info("Start apiserver...")

	// initialize tool
	tool.OpenAndInitDB()
	defer tool.CloseDB()

	s := server.NewServer()
	apiGroup := s.Group("/api/v1")
	controller.Regist(apiGroup)

	docs.SwaggerInfo.BasePath = "/api/v1"
	s.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	s.Run(fmt.Sprintf("%s:%d", *flagHost, *flagPort))
}
