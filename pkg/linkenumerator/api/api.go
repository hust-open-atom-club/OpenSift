package api

import (
	"encoding/json"

	"github.com/HUSTSecLab/criticality_score/pkg/linkenumerator/api/bitbucket"
	"github.com/HUSTSecLab/criticality_score/pkg/linkenumerator/api/cargo"
	"github.com/HUSTSecLab/criticality_score/pkg/linkenumerator/api/gitlab"
	"github.com/imroc/req/v3"
)

const (
	PER_PAGE      = 100
	TIME_INTERVAL = 2
	TIMEOUT       = 1000

	BITBUCKET_ENUMERATE_API_URL = "https://api.bitbucket.org/2.0/repositories?pagelen=200"
	GITLAB_ENUMERATE_API_URL    = "https://gitlab.com/api/v4/projects"
	GITEE_ENUMERATE_API_URL     = "https://api.indexea.com/v1/search/widget/wjawvtmm7r5t25ms1u3d"
	CRATES_IO_ENUMERATE_API_URL = "https://crates.io/api/v1/crates"
	PYPI_API_URL                = "https://pypi.org/simple/"
	NUGET_INDEX_URL             = "https://azuresearch-ea.nuget.org/query"

	GITLAB_TOTAL_PAGES = 100000

	BITBUCKET_ENUMERATE_PAGE = 40 //* repo_num = page * 10
	GITLAB_ENUMERATE_PAGE    = 20 //* repo_num = page * 100
	GITEE_ENUMERATE_PAGE     = 20 //* repo_num = page * 100
	CRATES_IO_ENUMERATE_PAGE = 20
)

func FromGitlab(res *req.Response) (*gitlab.Response, error) {
	resp := &gitlab.Response{}
	if err := json.Unmarshal(res.Bytes(), resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func FromBitbucket(res *req.Response) (*bitbucket.Response, error) {
	resp := &bitbucket.Response{}
	if err := json.Unmarshal(res.Bytes(), resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func FromCargo(res *req.Response) (*cargo.Response, error) {
	resp := &cargo.Response{}
	if err := json.Unmarshal(res.Bytes(), resp); err != nil {
		return nil, err
	}
	return resp, nil
}
