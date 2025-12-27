package pgtype

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/require"
	gen "github.com/toniphan21/go-mapper-gen"
)

//go:embed testdata/*.md
var goldenMarkdownFiles embed.FS

func TestGolden(t *testing.T) {
	cases := []struct {
		file        string
		printSetup  bool
		printActual bool
		printDiff   bool
	}{
		{file: "testdata/bool.md"},
		{file: "testdata/float4.md"},
		{file: "testdata/float8.md"},
		{file: "testdata/int2.md"},
		{file: "testdata/int4.md"},
		{file: "testdata/int8.md"},
		{file: "testdata/uint32.md"},
		{file: "testdata/uint64.md"},
		{file: "testdata/text.md"},
		{file: "testdata/date.md"},
		{file: "testdata/time.md"},
		{file: "testdata/timestamp.md"},
		{file: "testdata/timestamptz.md"},
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
