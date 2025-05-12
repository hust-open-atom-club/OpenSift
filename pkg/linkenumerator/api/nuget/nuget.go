package nuget

type Response struct {
	Context   Context `json:"@context"`
	Data      []Datum `json:"data"`
	TotalHits int64   `json:"totalHits"`
}

type Context struct {
	Base  string `json:"@base"`
	Vocab string `json:"@vocab"`
}

type Datum struct {
	ID              string        `json:"@id"`
	Type            string        `json:"@type"`
	Authors         []string      `json:"authors"`
	Deprecation     Deprecation   `json:"deprecation"`
	Description     string        `json:"description"`
	IconURL         string        `json:"iconUrl"`
	DatumID         string        `json:"id"`
	LicenseURL      string        `json:"licenseUrl"`
	Owners          []string      `json:"owners"`
	PackageTypes    []PackageType `json:"packageTypes"`
	ProjectURL      string        `json:"projectUrl"`
	Registration    string        `json:"registration"`
	Summary         string        `json:"summary"`
	Tags            []string      `json:"tags"`
	Title           string        `json:"title"`
	TotalDownloads  int64         `json:"totalDownloads"`
	Verified        bool          `json:"verified"`
	Version         string        `json:"version"`
	Versions        []Version     `json:"versions"`
	Vulnerabilities []interface{} `json:"vulnerabilities"`
}

type Deprecation struct {
	AlternatePackage AlternatePackage `json:"alternatePackage"`
	Message          string           `json:"message"`
	Reasons          []string         `json:"reasons"`
}

type AlternatePackage struct {
	ID    string `json:"id"`
	Range string `json:"range"`
}

type PackageType struct {
	Name string `json:"name"`
}

type Version struct {
	ID        string `json:"@id"`
	Downloads int64  `json:"downloads"`
	Version   string `json:"version"`
}
