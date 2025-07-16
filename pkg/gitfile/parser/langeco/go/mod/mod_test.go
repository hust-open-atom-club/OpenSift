package mod

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
			name:     "Standard go.mod",
			filename: "standard.mod",
			wantPkg: &langeco.Package{
				Name:    "github.com/gin-gonic/gin",
				Version: "",
				Eco:     parser.GO,
			},
			wantDeps: &langeco.Dependencies{
				{
					Name:    "github.com/bytedance/sonic",
					Version: "v1.13.2",
					Eco:     parser.GO,
				},
				{
					Name:    "github.com/gin-contrib/sse",
					Version: "v1.1.0",
					Eco:     parser.GO,
				},
				{
					Name:    "github.com/go-playground/validator/v10",
					Version: "v10.26.0",
					Eco:     parser.GO,
				},
				{
					Name:    "github.com/goccy/go-json",
					Version: "v0.10.2",
					Eco:     parser.GO,
				},
				{
					Name:    "github.com/goccy/go-yaml",
					Version: "v1.18.0",
					Eco:     parser.GO,
				},
				{
					Name:    "github.com/json-iterator/go",
					Version: "v1.1.12",
					Eco:     parser.GO,
				},
				{
					Name:    "github.com/mattn/go-isatty",
					Version: "v0.0.20",
					Eco:     parser.GO,
				},
				{
					Name:    "github.com/modern-go/reflect2",
					Version: "v1.0.2",
					Eco:     parser.GO,
				},
				{
					Name:    "github.com/pelletier/go-toml/v2",
					Version: "v2.2.4",
					Eco:     parser.GO,
				},
				{
					Name:    "github.com/quic-go/quic-go",
					Version: "v0.52.0",
					Eco:     parser.GO,
				},
				{
					Name:    "github.com/stretchr/testify",
					Version: "v1.10.0",
					Eco:     parser.GO,
				},
				{
					Name:    "github.com/ugorji/go/codec",
					Version: "v1.2.12",
					Eco:     parser.GO,
				},
				{
					Name:    "golang.org/x/net",
					Version: "v0.41.0",
					Eco:     parser.GO,
				},
				{
					Name:    "google.golang.org/protobuf",
					Version: "v1.36.6",
					Eco:     parser.GO,
				},
			},
			wantErr: false,
		},
		{
			name:     "Invalid go.mod",
			filename: "invalid.mod",
			wantPkg:  nil,
			wantDeps: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("TestData", tt.filename)
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

			require.Equal(t, *tt.wantPkg, *gotPkg)
			require.Len(t, *gotDeps, len(*tt.wantDeps))

			for i := range *tt.wantDeps {
				require.Equal(t, (*tt.wantDeps)[i], (*gotDeps)[i])
			}
		})
	}
}
