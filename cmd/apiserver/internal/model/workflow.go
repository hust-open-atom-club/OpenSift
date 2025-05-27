package model

type UpdateWorkflowStatusReq struct {
	Running bool `json:"running" binding:"required"`
}

type KillWorkflowJobReq struct {
	Type string `json:"type" binding:"required"` // "stop" or "kill"
}
