package rebar  
  
import (  
    "github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser"  
    "github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser/langeco"  
    rebarParser "github.com/scagogogo/erlang-rebar-config-parser/pkg/parser"  
)  
  
type rebarConfig struct {  
    appName string `json:"appname"`  
    deps    string `json:"dependencies"`  
}  
  
func Parse(content string) (*langeco.Package, *langeco.Dependencies, error) {  
    config, err := rebarParser.Parse(content)  
    if err != nil {  
        return nil, nil, err  
    }  
      
    var rebarCfg rebarConfig  
      
    if appName, ok := config.GetAppName(); ok {  
        rebarCfg.appName = appName  
    }  
      
    if deps, ok := config.GetDeps(); ok {  
        rebarCfg.deps = deps  
    }  
      
    pkg := &langeco.Package{  
        Name:    rebarCfg.appName,  
        Version: "",  
        Eco:     parser.REBAR, 
    }  
      
    deps := make(langeco.Dependencies, 0)  
      
    // TODO: 解析rebarCfg.deps并填充到deps中  
      
    return pkg, &deps, nil  
}