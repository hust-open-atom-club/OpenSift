package enumerator

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/HUSTSecLab/OpenSift/pkg/logger"
	"github.com/bytedance/gopkg/util/gopool"
)

// API begin

type pypiIndexResp struct {
	Projects []pypiIndexItem `json:"projects"`
}

type pypiIndexItem struct {
	LastSerial int    `json:"_last-serial"`
	Name       string `json:"name"`
}

func (p *pypiEnumerator) getPypiIndex() ([]pypiIndexItem, error) {
	r := p.client.R()
	r.SetURL("https://pypi.org/simple/")
	r.SetHeader("Accept", "application/vnd.pypi.simple.v1+json")
	resp := r.Do()
	if resp.IsErrorState() {
		logger.Errorf("Pypi index fetch failed: %v", resp)
		return nil, resp.Err
	}
	indexResp := &pypiIndexResp{}
	err := json.Unmarshal(resp.Bytes(), indexResp)
	if err != nil {
		logger.Errorf("Pypi index unmarshal failed: %v", err)
		return nil, err
	}
	return indexResp.Projects, nil
}

func UnmarshalPypiPackageResposne(data []byte) (PypiPackageResp, error) {
	var r PypiPackageResp
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *PypiPackageResp) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type PypiPackageResp struct {
	Info            Info             `json:"info"`
	LastSerial      int64            `json:"last_serial"`
	Releases        map[string][]URL `json:"releases"`
	Urls            []URL            `json:"urls"`
	Vulnerabilities []interface{}    `json:"vulnerabilities"`
}

type Info struct {
	Author                 string      `json:"author"`
	AuthorEmail            string      `json:"author_email"`
	BugtrackURL            interface{} `json:"bugtrack_url"`
	Classifiers            []string    `json:"classifiers"`
	Description            string      `json:"description"`
	DescriptionContentType string      `json:"description_content_type"`
	DocsURL                interface{} `json:"docs_url"`
	DownloadURL            interface{} `json:"download_url"`
	Downloads              Downloads   `json:"downloads"`
	Dynamic                interface{} `json:"dynamic"`
	HomePage               string      `json:"home_page"`
	Keywords               interface{} `json:"keywords"`
	License                string      `json:"license"`
	LicenseExpression      interface{} `json:"license_expression"`
	LicenseFiles           interface{} `json:"license_files"`
	Maintainer             interface{} `json:"maintainer"`
	MaintainerEmail        interface{} `json:"maintainer_email"`
	Name                   string      `json:"name"`
	PackageURL             string      `json:"package_url"`
	Platform               interface{} `json:"platform"`
	ProjectURL             string      `json:"project_url"`
	ProjectUrls            ProjectUrls `json:"project_urls"`
	ProvidesExtra          []string    `json:"provides_extra"`
	ReleaseURL             string      `json:"release_url"`
	RequiresDist           []string    `json:"requires_dist"`
	RequiresPython         string      `json:"requires_python"`
	Summary                string      `json:"summary"`
	Version                string      `json:"version"`
	Yanked                 bool        `json:"yanked"`
	YankedReason           interface{} `json:"yanked_reason"`
}

type Downloads struct {
	LastDay   int64 `json:"last_day"`
	LastMonth int64 `json:"last_month"`
	LastWeek  int64 `json:"last_week"`
}

type ProjectUrls struct {
	Documentation string `json:"Documentation"`
	Homepage      string `json:"Homepage"`
	Source        string `json:"Source"`
}

type URL struct {
	CommentText       string        `json:"comment_text"`
	Digests           Digests       `json:"digests"`
	Downloads         int64         `json:"downloads"`
	Filename          string        `json:"filename"`
	HasSig            bool          `json:"has_sig"`
	Md5Digest         string        `json:"md5_digest"`
	Packagetype       Packagetype   `json:"packagetype"`
	PythonVersion     PythonVersion `json:"python_version"`
	RequiresPython    *string       `json:"requires_python"`
	Size              int64         `json:"size"`
	UploadTime        string        `json:"upload_time"`
	UploadTimeISO8601 string        `json:"upload_time_iso_8601"`
	URL               string        `json:"url"`
	Yanked            bool          `json:"yanked"`
	YankedReason      *string       `json:"yanked_reason"`
}

type Digests struct {
	Blake2B256 string `json:"blake2b_256"`
	Md5        string `json:"md5"`
	Sha256     string `json:"sha256"`
}

type Packagetype string

const (
	BdistEgg   Packagetype = "bdist_egg"
	BdistWheel Packagetype = "bdist_wheel"
	Sdist      Packagetype = "sdist"
)

type PythonVersion string

const (
	Py2Py3 PythonVersion = "py2.py3"
	Py3    PythonVersion = "py3"
	Source PythonVersion = "source"
	The27  PythonVersion = "2.7"
)

func (p *pypiEnumerator) getPypiPackageInfoWithoutCache(name string) (*PypiPackageResp, error) {
	r := p.client.R()
	r.SetURL(fmt.Sprintf("https://pypi.org/pypi/%s/json", name))
	resp := r.Do()
	if resp.IsErrorState() {
		logger.Errorf("Pypi package %s fetch failed: %v", name, resp)
		return nil, resp.Err
	}
	packageResp, err := UnmarshalPypiPackageResposne(resp.Bytes())

	if err != nil {
		logger.Errorf("Pypi package %s unmarshal failed: %v", name, err)
		return nil, err
	}
	return &packageResp, nil
}

func (p *pypiEnumerator) getPypiPackageInfo(name string, serial int) (*PypiPackageResp, error) {
	// TODO: if serial is same, return cache, else fetch and update cache
	return p.getPypiPackageInfoWithoutCache(name)
}

// API end

type PypiEnumeratorConfig struct {
	Jobs int
}

type pypiEnumerator struct {
	enumeratorBase
	config *PypiEnumeratorConfig
}

// Enumerate implements Enumerator.
func (p *pypiEnumerator) Enumerate() error {
	logger.Info("Fetching pypi index")
	projects, err := p.getPypiIndex()
	if err != nil {
		logger.Errorf("Pypi index fetch failed: %v", err)
		return err
	}

	if err := p.writer.Open(); err != nil {
		return err
	}
	defer p.writer.Close()

	var wg sync.WaitGroup

	pool := gopool.NewPool("pypi_enumerator", int32(p.config.Jobs), &gopool.Config{})

	for _, project := range projects {
		wg.Add(1)
		pool.Go(func() {
			pkg, err := p.getPypiPackageInfo(project.Name, project.LastSerial)
			if pkg == nil || err != nil {
				logger.Errorf("Pypi package %s fetch failed: %v", project.Name, err)
				return
			}
			if pkg.Info.ProjectUrls.Source == "" {
				// logger.Warnf("Package %s has no source url", project.Name)
				return
			}

			if err := p.writer.Write(pkg.Info.ProjectUrls.Source); err != nil {
				logger.Errorf("Pypi package %s write failed: %v", project.Name, err)
				return
			}
			wg.Done()
		})
	}
	wg.Wait()
	return nil
}

var _ Enumerator = (*pypiEnumerator)(nil)

func NewPypiEnumerator(config *PypiEnumeratorConfig) Enumerator {
	return &pypiEnumerator{
		enumeratorBase: newEnumeratorBase(),
		config:         config,
	}
}
