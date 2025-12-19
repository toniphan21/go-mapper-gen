package gomappergen

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed features/*.md testdata/*.md
var goldenMarkdownFiles embed.FS

func TestGolden(t *testing.T) {
	cases := []struct {
		file        string
		printSetup  bool
		printActual bool
		printDiff   bool
	}{
		{file: "features/basic.md"},
		{file: "testdata/import.md"},
		{file: "testdata/placeholder.md"},
	}

	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			content, err := goldenMarkdownFiles.ReadFile(tc.file)
			require.NoError(t, err)

			mtc := Test.ParseMarkdownTestCases(content)
			for _, v := range mtc {
				gtc := v.ToGoldenTestCase()
				gtc.PrintSetup = tc.printSetup
				gtc.PrintActual = tc.printActual
				gtc.PrintDiff = tc.printDiff
				t.Run(gtc.Name, func(t *testing.T) {
					Test.RunGoldenTestCase(t, gtc)
				})
			}
		})
	}
}

func TestGoldenUseTestData(t *testing.T) {
	cases := []GoldenTestCaseFromTestData{
		{
			Name:        "same-pkg: basic configurations",
			GoModModule: "github.com/toniphan21/go-mapper-gen/golden",
			SourceFiles: map[string]string{"code.go": "same-pkg/basic.go"},
			PklFile:     "same-pkg/basic.pkl",
			GoldenFile:  "same-pkg/basic.golden",
		},

		{
			Name:        "same-pkg: multiple mappers configuration",
			GoModModule: "github.com/toniphan21/go-mapper-gen/golden",
			SourceFiles: map[string]string{"code.go": "same-pkg/multiple-mappers.go"},
			PklFile:     "same-pkg/multiple-mappers.pkl",
			GoldenFile:  "same-pkg/multiple-mappers.golden",
		},

		{
			Name:        "same-pkg: use converter functions",
			GoModModule: "github.com/toniphan21/go-mapper-gen/golden",
			SourceFiles: map[string]string{"code.go": "same-pkg/use-converter-functions.go"},
			PklFile:     "same-pkg/use-converter-functions.pkl",
			GoldenFile:  "same-pkg/use-converter-functions.golden",
		},

		{
			Name:        "same-pkg: missing field",
			GoModModule: "github.com/toniphan21/go-mapper-gen/golden",
			SourceFiles: map[string]string{"code.go": "same-pkg/missing-field.go"},
			PklFile:     "same-pkg/missing-field.pkl",
			GoldenFile:  "same-pkg/missing-field.golden",
		},

		{
			Name:        "cross-pkg: basic configurations",
			GoModModule: "github.com/toniphan21/go-mapper-gen/golden",
			SourceFiles: map[string]string{
				"domain/code.go": "cross-pkg/domain/basic.go",
				"grpc/code.go":   "cross-pkg/grpc/basic.go",
			},
			PklFile:        "cross-pkg/grpc/basic.pkl",
			GoldenFile:     "cross-pkg/grpc/basic.golden",
			OutputFileName: "grpc/gen_mapper.go",
		},

		// ---
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			Test.RunGoldenTestCase(t, tc.ToGoldenTestCase())
		})
	}
}
