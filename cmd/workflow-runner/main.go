package main

import (
	"github.com/HUSTSecLab/criticality_score/cmd/workflow-runner/internal/db"
	"github.com/HUSTSecLab/criticality_score/cmd/workflow-runner/internal/loop"
	"github.com/HUSTSecLab/criticality_score/cmd/workflow-runner/internal/manifest"
	"github.com/HUSTSecLab/criticality_score/cmd/workflow-runner/internal/rpcserver"
	"github.com/HUSTSecLab/criticality_score/cmd/workflow-runner/internal/workflow"
	"github.com/HUSTSecLab/criticality_score/pkg/config"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/spf13/pflag"
)

var handler workflow.RunningHandler

func StopCurrentWorkflow() {
	if handler != nil {
		handler.Stop()
	}
}

func main() {
	config.RegistCommonFlags(pflag.CommandLine)
	config.RegistGitStorageFlags(pflag.CommandLine)
	config.RegistGithubTokenFlags(pflag.CommandLine)
	config.RegistWorkflowRunnerFlags(pflag.CommandLine)
	config.RegistRpcFlags(pflag.CommandLine, false, true)
	config.ParseFlags(pflag.CommandLine)

	db.OpenAndInitDB()
	manifest.InitManifests()

	go func() {
		logger.Info("start rpc server...")
		port, err := config.GetRpcWorkflowPort()
		if err != nil {
			panic(err)
		}
		rpcserver.Start(port)
	}()

	loop.Loop()
}
