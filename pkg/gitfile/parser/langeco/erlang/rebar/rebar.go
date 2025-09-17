package rebar  
  
import (  
    "github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser"  
    "github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser/langeco"  
)  

type rebarConfig struct {
    appName   string `json:"appname"`
    deps string `json:"dependencies"`
}
  
func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {  
    var config rebarConfig
    config.appName = GetAppName(content)
    config.deps = GetDeps(content)
    
    pkg := &langeco.Package{
        Name:    config.appName,
        Version: "",
        Eco:     parser.REBAR,
    }
    
    deps := &langeco.Dependencies{
        DependencyList: []langeco.Dependency{},
    }

    return pkg, deps, nil
}