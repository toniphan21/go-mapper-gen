package gomappergen

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toniphan21/go-mapper-gen/internal/parse"
	"github.com/toniphan21/go-mapper-gen/internal/setup"
	"github.com/toniphan21/go-mapper-gen/internal/setup/file"
	"github.com/toniphan21/go-mapper-gen/internal/util"
	"golang.org/x/tools/go/packages"
)

type GoldenTestCase struct {
	Name               string
	GoModGoVersion     string
	GoModRequires      []string
	GoModModule        string
	Package            string
	SourceFileContents map[string][]byte
	PklFileContent     []byte
	GoldenFileContent  []byte
}

type ConverterTestCase struct {
	Name                 string
	Imports              map[string]string
	GoModGoVersion       string
	GoModRequires        []string
	GoModModule          string
	TargetType           string
	SourceType           string
	CurrentVarCount      int
	ExpectedCanConvert   bool
	ExpectedImports      []string
	ExpectedCode         []string
	ExpectedNextVarCount int
	PrintSetUp           bool
}

type testHelper struct {
}

func (h *testHelper) SetupConfig(t *testing.T, lines ...string) (map[string]Config, error) {
	dir := setup.SourceCode(t, setup.PklLibFiles(), file.PklDevConfigFile(lines...))

	return ParseConfig(filepath.Join(dir, "dev/config.pkl"))
}

func (h *testHelper) Parse(t *testing.T, files []file.File, runGoModTidy bool) ([]*packages.Package, error) {
	dir := setup.SourceCode(t, files)
	if runGoModTidy {
		setup.RunGoModTidy(t, dir)
	}

	pkgs, err := setup.LoadDir(dir)
	if err != nil {
		return nil, err
	}
	return pkgs, nil
}

func (h *testHelper) ParsePackage(t *testing.T, files []file.File, runGoModTidy bool, pkgPath string) (*packages.Package, error) {
	pkgs, err := h.Parse(t, files, runGoModTidy)
	if err != nil {
		return nil, err
	}
	for _, pkg := range pkgs {
		if pkg.PkgPath == pkgPath {
			return pkg, nil
		}
	}
	return nil, errors.New("package not found")
}

func (h *testHelper) SetupGoldenTestCase(t *testing.T, tc GoldenTestCase) (*packages.Package, Config) {
	configs, err := Test.SetupConfig(t, string(tc.PklFileContent))
	require.NoError(t, err, "cannot set up pkl config file")

	config, ok := configs[tc.Package]
	require.True(t, ok, fmt.Sprintf("there is no config for package %v in pkl file", tc.Package))

	goMod := &file.GoMod{
		Module:   tc.GoModModule,
		Version:  tc.GoModGoVersion,
		Requires: tc.GoModRequires,
	}
	if goMod.Module == "" {
		goMod.Module = tc.Package
	}

	var sourceFiles []file.File
	sourceFiles = append(sourceFiles, goMod)

	for filePath, fileContent := range tc.SourceFileContents {
		sourceFiles = append(sourceFiles, file.New(filePath, fileContent))
	}

	pkg, err := Test.ParsePackage(t, sourceFiles, len(goMod.Requires) > 0, tc.Package)
	require.NoError(t, err, fmt.Sprintf("cannot load package %v", tc.Package))

	return pkg, config
}

func (h *testHelper) RunGoldenTestCase(t *testing.T, tc GoldenTestCase) {
	pkg, config := h.SetupGoldenTestCase(t, tc)

	RegisterAllBuiltinConverters()

	jf := MakeJenFile(pkg, config)
	err := Generate(jf, pkg, config)
	require.NoError(t, err, "cannot generate failed")

	output := jf.GoString()
	expected := string(tc.GoldenFileContent)
	assert.Equal(t, expected, output, "generated file does not match golden file")
}

func (h *testHelper) RunConverterTestCase(t *testing.T, tc ConverterTestCase, converter Converter) {
	goMod := &file.GoMod{
		Module:   tc.GoModModule,
		Version:  tc.GoModGoVersion,
		Requires: tc.GoModRequires,
	}

	codeFile := &file.Go{
		Path: "code.go",
	}

	codeFile.Lines = append(codeFile.Lines, "package test")
	codeFile.Lines = append(codeFile.Lines, "")

	if tc.Imports != nil {
		for k, v := range tc.Imports {
			codeFile.Lines = append(codeFile.Lines, fmt.Sprintf(`import %s "%s"`, k, v))
		}
		codeFile.Lines = append(codeFile.Lines, "")
	}

	codeFile.Lines = append(codeFile.Lines, "type Source struct {")
	codeFile.Lines = append(codeFile.Lines, fmt.Sprintf("\tsourceField %v", tc.SourceType))
	codeFile.Lines = append(codeFile.Lines, "}")
	codeFile.Lines = append(codeFile.Lines, "")

	codeFile.Lines = append(codeFile.Lines, "type Target struct {")
	codeFile.Lines = append(codeFile.Lines, fmt.Sprintf("\ttargetField %v", tc.TargetType))
	codeFile.Lines = append(codeFile.Lines, "}")
	codeFile.Lines = append(codeFile.Lines, "")

	dir := setup.SourceCode(t, []file.File{goMod, codeFile})
	if len(tc.GoModRequires) != 0 {
		setup.RunGoModTidy(t, dir)
	}

	pkgs, err := setup.LoadDir(dir)
	require.NoError(t, err)

	var pkg *packages.Package
	for _, v := range pkgs {
		if v.Name == "test" {
			pkg = v
		}
	}
	require.NotNil(t, pkg)

	ts := parse.Struct(pkg, "Target")
	ss := parse.Struct(pkg, "Source")
	require.NotNil(t, ts)
	require.NotNil(t, ss)

	tt := parse.StructType(pkg, "Target")
	st := parse.StructType(pkg, "Source")
	require.NotNil(t, tt)
	require.NotNil(t, st)
	require.False(t, parse.IsInvalidType(tt), "Target type is invalid, use Imports and GoModRequires to import the package")
	require.False(t, parse.IsInvalidType(st), "Source type is invalid, use Imports and GoModRequires to import the package")

	tsf := parse.StructFields(pkg, ts)
	ssf := parse.StructFields(pkg, ss)

	targetType := tsf["targetField"]
	sourceType := ssf["sourceField"]

	require.NotNil(t, targetType)
	require.NotNil(t, sourceType)

	require.False(t, parse.IsInvalidType(targetType), "Target type is invalid, use Imports and GoModRequires to import the package")
	require.False(t, parse.IsInvalidType(sourceType), "Source type is invalid, use Imports and GoModRequires to import the package")

	if !tc.ExpectedCanConvert {
		var expected = converter.CanConvert(targetType, sourceType)
		assert.Equal(t, false, expected, "CanConvert should returns false but returns true")
		return
	}

	targetSymbol := Symbol{VarName: "out", FieldName: "targetField", Type: targetType}
	sourceSymbol := Symbol{VarName: "in", FieldName: "sourceField", Type: sourceType}

	jf := jen.NewFilePathName(goMod.GetModule(), "test")

	var blocks []jen.Code
	bc, nvc := converter.Before(jf, targetSymbol, sourceSymbol, tc.CurrentVarCount)
	if bc != nil {
		blocks = append(blocks, bc)
	}
	assert.Equal(t, tc.ExpectedNextVarCount, nvc)

	c := converter.Assign(jf, targetSymbol, sourceSymbol)
	if c != nil {
		blocks = append(blocks, c)
	}

	if len(blocks) == 0 {
		blocks = append(blocks, jen.Comment("empty"))
	}

	jf.Func().Id("convert").Params(
		jen.Id("in").Add(jen.Op("*").Add(GeneratorUtil.TypeToJenCode(st))),
		jen.Id("out").Add(jen.Op("*").Add(GeneratorUtil.TypeToJenCode(tt))),
	).Block(blocks...).Line()

	expected := []string{
		`package test`,
		``,
	}
	if len(tc.ExpectedImports) != 0 {
		for _, v := range tc.ExpectedImports {
			expected = append(expected, v)
		}
		expected = append(expected, "")
	}

	expected = append(expected, `func convert(in *Source, out *Target) {`)
	for _, v := range tc.ExpectedCode {
		expected = append(expected, "\t"+v)
	}
	expected = append(expected, "}")
	expected = append(expected, "")

	output := jf.GoString()
	assert.Equal(t, strings.Join(expected, "\n"), output)

	if tc.PrintSetUp {
		fmt.Println(util.ColorBlue("--- file: " + goMod.FilePath()))
		fmt.Print(string(goMod.FileContent()))
		fmt.Println(util.ColorBlue("--- end file: " + goMod.FilePath()))
		fmt.Println()

		fmt.Println(util.ColorBlue("--- file: " + codeFile.FilePath()))
		fmt.Print(string(codeFile.FileContent()))
		fmt.Println(util.ColorBlue("--- end file: " + codeFile.FilePath()))
		fmt.Println()

		fmt.Println(util.ColorBlue("--- expected code"))
		fmt.Println(strings.Join(expected, "\n"))
		fmt.Println(util.ColorBlue("--- end expected code"))
		fmt.Println()

		fmt.Println(util.ColorBlue("--- generated code"))
		fmt.Println(output)
		fmt.Println(util.ColorBlue("--- end generated code"))
		fmt.Println()
	}
}

var Test = &testHelper{}
