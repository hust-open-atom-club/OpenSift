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
      
    //if deps, ok := config.GetDeps(); ok {  
    //    rebarCfg.deps = deps  
    //}  
      
    pkg := &langeco.Package{  
        Name:    rebarCfg.appName,  
        Version: "",  
        Eco:     parser.REBAR, 
    }  
      
    deps := make(langeco.Dependencies, 0)  
      
    if depTerms, ok := config.GetDeps(); ok {
        for _, term := range depTerms {
            switch t := term.(type) {
            case rebarParser.List:
                // 遍历 List 中的每个元素（期望是 Tuple）
                for _, elem := range t.Elements {
                    if tuple, ok := elem.(rebarParser.Tuple); ok && len(tuple.Elements) >= 2 { // ✅ 修正1：len(tuple.Elements)
                        var name, version string

                        // 第一个元素：依赖名（Atom）
                        if atom, ok := tuple.Elements[0].(rebarParser.Atom); ok { // ✅ 修正2：tuple.Elements[0]
                            name = string(atom.Value)
                        } else {
                            continue // 跳过无效项
                        }

                        // 第二个元素：版本（String 或 Atom）
                        switch v := tuple.Elements[1].(type) { // ✅ 修正3：tuple.Elements[1]
                        case rebarParser.String:
                            version = string(v.Value)
                        case rebarParser.Atom:
                            version = string(v.Value)
                        default:
                            version = "unknown"
                        }

                        deps = append(deps, langeco.Package{ // ✅ 确保 langeco.Dependency 已定义
                            Name:    name,
                            Version: version,
                        })
                    }
                }
            }
        }
    }
      
    return pkg, &deps, nil  
}