package sum

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
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
			name:     "Standard go.sum",
			filename: "standard.sum",
			wantPkg: &langeco.Package{
				Eco: parser.GO,
			},
			wantDeps: &langeco.Dependencies{
				{
					Name:    "github.com/bytedance/sonic",
					Version: "v1.13.2",
					Eco:     parser.GO,
				},
				{
					Name:    "github.com/bytedance/sonic/loader",
					Version: "v0.2.4",
					Eco:     parser.GO,
				},
				{
					Name:    "github.com/cloudwego/base64x",
					Version: "v0.1.5",
					Eco:     parser.GO,
				},
				{
					Name:    "golang.org/x/net",
					Version: "v0.41.0",
					Eco:     parser.GO,
				},
				{
					Name:    "github.com/leodido/go-urn",
					Version: "v1.4.0",
					Eco:     parser.GO,
				},
				{
					Name:    "github.com/mattn/go-isatty",
					Version: "v0.0.20",
					Eco:     parser.GO,
				},
				{
					Name:    "github.com/modern-go/concurrent",
					Version: "v0.0.0-20180228061459-e0a39a4cb421",
					Eco:     parser.GO,
				},
			},
			wantErr: false,
		},
		{
			name:     "Multiple versions of same package",
			filename: "multiple_versions.sum",
			wantPkg: &langeco.Package{
				Eco: parser.GO,
			},
			wantDeps: &langeco.Dependencies{
				{
					Name:    "github.com/multi/pkg",
					Version: "v1.1.0",
					Eco:     parser.GO,
				},
				{
					Name:    "golang.org/x/net",
					Version: "v0.41.2",
					Eco:     parser.GO,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("TestData", tt.filename)
			data, err := os.ReadFile(path)
			require.NoError(t, err, "Failed to read test file")

			gotPkg, gotDeps, err := Parse(string(data))

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.wantPkg.Eco, gotPkg.Eco)

			require.Len(t, *gotDeps, len(*tt.wantDeps))
			for i := range *tt.wantDeps {
				require.Equal(t, (*tt.wantDeps)[i].Name, (*gotDeps)[i].Name)
				require.Equal(t, (*tt.wantDeps)[i].Version, (*gotDeps)[i].Version)
			}
		})
	}
}
