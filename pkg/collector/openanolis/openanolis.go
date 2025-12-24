package openanolis

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"strings"

	collector "github.com/HUSTSecLab/OpenSift/pkg/collector/internal"
	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/HUSTSecLab/OpenSift/pkg/storage/repository"
)

type OpenAnolisCollector struct {
	collector.CollecterInterface
}

func (oc *OpenAnolisCollector) Collect(outputPath string) {
	adc := storage.GetDefaultAppDatabaseContext()
	data := oc.GetPackageInfo(collector.OpenAnolisURL)
	oc.ParseInfo(data)
	oc.GetDep()
	oc.PageRank(0.85, 20)
	oc.GetDepCount()
	oc.UpdateRelationships(adc)
	oc.UpdateDistRepoCount(adc)
	oc.CalculateDistImpact()
	oc.UpdateOrInsertDatabase(adc)
	oc.UpdateOrInsertDistDependencyDatabase(adc)
	if outputPath != "" {
		err := oc.GenerateDependencyGraph(outputPath)
		if err != nil {
			log.Printf("Error generating dependency graph: %v\n", err)
			return
		}
	}
}
func (oc *OpenAnolisCollector) ParseInfo(data string) error {
	data = strings.Replace(data, "\x00", "", -1)
	decoder := xml.NewDecoder(strings.NewReader(data))
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		if charset == "utf-8" {
			return input, nil
		}
		return nil, fmt.Errorf("unsupported charset: %s", charset)
	}

	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		switch se := tok.(type) {
		case xml.StartElement:
			if se.Name.Local == "package" {
				var pkgData struct {
					Type string `xml:"type,attr"`
					XML  string `xml:",innerxml"`
				}
				err := decoder.DecodeElement(&pkgData, &se)
				if err != nil {
					return err
				}

				if pkgData.Type == "rpm" {
					lines := strings.Split(pkgData.XML, "\n")
					for i, line := range lines {
						if len(line) > 2 {
							lines[i] = line[2:]
						}
					}
					trimmedXML := strings.Join(lines, "\n")
					pkgInfo, err := parsePackageXML(trimmedXML)
					if err != nil {
						return err
					}

					if exists := oc.GetPkgInfo(pkgInfo.Name); exists == nil {
						oc.SetPkgInfo(pkgInfo.Name, &pkgInfo)
					}
				}
			}
		}
	}
	return nil
}

func parsePackageXML(data string) (collector.PackageInfo, error) {
	data = strings.Map(func(r rune) rune {
		if r == '\x00' || r > 127 {
			return -1
		}
		return r
	}, data)
	decoder := xml.NewDecoder(strings.NewReader(data))
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		if charset == "utf-8" {
			return input, nil
		}
		return nil, fmt.Errorf("unsupported charset: %s", charset)
	}
	var pkgInfo collector.PackageInfo
	var depends []string

	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return collector.PackageInfo{}, err
		}

		switch se := tok.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "name":
				var name string
				if err := decoder.DecodeElement(&name, &se); err != nil {
					return collector.PackageInfo{}, err
				}
				pkgInfo.Name = name
			case "description":
				var description string
				if err := decoder.DecodeElement(&description, &se); err != nil {
					return collector.PackageInfo{}, err
				}
				if len(description) > 255 {
					description = description[:254]
				}
				pkgInfo.Description = description
			case "url":
				var url string
				if err := decoder.DecodeElement(&url, &se); err != nil {
					return collector.PackageInfo{}, err
				}
				pkgInfo.Homepage = url
			case "version":
				var version struct {
					Epoch string `xml:"epoch,attr"`
					Ver   string `xml:"ver,attr"`
					Rel   string `xml:"rel,attr"`
				}
				if err := decoder.DecodeElement(&version, &se); err != nil {
					return collector.PackageInfo{}, err
				}
				pkgInfo.Version = fmt.Sprintf("%s:%s-%s", version.Epoch, version.Ver, version.Rel)
			case "entry":
				var entry struct {
					Name string `xml:"name,attr"`
				}
				if err := decoder.DecodeElement(&entry, &se); err != nil {
					return collector.PackageInfo{}, err
				}
				depends = append(depends, entry.Name)
			}
		}
	}

	pkgInfo.DirectDepends = depends
	return pkgInfo, nil
}
func NewOpenAnolisCollector() *OpenAnolisCollector {
	return &OpenAnolisCollector{
		CollecterInterface: collector.NewCollector(
			repository.OpenAnolis,
			repository.DistPackageTablePrefix("openanolis"),
		),
	}
}
