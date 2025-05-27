package rpc

import "time"

type RoundResp struct {
	CurrentRound int `json:"currentRound"`
}

type GetRoundReq struct {
	RoundID int `json:"roundId"`
}

type TaskStatus string

const (
	TaskStatusPending TaskStatus = "pending"
	TaskStatusRunning TaskStatus = "running"
	TaskStatusSuccess TaskStatus = "success"
	TaskStatusFailed  TaskStatus = "failed"
)

type TaskDTO struct {
	Name         string     `json:"name"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Args         string     `json:"args"`
	Status       TaskStatus `json:"status"`
	Type         string     `json:"type"`
	Dependencies []string   `json:"dependencies"`
	StartTime    *time.Time `json:"startTime"`
	EndTime      *time.Time `json:"endTime"`
}

type RoundDTO struct {
	ID        string    `json:"id"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Tasks     []TaskDTO `json:"tasks"`
}

type StopRunningReq struct {
	Type string `json:"type"`
}

type RpcService interface {
	GetCurrentRoundID(req struct{}, resp *RoundResp) error
	Start(req struct{}, resp *struct{}) error
	Stop(req struct{}, resp *struct{}) error
	GetRound(req GetRoundReq, resp *RoundDTO) error
	StopCurrentRunning(req StopRunningReq, resp *struct{}) error
}
