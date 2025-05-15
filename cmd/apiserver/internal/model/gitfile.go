package model

import (
	"time"

	"github.com/HUSTSecLab/criticality_score/cmd/git-metadata-collector/rpc"
	"github.com/HUSTSecLab/criticality_score/pkg/storage/repository"
)

type GitFileStatusResp struct {
	GitFile   *GitFileStatisticsResultDTO `json:"gitFile"`
	Collector *rpc.StatusResp             `json:"collector"`
}

type GitFileAppendManualReq struct {
	GitLink string `json:"gitLink"`
}

type GitFileDTO struct {
	GitLink     string     `json:"gitLink"`
	FilePath    string     `json:"filePath"`
	Success     bool       `json:"success"`
	Message     *string    `json:"message"`
	UpdateTime  *time.Time `json:"updateTime"`
	FailedTimes *int       `json:"failedTimes"`
	LastSuccess *time.Time `json:"lastSuccess"`
	TakeTimeMs  *int64     `json:"takeTimeMs"`
	TakeStorage *int64     `json:"takeStorage"`
}

func GitFileDOToDTO(f *repository.GitFile) *GitFileDTO {
	return &GitFileDTO{
		*f.GitLink,
		*f.FilePath,
		*f.Success,
		*f.Message,
		*f.UpdateTime,
		*f.FailedTimes,
		*f.LastSuccess,
		*f.TakeTimeMs,
		*f.TakeStorage,
	}
}

type GitFileStatisticsResultDTO struct {
	Total        *int `json:"total"`
	Success      *int `json:"success"`
	Fail         *int `json:"fail"`
	NeverSuccess *int `json:"neverSuccess"`
}

func GitFileStatisticsResultDOToDTO(r *repository.GitFileStatisticsResult) *GitFileStatisticsResultDTO {
	return &GitFileStatisticsResultDTO{
		r.Total,
		r.Success,
		r.Fail,
		r.NeverSuccess,
	}

}
