package maven

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
)

type Pom struct {
	GroupId                string                 `xml:"groupId"`
	ArtifactId             string                 `xml:"artifactId"`
	Build                  PomBuild               `xml:"build"`
	Contributors           Contributors           `xml:"contributors"`
	Modules                Modules                `xml:"modules"`
	Dependencies           Dependencies           `xml:"dependencies"`
	DependencyManagement   DependencyManagement   `xml:"dependencyManagement"`
	Description            string                 `xml:"description"`
	Developers             Developers             `xml:"developers"`
	DistributionManagement DistributionManagement `xml:"distributionManagement"`
	InceptionYear          string                 `xml:"inceptionYear"`
	IssueManagement        IssueManagement        `xml:"issueManagement"`
	ModelVersion           string                 `xml:"modelVersion"`
	Name                   string                 `xml:"name"`
	Parent                 Parent                 `xml:"parent"`
	Profiles               Profiles               `xml:"profiles"`
	Properties             Properties             `xml:"properties"`
	Reporting              Reporting              `xml:"reporting"`
	SCM                    SCM                    `xml:"scm"`
	URL                    string                 `xml:"url"`
	Version                string                 `xml:"version"`
}

type PomBuild struct {
	DefaultGoal      string           `xml:"defaultGoal"`
	PluginManagement PluginManagement `xml:"pluginManagement"`
	Plugins          PurplePlugins    `xml:"plugins"`
}

type PluginManagement struct {
	Plugins PluginManagementPlugins `xml:"plugins"`
}

type PluginManagementPlugins struct {
	Plugin PurplePlugin `xml:"plugin"`
}

type PurplePlugin struct {
	ArtifactID    string              `xml:"artifactId"`
	Configuration PurpleConfiguration `xml:"configuration"`
	GroupID       string              `xml:"groupId"`
}

type PurpleConfiguration struct {
	Excludes PurpleExcludes `xml:"excludes"`
}

type PurpleExcludes struct {
	Exclude []string `xml:"exclude"`
}

type PurplePlugins struct {
	Plugin []FluffyPlugin `xml:"plugin"`
}

type FluffyPlugin struct {
	ArtifactID    string              `xml:"artifactId"`
	Configuration FluffyConfiguration `xml:"configuration"`
	Dependencies  PluginDependencies  `xml:"dependencies"`
	Executions    PurpleExecutions    `xml:"executions"`
	GroupID       string              `xml:"groupId"`
	Version       string              `xml:"version"`
}

type FluffyConfiguration struct {
	Archive                    *Archive             `xml:"archive,omitempty"`
	ConfigLocation             *string              `xml:"configLocation,omitempty"`
	Descriptors                *Descriptors         `xml:"descriptors,omitempty"`
	EnableRulesSummary         *string              `xml:"enableRulesSummary,omitempty"`
	ExcludeFilterFile          string               `xml:"excludeFilterFile"`
	IgnorePathsToDelete        *IgnorePathsToDelete `xml:"ignorePathsToDelete,omitempty"`
	IncludeTestSourceDirectory *string              `xml:"includeTestSourceDirectory,omitempty"`
	Links                      *Links               `xml:"links,omitempty"`
	Notimestamp                *string              `xml:"notimestamp,omitempty"`
	Quiet                      *string              `xml:"quiet,omitempty"`
	Source                     *string              `xml:"source,omitempty"`
	TarLongFileMode            *string              `xml:"tarLongFileMode,omitempty"`
}

type Archive struct {
	Manifest        *Manifest       `xml:"manifest,omitempty"`
	ManifestEntries ManifestEntries `xml:"manifestEntries"`
}

type Manifest struct {
	AddDefaultImplementationEntries string `xml:"addDefaultImplementationEntries"`
	AddDefaultSpecificationEntries  string `xml:"addDefaultSpecificationEntries"`
}

type ManifestEntries struct {
	AutomaticModuleName string `xml:"Automatic-Module-Name"`
}

type Descriptors struct {
	Descriptor []string `xml:"descriptor"`
}

type IgnorePathsToDelete struct {
	IgnorePathToDelete string `xml:"ignorePathToDelete"`
}

type Links struct {
	Link []string `xml:"link"`
}

type PluginDependencies struct {
	Dependency PurpleDependency `xml:"dependency"`
}

type PurpleDependency struct {
	ArtifactID string `xml:"artifactId"`
	GroupID    string `xml:"groupId"`
	Version    string `xml:"version"`
}

type PurpleExecutions struct {
	Execution PurpleExecution `xml:"execution"`
}

type PurpleExecution struct {
	Configuration *TentacledConfiguration `xml:"configuration,omitempty"`
	Goals         PurpleGoals             `xml:"goals"`
	ID            *string                 `xml:"id,omitempty"`
	Phase         *string                 `xml:"phase,omitempty"`
}

type TentacledConfiguration struct {
	Includes Includes `xml:"includes"`
	RunOrder string   `xml:"runOrder"`
}

type Includes struct {
	Include string `xml:"include"`
}

type PurpleGoals struct {
	Goal *Goal `xml:"goal"`
}

type Contributors struct {
	Contributor []Contributor `xml:"contributor"`
}

type Contributor struct {
	Name string `xml:"name"`
}

type Dependencies struct {
	Dependency []Dependency `xml:"dependency"`
}

type Exclusion struct {
	Text       string `xml:",chardata"`
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
}

type Exclusions struct {
	Text      string    `xml:",chardata"`
	Exclusion Exclusion `xml:"exclusion"`
}

type Dependency struct {
	Text       string     `xml:",chardata"`
	GroupId    string     `xml:"groupId"`
	ArtifactId string     `xml:"artifactId"`
	Version    string     `xml:"version"`
	Scope      string     `xml:"scope"`
	Type       string     `xml:"type"`
	Exclusions Exclusions `xml:"exclusions"`
}

type DependencyManagement struct {
	Dependencies DependencyManagementDependencies `xml:"dependencies"`
}

type DependencyManagementDependencies struct {
	Dependency FluffyDependency `xml:"dependency"`
}

type FluffyDependency struct {
	ArtifactID string `xml:"artifactId"`
	GroupID    string `xml:"groupId"`
	Scope      string `xml:"scope"`
	Type       string `xml:"type"`
	Version    string `xml:"version"`
}

type Developers struct {
	Developer []Developer `xml:"developer"`
}

type Developer struct {
	Email        string `xml:"email"`
	ID           string `xml:"id"`
	Name         string `xml:"name"`
	Organization string `xml:"organization"`
	Roles        Roles  `xml:"roles"`
	Timezone     string `xml:"timezone"`
}

type Roles struct {
	Role string `xml:"role"`
}

type DistributionManagement struct {
	Site Site `xml:"site"`
}

type Site struct {
	ID   string `xml:"id"`
	Name string `xml:"name"`
	URL  string `xml:"url"`
}

type IssueManagement struct {
	System string `xml:"system"`
	URL    string `xml:"url"`
}

type Parent struct {
	ArtifactID string `xml:"artifactId"`
	GroupID    string `xml:"groupId"`
	Version    string `xml:"version"`
}

type Profiles struct {
	Profile []Profile `xml:"profile"`
}

type Profile struct {
	Activation Activation        `xml:"activation"`
	Build      ProfileBuild      `xml:"build"`
	ID         string            `xml:"id"`
	Properties ProfileProperties `xml:"properties"`
}

type Activation struct {
	File *File  `xml:"file,omitempty"`
	JDK  string `xml:"jdk"`
}

type File struct {
	Missing string `xml:"missing"`
}

type ProfileBuild struct {
	Plugins FluffyPlugins `xml:"plugins"`
}

type FluffyPlugins struct {
	Plugin TentacledPlugin `xml:"plugin"`
}

type TentacledPlugin struct {
	ArtifactID    string               `xml:"artifactId"`
	Configuration *StickyConfiguration `xml:"configuration,omitempty"`
	Executions    FluffyExecutions     `xml:"executions"`
	GroupID       string               `xml:"groupId"`
	Version       string               `xml:"version"`
}

type StickyConfiguration struct {
	Excludes FluffyExcludes `xml:"excludes"`
}

type FluffyExcludes struct {
	Exclude string `xml:"exclude"`
}

type FluffyExecutions struct {
	Execution FluffyExecution `xml:"execution"`
}

type FluffyExecution struct {
	Configuration IndigoConfiguration `xml:"configuration"`
	Goals         FluffyGoals         `xml:"goals"`
	ID            string              `xml:"id"`
	Phase         string              `xml:"phase"`
}

type IndigoConfiguration struct {
	Arguments      Arguments `xml:"arguments"`
	ClasspathScope string    `xml:"classpathScope"`
	Executable     string    `xml:"executable"`
	Tasks          *Tasks    `xml:"tasks,omitempty"`
}

type Arguments struct {
	Argument  []string `xml:"argument"`
	Classpath string   `xml:"classpath"`
}

type Tasks struct {
	Exec        []Exec      `xml:"exec"`
	Pathconvert Pathconvert `xml:"pathconvert"`
}

type Exec struct {
	Arg string `xml:"arg"`
}

type Pathconvert struct {
	Dirset string `xml:"dirset"`
}

type FluffyGoals struct {
	Goal string `xml:"goal"`
}

type ProfileProperties struct {
	ArgLine       *string `xml:"argLine,omitempty"`
	Benchmark     string  `xml:"benchmark"`
	CoverallsSkip *string `xml:"coveralls.skip,omitempty"`
	JacocoSkip    *string `xml:"jacoco.skip,omitempty"`
	SkipTests     string  `xml:"skipTests"`
}

type Properties map[string]string

func (m *Properties) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	*m = Properties{}
	for {
		token, err := d.Token()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		switch tt := token.(type) {
		case xml.StartElement:
			var value string
			if err := d.DecodeElement(&value, &tt); err != nil {
				return err
			}
			(*m)[tt.Name.Local] = value

		case xml.EndElement:
			if tt == start.End() {
				return nil
			}
		}
	}
}

type Reporting struct {
	Plugins ReportingPlugins `xml:"plugins"`
}

type ReportingPlugins struct {
	Plugin []StickyPlugin `xml:"plugin"`
}

type StickyPlugin struct {
	ArtifactID    string                `xml:"artifactId"`
	Configuration IndecentConfiguration `xml:"configuration"`
	GroupID       string                `xml:"groupId"`
	ReportSets    *ReportSets           `xml:"reportSets,omitempty"`
	Version       string                `xml:"version"`
}

type IndecentConfiguration struct {
	ConfigLocation             *string        `xml:"configLocation,omitempty"`
	EnableRulesSummary         *string        `xml:"enableRulesSummary,omitempty"`
	ExcludeFilterFile          *string        `xml:"excludeFilterFile,omitempty"`
	IncludeTestSourceDirectory *string        `xml:"includeTestSourceDirectory,omitempty"`
	TagListOptions             TagListOptions `xml:"tagListOptions"`
	TargetJDK                  *string        `xml:"targetJdk,omitempty"`
}

type TagListOptions struct {
	TagClasses TagClasses `xml:"tagClasses"`
}

type TagClasses struct {
	TagClass []TagClass `xml:"tagClass"`
}

type TagClass struct {
	DisplayName string `xml:"displayName"`
	Tags        Tags   `xml:"tags"`
}

type Tags struct {
	Tag []Tag `xml:"tag"`
}

type Tag struct {
	MatchString string `xml:"matchString"`
	MatchType   string `xml:"matchType"`
}

type ReportSets struct {
	ReportSet ReportSet `xml:"reportSet"`
}

type ReportSet struct {
	Reports Reports `xml:"reports"`
}

type Reports struct {
	Report string `xml:"report"`
}

type SCM struct {
	Connection          string `xml:"connection"`
	DeveloperConnection string `xml:"developerConnection"`
	Tag                 string `xml:"tag"`
	URL                 string `xml:"url"`
}

type Goal struct {
	String      *string
	StringArray []string
}

type Modules struct {
	M []string `xml:"module"`
}

func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {
	pom := Pom{}
	if err := xml.Unmarshal([]byte(content), &pom); err != nil {
		return nil, nil, err
	}

	pkg := langeco.Package{
		Name:    fmt.Sprintf("%s/%s", pom.GroupId, pom.ArtifactId),
		Version: pom.Version,
		Eco:     parser.MAVEN,
	}

	deps := make(langeco.Dependencies, 0)

	for _, dep := range pom.Dependencies.Dependency {
		version := checkMacro(&pom.Properties, dep.Version)
		groupId := checkMacro(&pom.Properties, dep.GroupId)
		artifactId := checkMacro(&pom.Properties, dep.ArtifactId)
		deps = append(deps, langeco.Package{
			Name:    fmt.Sprintf("%s/%s", groupId, artifactId),
			Version: version,
			Eco:     parser.MAVEN,
		})
	}

	return &pkg, &deps, nil
}

func checkMacro(p *Properties, s string) string {
	if strings.Contains(s, "${") {
		if v, ok := (*p)[s[2:len(s)-1]]; ok {
			return v
		}
	}
	return s
}
