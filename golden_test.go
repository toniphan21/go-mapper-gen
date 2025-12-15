package gomappergen

import (
	"testing"

	"github.com/toniphan21/go-mapper-gen/internal/setup/file"
)

func TestGolden(t *testing.T) {
	cases := []struct {
		name           string
		pkgPath        string
		goModGoVersion string
		goModRequires  map[string]string
		goModModule    string
		sourceFiles    map[string]string
		pklFile        string
		goldenFile     string
		outputFileName string
		printActual    bool
	}{
		{
			name:        "same-pkg: basic configurations",
			goModModule: "github.com/toniphan21/go-mapper-gen/golden",
			sourceFiles: map[string]string{"code.go": "same-pkg/basic.go"},
			pklFile:     "same-pkg/basic.pkl",
			goldenFile:  "same-pkg/basic.golden",
		},

		{
			name:        "same-pkg: multiple mappers configuration",
			goModModule: "github.com/toniphan21/go-mapper-gen/golden",
			sourceFiles: map[string]string{"code.go": "same-pkg/multiple-mappers.go"},
			pklFile:     "same-pkg/multiple-mappers.pkl",
			goldenFile:  "same-pkg/multiple-mappers.golden",
		},

		{
			name:        "cross-pkg: basic configurations",
			goModModule: "github.com/toniphan21/go-mapper-gen/golden",
			sourceFiles: map[string]string{
				"domain/code.go": "cross-pkg/domain/basic.go",
				"grpc/code.go":   "cross-pkg/grpc/basic.go",
			},
			pklFile:    "cross-pkg/grpc/basic.pkl",
			goldenFile: "cross-pkg/grpc/basic.golden",
		},

		{
			name: "use-import: import as source",
			goModRequires: map[string]string{
				"github.com/toniphan21/gmg-lib": "v0.1.0",
			},
			goModModule: "github.com/toniphan21/go-mapper-gen/golden",
			sourceFiles: map[string]string{
				"code.go": "use-import/import-as-source.go",
			},
			pklFile:    "use-import/import-as-source.pkl",
			goldenFile: "use-import/import-as-source.golden",
		},
		// ---
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sourceFiles := make(map[string][]byte)
			for k, v := range tc.sourceFiles {
				sourceFiles[k] = file.ContentFromTestData(v)
			}

			outputFileName := Default.Output.FileName
			if tc.outputFileName != "" {
				outputFileName = tc.outputFileName
			}
			tc := GoldenTestCase{
				Name:               tc.name,
				PrintActual:        tc.printActual,
				GoModGoVersion:     tc.goModGoVersion,
				GoModRequires:      tc.goModRequires,
				GoModModule:        tc.goModModule,
				SourceFileContents: sourceFiles,
				PklFileContent:     file.ContentFromTestData(tc.pklFile),
				GoldenFileContent: map[string][]byte{
					outputFileName: file.ContentFromTestData(tc.goldenFile),
				},
			}
			Test.RunGoldenTestCase(t, tc)
		})
	}
}
