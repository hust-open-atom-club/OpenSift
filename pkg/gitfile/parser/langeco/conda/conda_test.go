package conda

import (
	"os"
	"path/filepath"
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
			name:     "Standard conda package",
			filename: "standard.json",
			wantPkg: &langeco.Package{
				Name:    "numpy",
				Version: "1.21.5",
				Eco:     parser.CONDA,
			},
			wantDeps: &langeco.Dependencies{},
			wantErr:  nil,
		},
		{
			name:     "Package with full metadata",
			filename: "full_metadata.json",
			wantPkg: &langeco.Package{
				Name:    "scipy",
				Version: "1.7.3",
				Eco:     parser.CONDA,
			},
			wantDeps: &langeco.Dependencies{},
			wantErr:  nil,
		},
		{
			name:     "Package with build info and dependencies",
			filename: "with_build.json",
			wantPkg: &langeco.Package{
				Name:    "pandas",
				Version: "1.3.5",
				Eco:     parser.CONDA,
			},
			wantDeps: &langeco.Dependencies{},
			wantErr:  nil,
		},
		{
			name:     "Package with license",
			filename: "with_license.json",
			wantPkg: &langeco.Package{
				Name:    "matplotlib",
				Version: "3.5.1",
				Eco:     parser.CONDA,
			},
			wantDeps: &langeco.Dependencies{},
			wantErr:  nil,
		},
		{
			name:     "Invalid JSON format",
			filename: "invalid.json",
			wantPkg:  nil,
			wantDeps: nil,
			wantErr:  ErrDecodingFailed,
		},
		{
			name:     "Missing required fields",
			filename: "missing_fields.json",
			wantPkg:  nil,
			wantDeps: nil,
			wantErr:  ErrParsingFailed,
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
			require.Equal(t, tt.wantDeps, gotDeps)
		})
	}
}
