package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/HUSTSecLab/OpenSift/cmd/workflow-runner/internal/db"
	"github.com/HUSTSecLab/OpenSift/cmd/workflow-runner/internal/workflow"
	"github.com/HUSTSecLab/OpenSift/cmd/workflow-runner/rpc"
	"github.com/HUSTSecLab/OpenSift/pkg/logger"
	"github.com/samber/lo"
)

func WorkflowRunExec(ctx workflow.RunningCtx, args []string, stop chan struct{}, kill chan struct{}) error {
	if len(args) == 0 {
		return nil // No command to run
	}

	finish := make(chan struct{})
	defer close(finish)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = ctx.LoggerFile
	cmd.Stderr = ctx.LoggerFile
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run cmd fail: %v", err)
	}

	// if stop is received, or stop, kill the process
	go func() {
		select {
		case <-stop:
			logger.Infof("stopping process %s (%s)", ctx.Node.Name, ctx.Node.Title)
			err := cmd.Process.Signal(os.Interrupt)
			if err != nil {
				logger.Errorf("failed to stop process %s: %v", ctx.Node.Name, err)
			}
		case <-kill:
			logger.Infof("killing process %s", ctx.Node.Name)
			err := cmd.Process.Kill()
			if err != nil {
				logger.Errorf("failed to kill process %s: %v", ctx.Node.Name, err)
			}
		case <-finish:
		}
	}()

	return nil
}

func WorkflowRunExecWrapper(args []string) workflow.WorkflowRunFunc {
	return func(ctx *workflow.RunningCtx, stop chan struct{}, kill chan struct{}) error {
		return WorkflowRunExec(*ctx, args, stop, kill)
	}
}

func WorkflowRunBefore(ctx *workflow.RunningCtx) error {
	var argsStr string
	if ctx.Args != nil {
		argsJSON, err := json.Marshal(ctx.Args)
		if err != nil {
			return fmt.Errorf("failed to marshal args: %v", err)
		}
		argsStr = string(argsJSON)
	}

	err := db.UpdateTask(ctx.RoundID, &rpc.TaskDTO{
		Name:      ctx.Node.Name,
		Args:      lo.ToPtr(argsStr),
		Status:    rpc.TaskStatusRunning,
		StartTime: lo.ToPtr(time.Now()),
	})

	return err

}

func WorkflowRunAfter(ctx *workflow.RunningCtx, result error) error {
	var status rpc.TaskStatus = rpc.TaskStatusSuccess
	if result != nil {
		status = rpc.TaskStatusFailed
	}

	t, err := db.GetTask(ctx.RoundID, ctx.Node.Name)
	if err != nil {
		return fmt.Errorf("failed to get task before update: %v", err)
	}
	t.Status = status
	t.EndTime = lo.ToPtr(time.Now())

	err = db.UpdateTask(ctx.RoundID, t)

	if err != nil {
		return fmt.Errorf("failed to update task after run: %v", err)
	}
	logger.Infof("task %s (%s) finished with status: %s", ctx.Node.Name, ctx.Node.Title, status)

	return nil
}

func NeedUpdateWrapper(n *workflow.WorkflowNode, interval time.Duration) func() bool {
	return func() bool {
		// get last update time from db
		lastUpdateTime, _, _ := db.GetLastTriggerTime(n.Name)

		if time.Since(lastUpdateTime) > interval {
			return true
		}
		return false
	}

}
