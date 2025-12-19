package grpc

import (
	"testing"

	gen "github.com/toniphan21/go-mapper-gen"
)

func ignoreTestGolden(t *testing.T) {
	cases := []gen.GoldenTestCaseFromTestData{
		{
			Name:        "timestamp",
			GoModModule: "github.com/toniphan21/go-mapper-gen/golden",
			GoModRequires: map[string]string{
				"github.com/toniphan21/gmg-lib": "v0.2.0",
			},
			SourceFiles: map[string]string{"code.go": "timestamp.go"},
			PklFile:     "timestamp.pkl",
			GoldenFile:  "timestamp.golden.go",
			GoSumFileContent: gen.Test.FileLines(
				`github.com/toniphan21/gmg-lib v0.2.0 h1:MaJMPtFRPVv8DTHhRdr756xaTgsp7sf1TwIYCztFD6w=`,
				`github.com/toniphan21/gmg-lib v0.2.0/go.mod h1:BZFmDSo4YijtddTGNCaGQmUFuT6QHn+MnQ0pdYKditE=`,
				`google.golang.org/protobuf v1.36.11 h1:fV6ZwhNocDyBLK0dj+fg8ektcVegBBuEolpbTQyBNVE=`,
				`google.golang.org/protobuf v1.36.11/go.mod h1:HTf+CrKn2C3g5S8VImy6tdcUvCska2kB7j23XfzDpco=`,
			),
			PrintDiff: true,
		},
		// ---
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			gen.Test.RunGoldenTestCase(t, tc.ToGoldenTestCase(), gen.TestWithSetupConverter(func() {
				gen.ClearAllRegisteredConverters()
				gen.RegisterAllBuiltinConverters()
				RegisterConverters()
			}))
		})
	}
}
