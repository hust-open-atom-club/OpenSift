package rpc

import "time"

type RunningTaskDTO struct {
	Link     string    `json:"link"`
	Start    time.Time `json:"start"`
	Progress string    `json:"progress"`
}

type StatusResp struct {
	CurrentTasks []RunningTaskDTO `json:"currentTasks"`
	PendingTasks []string         `json:"pendingTasks"`
	IsRunning    bool             `json:"isRunning"`
}

type RpcService interface {
	Start(req struct{}, resp *struct{}) error
	Stop(req struct{}, resp *struct{}) error
	AddManualTask(req struct {
		GitLink string
	}, resp *struct{}) error
	QueryCurrent(req struct{}, resp *StatusResp) error
}
