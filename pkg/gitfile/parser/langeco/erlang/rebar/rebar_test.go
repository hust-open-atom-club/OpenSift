package rebar  
  
import (  
    "os"  
    "path/filepath"  
    "testing"  
  
    "github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser"  
    "github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser/langeco"  
    "github.com/stretchr/testify/require"  
)  
  
func TestParse(t *testing.T) {  
    tests := []struct {  
        name     string  
        filename string  
        wantPkg  *langeco.Package  
        wantDeps *langeco.Dependencies  
        wantErr  bool  
    }{  
        {  
            name:     "Standard rebar.config",  
            filename: "standard.config",  
            wantPkg: &langeco.Package{  
                Name:    "my_app",  
                Version: "",  
                Eco:     parser.REBAR,  
            },  
            wantDeps: &langeco.Dependencies{  
                {  
                    Name:    "cowboy",  
                    Version: "2.9.0",  
                    Eco:     parser.REBAR,  
                },  
                {  
                    Name:    "jsx",  
                    Version: "3.1.0",  
                    Eco:     parser.REBAR,  
                },  
                {  
                    Name:    "lager",  
                    Version: "3.9.2",  
                    Eco:     parser.REBAR,  
                },  
            },  
            wantErr: false,  
        },  
        {  
            name:     "Rebar.config with git dependencies",  
            filename: "git_deps.config",  
            wantPkg: &langeco.Package{  
                Name:    "web_server",  
                Version: "",  
                Eco:     parser.REBAR,  
            },  
            wantDeps: &langeco.Dependencies{  
                {  
                    Name:    "cowboy",  
                    Version: "master",  
                    Eco:     parser.REBAR,  
                },  
                {  
                    Name:    "jiffy",  
                    Version: "1.1.1",  
                    Eco:     parser.REBAR,  
                },  
            },  
            wantErr: false,  
        },  
        {  
            name:     "Empty rebar.config",  
            filename: "empty.config",  
            wantPkg: &langeco.Package{  
                Name:    "",  
                Version: "",  
                Eco:     parser.REBAR,  
            },  
            wantDeps: &langeco.Dependencies{},  
            wantErr:  false,  
        },  
        {  
            name:     "Invalid rebar.config",  
            filename: "invalid.config",  
            wantPkg:  nil,  
            wantDeps: nil,  
            wantErr:  true,  
        },  
    }  
  
    for _, tt := range tests {  
        t.Run(tt.name, func(t *testing.T) {  
            path := filepath.Join("testdata", tt.filename)  
            data, err := os.ReadFile(path)  
            if err != nil {  
                t.Fatalf("Failed to read test file %s: %v", path, err)  
            }  
            input := string(data)  
  
            gotPkg, gotDeps, err := Parse(input)  
  
            if tt.wantErr {  
                require.Error(t, err)  
                return  
            }  
            require.NoError(t, err)  
  
            require.Equal(t, tt.wantPkg, gotPkg)  
            require.Equal(t, tt.wantDeps, gotDeps)  
        })  
    }  
}  
  
func TestGetAppName(t *testing.T) {  
    tests := []struct {  
        name     string  
        content  string  
        expected string  
    }{  
        {  
            name:     "Standard app name",  
            content:  `{application, my_app, [{description, "My Application"}]}.`,  
            expected: "my_app",  
        },  
    }  
  
    for _, tt := range tests {  
        t.Run(tt.name, func(t *testing.T) {  
            config, err := rebarParser.Parse(tt.content)  
            require.NoError(t, err)  
              
            if appName, ok := config.GetAppName(); ok {  
                require.Equal(t, tt.expected, appName)  
            } else {  
                require.Equal(t, tt.expected, "")  
            }  
        })  
    }  
}  
  
func TestGetDeps(t *testing.T) {  
    tests := []struct {  
        name     string  
        content  string  
        expected string  
    }{  
        {  
            name:     "Standard dependencies",  
            content:  `{deps, [{cowboy, "2.9.0"}, {jsx, "3.1.0"}]}.`,  
            expected: `[{cowboy, "2.9.0"}, {jsx, "3.1.0"}]`,  
        },  
    }  
  
    for _, tt := range tests {  
        t.Run(tt.name, func(t *testing.T) {  
            config, err := rebarParser.Parse(tt.content)  
            require.NoError(t, err)  
              
            if deps, ok := config.GetDeps(); ok {  
                require.Equal(t, tt.expected, deps)  
            } else {  
                require.Equal(t, tt.expected, "")  
            }  
        })  
    }  
}