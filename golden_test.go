package gomappergen

import (
	"testing"
)

func TestGolden(t *testing.T) {
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
			PklFile:    "cross-pkg/grpc/basic.pkl",
			GoldenFile: "cross-pkg/grpc/basic.golden",
		},

		{
			Name: "use-import: import as source",
			GoModRequires: map[string]string{
				"github.com/toniphan21/gmg-lib": "v0.1.0",
			},
			GoModModule: "github.com/toniphan21/go-mapper-gen/golden",
			SourceFiles: map[string]string{
				"code.go": "use-import/import-as-source.go",
			},
			PklFile:    "use-import/import-as-source.pkl",
			GoldenFile: "use-import/import-as-source.golden",
		},
		// ---
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			Test.RunGoldenTestCase(t, tc.ToGoldenTestCase())
		})
	}
}
