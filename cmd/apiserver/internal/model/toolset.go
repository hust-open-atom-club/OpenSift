package model

import (
	"time"

	"github.com/HUSTSecLab/OpenSift/cmd/apiserver/internal/tool"
	"github.com/samber/lo"
)

type ToolSignalDTO struct {
	Value       int    `json:"value"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ToolDTO struct {
	// ID is the unique identifier for the toolset.
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Group        string          `json:"group"`
	Args         []ToolArgDTO    `json:"args"`
	AllowSignals []ToolSignalDTO `json:"allowedSignals"`
}

type ToolArgDTO struct {
	// Name is the name of the argument.
	Name        string `json:"name"`
	Type        string `json:"type"`
	Default     any    `json:"default"`
	Description string `json:"description"`
}

type ToolCreateInstanceReq struct {
	ToolID string         `json:"toolId"`
	Args   map[string]any `json:"args"`
}

type ToolInstanceHistoryDTO struct {
	ID             string     `json:"id"`
	ToolID         string     `json:"toolId"`
	ToolName       string     `json:"toolName"`
	Tool           *ToolDTO   `json:"tool"`
	LaunchUserName string     `json:"launchUserName"`
	StartTime      *time.Time `json:"startTime"`
	EndTime        *time.Time `json:"endTime"`
	Ret            *int       `json:"ret"`
	Err            *string    `json:"err"`
	IsRunning      bool       `json:"isRunning"`
}

func ToToolArgDTO(arg *tool.ToolArg) *ToolArgDTO {
	return &ToolArgDTO{
		Name:        arg.Name,
		Type:        string(arg.Type),
		Default:     arg.Default,
		Description: arg.Description,
	}
}

func ToToolSignalDTO(signal *tool.ToolSignal) *ToolSignalDTO {
	return &ToolSignalDTO{
		Value:       signal.Value,
		Name:        signal.Name,
		Description: signal.Description,
	}
}

func ToToolDTO(t *tool.Tool) *ToolDTO {
	tDTO := &ToolDTO{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		Group:       t.Group,
	}
	tDTO.Args = lo.Map(t.Args, func(arg tool.ToolArg, _ int) ToolArgDTO {
		return *ToToolArgDTO(&arg)
	})
	tDTO.AllowSignals = lo.Map(t.AllowSignals, func(signal tool.ToolSignal, _ int) ToolSignalDTO {
		return *ToToolSignalDTO(&signal)
	})

	return tDTO
}

func ToToolInstanceHistoryDTO(inst *tool.ToolInstanceHistory) *ToolInstanceHistoryDTO {
	tool, _ := tool.GetTool(inst.ToolID)
	t := ToToolDTO(tool)

	return &ToolInstanceHistoryDTO{
		ID:             inst.ID,
		ToolID:         inst.ToolID,
		ToolName:       inst.ToolName,
		Tool:           t,
		LaunchUserName: inst.LaunchUserName,
		StartTime:      inst.StartTime,
		EndTime:        inst.EndTime,
		Ret:            inst.Ret,
		Err:            inst.Err,
		IsRunning:      inst.IsRunning,
	}
}

type KillToolInstanceReq struct {
	Signal int `json:"signal" binding:"required"`
}
