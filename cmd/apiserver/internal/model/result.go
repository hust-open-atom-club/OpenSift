package model

import (
	"time"

	"github.com/HUSTSecLab/OpenSift/pkg/storage/repository"
)

type ResultGitMetadataDTO struct {
	License          *[]string  `json:"license"`
	Language         *[]string  `json:"language"`
	CreatedSince     *time.Time `json:"createdSince"`
	UpdatedSince     *time.Time `json:"updatedSince"`
	ContributorCount *int       `json:"contributorCount"`
	OrgCount         *int       `json:"orgCount"`
	CommitFrequency  *float64   `json:"commitFrequency"`
	UpdateTime       *time.Time `json:"updateTime"`
}

type ResultLangDetailDTO struct {
	Type          *int       `json:"type"`
	LangEcoImpact *float64   `json:"langEcoImpact"`
	DepCount      *int       `json:"depCount"`
	UpdateTime    *time.Time `json:"updateTime"`
}

type ResultDistDetailDTO struct {
	Type       *int       `json:"type"`
	Count      *int       `json:"count"`
	Impact     *float64   `json:"impact"`
	PageRank   *float64   `json:"pageRank"`
	UpdateTime *time.Time `json:"updateTime"`
}

type ResultDTO struct {
	ScoreID     *int                   `json:"scoreID"`
	GitLink     string                 `json:"link"`
	GitScore    *float64               `json:"gitScore"`
	GitDetail   []ResultGitMetadataDTO `json:"gitDetail"`
	LangDetail  []ResultLangDetailDTO  `json:"langDetail"`
	DistDetail  []ResultDistDetailDTO  `json:"distDetail"`
	DistroScore *float64               `json:"distroScore"`
	LangScore   *float64               `json:"langScore"`
	Score       *float64               `json:"score"`
	UpdateTime  *time.Time             `json:"updateTime"`
}

type RankingResultDTO struct {
	ResultDTO
	Ranking int `json:"ranking"`
}

func ResultDOToDTO(r *repository.Result) *ResultDTO {
	return &ResultDTO{
		ScoreID:     *r.ScoreID,
		GitLink:     *r.GitLink,
		GitScore:    *r.GitScore,
		DistroScore: *r.DistScore,
		LangScore:   *r.LangScore,
		Score:       *r.Score,
		UpdateTime:  *r.UpdateTime,
	}
}

func ResultGitDetailDOToDTO(r *repository.ResultGitDetail) *ResultGitMetadataDTO {
	return &ResultGitMetadataDTO{
		License:          (*[]string)(*r.License),
		Language:         (*[]string)(*r.Language),
		CreatedSince:     *r.CreatedSince,
		UpdatedSince:     *r.UpdatedSince,
		ContributorCount: *r.ContributorCount,
		OrgCount:         *r.OrgCount,
		CommitFrequency:  *r.CommitFrequency,
		UpdateTime:       *r.UpdateTime,
	}
}

func ResultLangDetailDOToDTO(r *repository.ResultLangDetail) *ResultLangDetailDTO {
	return &ResultLangDetailDTO{
		Type:          *r.Type,
		LangEcoImpact: *r.LangEcoImpact,
		DepCount:      *r.DepCount,
		UpdateTime:    *r.UpdateTime,
	}
}

func ResultDistDetailDOToDTO(r *repository.ResultDistDetail) *ResultDistDetailDTO {
	return &ResultDistDetailDTO{
		Type:       *r.Type,
		Count:      *r.Count,
		Impact:     *r.Impact,
		PageRank:   *r.PageRank,
		UpdateTime: *r.UpdateTime,
	}
}

func RankingDOToDTO(r *repository.RankingResult) *RankingResultDTO {
	return &RankingResultDTO{
		ResultDTO: *ResultDOToDTO(&repository.Result{
			ScoreID:    r.ScoreID,
			GitLink:    r.GitLink,
			GitScore:   r.GitScore,
			DistScore:  r.DistScore,
			LangScore:  r.LangScore,
			Score:      r.Score,
			UpdateTime: r.UpdateTime,
		}),
		Ranking: *r.Ranking,
	}
}
