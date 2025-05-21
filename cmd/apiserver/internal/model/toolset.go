package model

import (
	"time"

	"github.com/HUSTSecLab/criticality_score/cmd/apiserver/internal/tool"
)

type ToolDTO struct {
	// ID is the unique identifier for the toolset.
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Args        []ToolArgDTO `json:"args"`
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

type ToolInstanceDTO struct {
	ID        string    `json:"id"`
	Tool      *ToolDTO  `json:"tool"`
	StartTime time.Time `json:"startTime"`
}

func ToolArgToToolArgDTO(arg tool.ToolArg) ToolArgDTO {
	return ToolArgDTO{
		Name:        arg.Name,
		Type:        string(arg.Type),
		Default:     arg.Default,
		Description: arg.Description,
	}
}

func ToolToToolDTO(tool *tool.Tool) *ToolDTO {
	toolDTO := &ToolDTO{
		ID:          tool.ID,
		Name:        tool.Name,
		Description: tool.Description,
	}
	if len(tool.Args) > 0 {
		toolDTO.Args = make([]ToolArgDTO, len(tool.Args))
		for i, arg := range tool.Args {
			toolDTO.Args[i] = ToolArgToToolArgDTO(arg)
		}
	}
	return toolDTO
}

func ToolInstanceToToolInstanceDTO(inst *tool.ToolInstance) *ToolInstanceDTO {
	return &ToolInstanceDTO{
		ID:        inst.ID,
		Tool:      ToolToToolDTO(inst.Tool),
		StartTime: inst.StartTime,
	}
}
