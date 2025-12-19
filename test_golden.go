package gomappergen

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toniphan21/go-mapper-gen/internal/setup"
	"github.com/toniphan21/go-mapper-gen/internal/setup/file"
	"github.com/toniphan21/go-mapper-gen/internal/util"
	"golang.org/x/tools/go/packages"
)

type GoldenTestCase struct {
	Name              string
	GoModFileContent  []byte
	GoSumFileContent  []byte
	PklDevFileContent []byte
	SourceFiles       map[string][]byte
	GoldenFiles       map[string][]byte
	PrintSetup        bool
	PrintActual       bool
	PrintDiff         bool
}

type GoldenTestCaseFromTestData struct {
	Name             string
	PkgPath          string
	GoModGoVersion   string
	GoModRequires    map[string]string
	GoModModule      string
	GoSumFileContent []byte
	SourceFiles      map[string]string
	PklFile          string
	GoldenFile       string
	OutputFileName   string
	PrintSetup       bool
	PrintActual      bool
	PrintDiff        bool
}

func (g *GoldenTestCaseFromTestData) ToGoldenTestCase() GoldenTestCase {
	sourceFiles := make(map[string][]byte)
	for k, v := range g.SourceFiles {
		sourceFiles[k] = Test.ContentFromTestData(v)
	}

	outputFileName := Default.Output.FileName
	if g.OutputFileName != "" {
		outputFileName = g.OutputFileName
	}

	goMod := &file.GoMod{
		Module:   g.GoModModule,
		Version:  g.GoModGoVersion,
		Requires: g.GoModRequires,
	}

	return GoldenTestCase{
		Name:              g.Name,
		PrintSetup:        g.PrintSetup,
		PrintActual:       g.PrintActual,
		PrintDiff:         g.PrintDiff,
		GoModFileContent:  goMod.FileContent(),
		GoSumFileContent:  g.GoSumFileContent,
		SourceFiles:       sourceFiles,
		PklDevFileContent: Test.ContentFromTestData(g.PklFile),
		GoldenFiles: map[string][]byte{
			outputFileName: Test.ContentFromTestData(g.GoldenFile),
		},
	}
}

type GoldenTestCaseRunOptions struct {
	SetupConverter func()
}

type GoldenTestCaseRunOptionsFunc func(opts *GoldenTestCaseRunOptions)

func TestWithSetupConverter(fn func()) GoldenTestCaseRunOptionsFunc {
	return func(opts *GoldenTestCaseRunOptions) {
		opts.SetupConverter = fn
	}
}

type goldenTest struct {
}

func (h *goldenTest) configFromPklDevFileContent(t *testing.T, content []byte) (*Config, error) {
	dir := setup.SourceCode(t, setup.PklLibFiles(), file.PklDevConfigFile(string(content)))

	return ParseConfig(filepath.Join(dir, "dev/config.pkl"))
}

func (h *goldenTest) SetupGoldenTestCaseForPackage(t *testing.T, tc GoldenTestCase, pkgPath string) (Parser, *packages.Package, []PackageConfig) {
	parser, config := h.SetupGoldenTestCase(t, tc)

	pkgCf, ok := config.Packages[pkgPath]
	require.True(t, ok, fmt.Sprintf("there is no config for package %v in pkl file", pkgPath))

	for _, pkg := range parser.SourcePackages() {
		if pkg.PkgPath == pkgPath {
			return parser, pkg, pkgCf
		}
	}
	panic(fmt.Sprintf("package %v not found", pkgPath))
}

func (h *goldenTest) SetupGoldenTestCase(t *testing.T, tc GoldenTestCase) (Parser, *Config) {
	// TODO: handle pkl file which is used in production with pkl package
	config, err := h.configFromPklDevFileContent(t, tc.PklDevFileContent)
	require.NoError(t, err, "cannot set up pkl config file")

	var sourceFiles []file.File
	if tc.GoModFileContent != nil {
		sourceFiles = append(sourceFiles, file.New("go.mod", tc.GoModFileContent))
	}

	if tc.GoSumFileContent != nil {
		goSum := file.New("go.sum", tc.GoSumFileContent)
		sourceFiles = append(sourceFiles, goSum)
	}

	for filePath, fileContent := range tc.SourceFiles {
		sourceFiles = append(sourceFiles, file.New(filePath, fileContent))
	}

	if tc.PrintSetup {
		for _, f := range sourceFiles {
			util.PrintFile(f.FilePath(), f.FileContent())
		}
		for p, c := range tc.GoldenFiles {
			util.PrintFile("golden-file: "+p, c)
		}
	}

	parser, err := Test.Parse(t, sourceFiles)
	require.NoError(t, err, "cannot parse source files")
	return parser, config
}

func (h *goldenTest) RunGoldenTestCase(t *testing.T, tc GoldenTestCase, opts ...GoldenTestCaseRunOptionsFunc) {
	parser, config := h.SetupGoldenTestCase(t, tc)

	o := GoldenTestCaseRunOptions{
		SetupConverter: func() {
			ClearAllRegisteredConverters()
			RegisterBuiltinConverters(config.BuiltInConverters)
			InitAllRegisteredConverters(parser, *config)
		},
	}
	for _, opt := range opts {
		opt(&o)
	}

	o.SetupConverter()

	outputs := make(map[string]string)
	UseTestVersion()
	fm := DefaultFileManager()

	generator := New(parser, WithFileManager(fm), WithLogger(NewNoopLogger()))

	for _, pkg := range parser.SourcePackages() {
		pkgPath := pkg.PkgPath
		configs, have := config.Packages[pkgPath]
		if !have {
			continue
		}

		err := generator.Generate(pkg, configs)
		require.NoError(t, err, "cannot generate for package %s", pkgPath)
	}

	for fileName, jf := range fm.JenFiles() {
		outputs[fileName] = jf.GoString()
	}

	for fileName, expected := range tc.GoldenFiles {
		output, have := outputs[fileName]
		if !have {
			assert.Failf(t, "output file not found", "expected file %v not found in outputs", fileName)
			return
		}
		if tc.PrintActual {
			util.PrintGeneratedFile(fileName, []byte(output))
		}

		if tc.PrintDiff {
			fmt.Println(util.ColorGreen(fileName))
			util.PrintDiff("expected", expected, "output", []byte(output))
		}
		assert.Equal(t, string(expected), output, "generated file content %v does not match golden file", fileName)
	}
}
