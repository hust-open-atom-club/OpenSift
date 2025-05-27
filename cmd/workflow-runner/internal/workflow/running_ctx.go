package workflow

import (
	"fmt"
	"os"
	"sync"
)

type RunningHandler interface {
	CurrentRunning() []*WorkflowNode
	Wait() error
	Stop() error
	Kill() error
}

type runningHandler struct {
	finish           chan error
	currentRunning   []*RunningCtx
	muCurrentRunning sync.Mutex
	stop             chan struct{}
	kill             chan struct{}
}

var _ RunningHandler = (*runningHandler)(nil)

func newRunningHandler() *runningHandler {
	return &runningHandler{
		finish: make(chan error),
		stop:   make(chan struct{}),
		kill:   make(chan struct{}),
	}
}

func (h *runningHandler) Stop() error {
	h.muCurrentRunning.Lock()
	if len(h.currentRunning) == 0 {
		h.muCurrentRunning.Unlock()
		return fmt.Errorf("no running process")
	}
	h.muCurrentRunning.Unlock()

	h.stop <- struct{}{}
	return nil
}

func (h *runningHandler) Kill() error {
	h.muCurrentRunning.Lock()
	if len(h.currentRunning) == 0 {
		h.muCurrentRunning.Unlock()
		return fmt.Errorf("no running process")
	}
	h.muCurrentRunning.Unlock()

	h.kill <- struct{}{}
	return nil
}

func (h *runningHandler) Wait() error {
	return <-h.finish
}

func (h *runningHandler) CurrentRunning() []*WorkflowNode {
	h.muCurrentRunning.Lock()
	defer h.muCurrentRunning.Unlock()
	nodes := make([]*WorkflowNode, 0, len(h.currentRunning))
	for _, ctx := range h.currentRunning {
		nodes = append(nodes, ctx.Node)
	}
	return nodes
}

type RunningCtx struct {
	RoundID    int
	Args       any
	LoggerFile *os.File
	Node       *WorkflowNode

	runningHandler *runningHandler
}

func (ctx *RunningCtx) Run() error {
	n := ctx.Node
	needUpdate := true

	if n.NeedUpdate != nil && !n.NeedUpdate() {
		needUpdate = false
	}

	ctx.runningHandler.muCurrentRunning.Lock()
	ctx.runningHandler.currentRunning = append(ctx.runningHandler.currentRunning, ctx)
	ctx.runningHandler.muCurrentRunning.Unlock()

	defer func() {
		ctx.runningHandler.muCurrentRunning.Lock()
		for i, running := range ctx.runningHandler.currentRunning {
			if running.Node == n {
				ctx.runningHandler.currentRunning = append(ctx.runningHandler.currentRunning[:i], ctx.runningHandler.currentRunning[i+1:]...)
				break
			}
		}
		ctx.runningHandler.muCurrentRunning.Unlock()
	}()

	var runErr error

	if needUpdate {
		if n.RunBefore != nil {
			if err := n.RunBefore(ctx); err != nil {
				return err
			}
		}

		if n.Run != nil {
			if err := n.Run(ctx, ctx.runningHandler.stop, ctx.runningHandler.kill); err != nil {
				runErr = err
			}
		}

		if n.RunAfter != nil {
			if err := n.RunAfter(ctx, runErr); err != nil {
				return err
			}
		}
	}
	return runErr
}
