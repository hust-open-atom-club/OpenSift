package nuget

import (
	"encoding/xml"

	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
)

type cfgPackageReference struct {
	XMLName         xml.Name `xml:"package"`
	TargetFramework string   `xml:"targetFramework,attr"`
	Version         string   `xml:"version,attr"`
	DevDependency   bool     `xml:"developmentDependency,attr"`
	ID              string   `xml:"id,attr"`
}

type config struct {
	XMLName  xml.Name              `xml:"packages"`
	Packages []cfgPackageReference `xml:"package"`
}

func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {
	var cfgData config
	if err := xml.Unmarshal([]byte(content), &cfgData); err != nil {
		return nil, nil, err
	}

	var deps langeco.Dependencies
	for _, dep := range cfgData.Packages {
		if dep.ID == "" || dep.DevDependency {
			continue
		}

		deps = append(deps, langeco.Package{
			Name:    dep.ID,
			Version: dep.Version,
			Eco:     parser.NUGET,
		})
	}

	return &langeco.Package{
		Eco: parser.NUGET,
	}, &deps, nil
}
