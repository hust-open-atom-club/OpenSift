package cargo

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantPkg  *langeco.Package
		wantDeps *langeco.Dependencies
		wantErr  error
	}{
		{
			name:     "Standard Cargo.toml",
			filename: "standard.toml",
			wantPkg: &langeco.Package{
				Name:    "basic_project",
				Version: "0.1.0",
				Eco:     parser.CARGO,
			},
			wantDeps: &langeco.Dependencies{
				{
					Name:    "reqwest",
					Version: "0.11",
					Eco:     parser.CARGO,
				},
				{
					Name:    "serde",
					Version: "1.0",
					Eco:     parser.CARGO,
				},
				{
					Name:    "tokio",
					Version: "1.0",
					Eco:     parser.CARGO,
				},
			},
			wantErr: nil,
		},
		{
			name:     "Full featured Cargo.toml",
			filename: "full.toml",
			wantPkg: &langeco.Package{
				Name:    "example_project",
				Version: "0.1.0",
				Eco:     parser.CARGO,
			},
			wantDeps: &langeco.Dependencies{
				{
					Name:    "cc",
					Version: "1.0",
					Eco:     parser.CARGO,
				},
				{
					Name:    "mockito",
					Version: "0.31",
					Eco:     parser.CARGO,
				},
				{
					Name:    "protobuf-codegen",
					Version: "3.0",
					Eco:     parser.CARGO,
				},
				{
					Name:    "reqwest",
					Version: "0.11",
					Eco:     parser.CARGO,
				},
				{
					Name:    "serde",
					Version: "1.0",
					Eco:     parser.CARGO,
				},
				{
					Name:    "test-case",
					Version: "2.0",
					Eco:     parser.CARGO,
				},
				{
					Name:    "tokio",
					Version: "1.0",
					Eco:     parser.CARGO,
				},
			},
			wantErr: nil,
		},
		{
			name:     "Invalid TOML format",
			filename: "invalid.toml",
			wantPkg:  nil,
			wantDeps: nil,
			wantErr:  ErrDecodingFailed,
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

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.wantPkg, gotPkg)

			sortDependencies := func(deps *langeco.Dependencies) {
				if deps != nil {
					sort.Slice(*deps, func(i, j int) bool {
						return (*deps)[i].Name < (*deps)[j].Name
					})
				}
			}

			sortDependencies(gotDeps)
			sortDependencies(tt.wantDeps)
			require.Equal(t, tt.wantDeps, gotDeps)
		})
	}
}
