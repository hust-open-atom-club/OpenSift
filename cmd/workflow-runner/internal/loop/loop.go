package loop

import (
	"sync"
	"time"

	"github.com/HUSTSecLab/criticality_score/cmd/workflow-runner/internal/db"
	"github.com/HUSTSecLab/criticality_score/cmd/workflow-runner/internal/manifest"
	"github.com/HUSTSecLab/criticality_score/cmd/workflow-runner/internal/workflow"
	"github.com/HUSTSecLab/criticality_score/pkg/config"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
)

var handler workflow.RunningHandler

var running bool = true
var muRunning sync.Mutex
var condRunning = sync.NewCond(&muRunning)

func SetRunningState(state bool) {
	muRunning.Lock()
	defer muRunning.Unlock()
	running = state
	if state {
		logger.Info("workflow runner is now running")
		condRunning.Broadcast() // Notify all waiting goroutines to wake up
	} else {
		logger.Info("workflow runner is paused")
	}
}

func StopCurrentJob(kill bool) {
	if handler == nil {
		logger.Warn("no workflow is currently running")
		return
	}

	var err error
	if kill {
		err = handler.Kill()
	} else {
		err = handler.Stop()
	}

	if err != nil {
		logger.Errorf("failed to stop current workflow: %v", err)
	} else {
		logger.Info("current workflow stopped successfully")
	}
}

func Loop() {
	var err error
	var rnd int
	target := manifest.GetTargetTask()
	if target == nil {
		panic("target task is nil")
	}

	for {
		<-time.After(10 * time.Second) // Wait for 10 seconds before starting the next round
		for !running {
			muRunning.Lock()
			condRunning.Wait()
			muRunning.Unlock()
		}
		var allUpToDate bool
		allUpToDate, err = target.AllUpToDate()

		if err != nil {
			logger.Errorf("failed to check if all tasks are up to date: %v", err)
			continue
		}

		if allUpToDate {
			continue // Skip if all tasks are up to date
		}

		rnd, err = db.CreateRound(manifest.GetAllTasks())
		if err != nil {
			logger.Errorf("failed to create round: %v", err)
			continue
		}

		handler, err = target.StartWorkflow(&workflow.WorkflowStartOption{
			OutputDir:         config.GetWorkflowHistoryDir(),
			ArgsGetter:        nil, // TODO: args
			RoundID:           rnd,
			NeedUpdateDefault: true,
		})

		if err != nil {
			logger.Errorf("failed to start workflow: %v", err)
			continue
		}

		logger.Infof("workflow started with round ID: %d", rnd)

		err = handler.Wait()
		if err != nil {
			logger.Errorf("workflow failed: %v", err)
		} else {
			logger.Info("workflow completed successfully")
		}
	}

}
