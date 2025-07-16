package model

import "github.com/HUSTSecLab/OpenSift/pkg/storage/repository"

type UpdateDistributionGitLinkReq struct {
	Distribution string   `json:"distribution" binding:"required"`
	PackageName  string   `json:"packageName" binding:"required"`
	Link         *string  `json:"link"`
	Confidence   *float32 `json:"confidence"`
}

type DistributionPackageDTO struct {
	Package        string   `json:"package"`
	HomePage       string   `json:"homePage"`
	Description    string   `json:"description"`
	Version        string   `json:"version"`
	GitLink        string   `json:"gitLink"`
	LinkConfidence *float32 `json:"linkConfidence"`
}

type GitLinkAICompletionReq struct {
	Distribution string `json:"distribution" binding:"required"`
	PackageName  string `json:"packageName" binding:"required"`
	Description  string `json:"description"`
	HomePage     string `json:"homePage"`
}

func ToDistributionPackageDTO(pkg *repository.DistPackage) *DistributionPackageDTO {
	if pkg == nil {
		return nil
	}
	return &DistributionPackageDTO{
		Package:        *pkg.Package,
		HomePage:       *pkg.HomePage,
		Description:    *pkg.Description,
		Version:        *pkg.Version,
		GitLink:        *pkg.GitLink,
		LinkConfidence: *pkg.LinkConfidence,
	}
}
