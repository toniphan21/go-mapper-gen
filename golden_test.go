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
		printActual    bool
	}{
		{
			name:        "same-pkg: basic configurations",
			pkgPath:     "github.com/toniphan21/go-mapper-gen/golden",
			sourceFiles: map[string]string{"code.go": "same-pkg/basic.go"},
			pklFile:     "same-pkg/basic.pkl",
			goldenFile:  "same-pkg/basic.golden.go",
		},

		{
			name:        "same-pkg: multiple mappers configuration",
			pkgPath:     "github.com/toniphan21/go-mapper-gen/golden",
			sourceFiles: map[string]string{"code.go": "same-pkg/multiple-mappers.go"},
			pklFile:     "same-pkg/multiple-mappers.pkl",
			goldenFile:  "same-pkg/multiple-mappers.golden.go",
		},
		// ---
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sourceFiles := make(map[string][]byte)
			for k, v := range c.sourceFiles {
				sourceFiles[k] = file.ContentFromTestData(v)
			}

			tc := GoldenTestCase{
				Name:               c.name,
				PrintActual:        c.printActual,
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
			Test.RunGoldenTestCase(t, tc)
		})
	}
}
