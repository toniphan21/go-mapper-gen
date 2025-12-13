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
		goModRequires  []string
		goModModule    string
		sourceFiles    map[string]string
		pklFile        string
		goldenFile     string
	}{
		{
			name:        "same-pkg: identical",
			pkgPath:     "github.com/toniphan21/go-mapper-gen/golden",
			sourceFiles: map[string]string{"code.go": "same-pkg/identical.go"},
			pklFile:     "same-pkg/identical.pkl",
			goldenFile:  "same-pkg/identical.golden",
		},
		// ---
	}

	var tcs []GoldenTestCase
	for _, c := range cases {
		sourceFiles := make(map[string][]byte)
		for k, v := range c.sourceFiles {
			sourceFiles[k] = file.ContentFromTestData(v)
		}

		tc := GoldenTestCase{
			Name:               c.name,
			GoModGoVersion:     c.goModGoVersion,
			GoModRequires:      c.goModRequires,
			GoModModule:        c.goModModule,
			Package:            c.pkgPath,
			SourceFileContents: sourceFiles,
			PklFileContent:     file.ContentFromTestData(c.pklFile),
			GoldenFileContent: map[string][]byte{
				Default.Output.FileName: file.ContentFromTestData(c.goldenFile),
			},
		}

		tcs = append(tcs, tc)
	}

	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {
			Test.RunGoldenTestCase(t, tc)
		})
	}
}
