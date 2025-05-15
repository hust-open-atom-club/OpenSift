package rpc

import "github.com/HUSTSecLab/criticality_score/cmd/git-metadata-collector/internal/task"

type StatusResp struct {
	CurrentTasks []task.RunningTask `json:"currentTasks"`
	PendingTasks []string           `json:"pendingTasks"`
	IsRunning    bool               `json:"isRunning"`
}

type RpcService interface {
	Start(req struct{}, resp *struct{}) error
	Stop(req struct{}, resp *struct{}) error
	AddManualTask(req struct {
		GitLink string
	}, resp *struct{}) error
	QueryCurrent(req struct{}, resp *StatusResp) error
}
