package grpc

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/require"
	gen "github.com/toniphan21/go-mapper-gen"
)

//go:embed features/*.md
var goldenMarkdownFiles embed.FS

func TestGolden(t *testing.T) {
	cases := []struct {
		file        string
		printSetup  bool
		printActual bool
		printDiff   bool
	}{
		{file: "features/timestamp.md"},
	}

	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			content, err := goldenMarkdownFiles.ReadFile(tc.file)
			require.NoError(t, err)

			mtc := gen.Test.ParseMarkdownTestCases(content)
			for _, v := range mtc {
				gtc := v.ToGoldenTestCase()
				gtc.PrintSetup = tc.printSetup
				gtc.PrintActual = tc.printActual
				gtc.PrintDiff = tc.printDiff
				t.Run(gtc.Name, func(t *testing.T) {
					gen.Test.RunGoldenTestCase(t, gtc, gen.TestWithSetupConverter(func() {
						gen.ClearAllRegisteredConverters()
						gen.RegisterAllBuiltinConverters()
						RegisterConverters()
					}))
				})
			}
		})
	}
}
