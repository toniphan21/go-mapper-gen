package gomappergen

import (
	"fmt"
	"go/types"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toniphan21/go-mapper-gen/internal/setup"
	"github.com/toniphan21/go-mapper-gen/internal/setup/file"
	"github.com/toniphan21/go-mapper-gen/internal/util"
	"golang.org/x/tools/go/packages"
)

type GoldenTestCase struct {
	Name               string
	GoModGoVersion     string
	GoModRequires      map[string]string
	GoModModule        string
	SourceFileContents map[string][]byte
	PklFileContent     []byte
	GoldenFileContent  map[string][]byte
	PrintActual        bool
}

type ConverterTestCase struct {
	Name                         string
	Imports                      map[string]string
	GoModGoVersion               string
	GoModRequires                map[string]string
	GoModModule                  string
	TargetType                   string
	SourceType                   string
	ConverterOption              ConverterOption
	CurrentVarCount              int
	TargetSymbolWithoutFieldName bool
	SourceSymbolWithoutFieldName bool
	ExpectedCanConvert           bool
	ExpectedImports              []string
	ExpectedCode                 []string
	ExpectedNextVarCount         int
	PrintSetUp                   bool
}

type testHelper struct {
}

func (h *testHelper) SetupConfig(t *testing.T, lines ...string) (map[string][]Config, error) {
	dir := setup.SourceCode(t, setup.PklLibFiles(), file.PklDevConfigFile(lines...))

	return ParseConfig(filepath.Join(dir, "dev/config.pkl"))
}

func (h *testHelper) Parse(t *testing.T, files []file.File, requires map[string]string) (Parser, error) {
	dir := setup.SourceCode(t, files)
	if len(requires) > 0 {
		setup.RunGoGet(t, dir, requires)
	}
	return DefaultParser(dir)
}

func (h *testHelper) SetupGoldenTestCaseForPackage(t *testing.T, tc GoldenTestCase, pkgPath string) (Parser, *packages.Package, []Config) {
	parser, configs := h.SetupGoldenTestCase(t, tc)

	config, ok := configs[pkgPath]
	require.True(t, ok, fmt.Sprintf("there is no config for package %v in pkl file", pkgPath))

	for _, pkg := range parser.SourcePackages() {
		if pkg.PkgPath == pkgPath {
			return parser, pkg, config
		}
	}
	panic(fmt.Sprintf("package %v not found", pkgPath))
}

func (h *testHelper) SetupGoldenTestCase(t *testing.T, tc GoldenTestCase) (Parser, map[string][]Config) {
	configs, err := Test.SetupConfig(t, string(tc.PklFileContent))
	require.NoError(t, err, "cannot set up pkl config file")

	goMod := &file.GoMod{
		Module:   tc.GoModModule,
		Version:  tc.GoModGoVersion,
		Requires: tc.GoModRequires,
	}

	var sourceFiles []file.File
	sourceFiles = append(sourceFiles, goMod)

	for filePath, fileContent := range tc.SourceFileContents {
		sourceFiles = append(sourceFiles, file.New(filePath, fileContent))
	}

	parser, err := Test.Parse(t, sourceFiles, goMod.Requires)
	require.NoError(t, err, "cannot parse source files")
	return parser, configs
}

func (h *testHelper) RunGoldenTestCase(t *testing.T, tc GoldenTestCase) {
	parser, pkgConfigs := h.SetupGoldenTestCase(t, tc)

	RegisterAllBuiltinConverters()

	outputs := make(map[string]string)
	fm := defaultFileManager("test")

	for _, pkg := range parser.SourcePackages() {
		pkgPath := pkg.PkgPath
		configs, have := pkgConfigs[pkgPath]
		if !have {
			continue
		}

		err := Generate(parser, fm, pkg, configs)
		require.NoError(t, err, "cannot generate for package %s", pkgPath)
	}

	for fileName, jf := range fm.JenFiles() {
		outputs[fileName] = jf.GoString()
	}
	for fileName, expected := range tc.GoldenFileContent {
		output, have := outputs[fileName]
		if !have {
			assert.Failf(t, "expected file %v not found in outputs", fileName)
		}
		if tc.PrintActual {
			fmt.Println("file: ", fileName)
			fmt.Println(output)
		}
		assert.Equal(t, string(expected), output, "generated file content %v does not match golden file", fileName)
	}
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

	parser, err := DefaultParser(dir)
	require.NoError(t, err)

	var pkg *packages.Package
	for _, v := range parser.SourcePackages() {
		if v.Name == "test" {
			pkg = v
		}
	}
	require.NotNil(t, pkg)

	targetStruct, ok := parser.FindStruct(pkg.PkgPath, "Target")
	require.True(t, ok)

	sourceStruct, ok := parser.FindStruct(pkg.PkgPath, "Source")
	require.True(t, ok)

	targetFieldInfo := targetStruct.Fields["targetField"]
	sourceFieldInfo := sourceStruct.Fields["sourceField"]

	require.NotNil(t, targetFieldInfo)
	require.NotNil(t, sourceFieldInfo)

	require.False(t, isInvalidType(targetFieldInfo.Type), "Target type is invalid, use Imports and GoModRequires to import the package")
	require.False(t, isInvalidType(sourceFieldInfo.Type), "Source type is invalid, use Imports and GoModRequires to import the package")

	if !tc.ExpectedCanConvert {
		var expected = converter.CanConvert(targetFieldInfo.Type, sourceFieldInfo.Type)
		assert.Equal(t, false, expected, "CanConvert should returns false but returns true")
		return
	}

	targetField := "targetField"
	sourceField := "sourceField"
	targetSymbol := Symbol{VarName: "out", Type: targetFieldInfo.Type}
	sourceSymbol := Symbol{VarName: "in", Type: sourceFieldInfo.Type}
	if !tc.TargetSymbolWithoutFieldName {
		targetSymbol.FieldName = &targetField
	} else {
		targetSymbol.VarName = "target"
	}

	if !tc.SourceSymbolWithoutFieldName {
		sourceSymbol.FieldName = &sourceField
	} else {
		sourceSymbol.VarName = "source"
	}

	jf := jen.NewFilePathName(goMod.GetModule(), "test")

	var blocks []jen.Code
	if tc.TargetSymbolWithoutFieldName {
		blocks = append(blocks, jen.Var().Id("target").Add(GeneratorUtil.TypeToJenCode(targetFieldInfo.Type)))
	}
	if tc.SourceSymbolWithoutFieldName {
		blocks = append(blocks, jen.Id("source").Op(":=").Id("in").Dot("sourceField"))
	}

	bc, nvc := converter.Before(jf, targetSymbol, sourceSymbol, tc.CurrentVarCount, tc.ConverterOption)
	if bc != nil {
		blocks = append(blocks, bc)
	}
	assert.Equal(t, tc.ExpectedNextVarCount, nvc)

	c := converter.Assign(jf, targetSymbol, sourceSymbol, tc.ConverterOption)
	if c != nil {
		blocks = append(blocks, c)
	}

	if tc.TargetSymbolWithoutFieldName {
		blocks = append(blocks, jen.Id("out").Dot("targetField").Op("=").Id("target"))
	}

	if len(blocks) == 0 {
		blocks = append(blocks, jen.Comment("empty"))
	}

	jf.Func().Id("convert").Params(
		jen.Id("in").Add(jen.Op("*").Add(GeneratorUtil.TypeToJenCode(sourceStruct.Type))),
		jen.Id("out").Add(jen.Op("*").Add(GeneratorUtil.TypeToJenCode(targetStruct.Type))),
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
	if tc.TargetSymbolWithoutFieldName {
		expected = append(expected, "\tvar target "+tc.TargetType)
	}
	if tc.SourceSymbolWithoutFieldName {
		expected = append(expected, "\tsource := in.sourceField")
	}
	for _, v := range tc.ExpectedCode {
		expected = append(expected, "\t"+v)
	}
	if tc.TargetSymbolWithoutFieldName {
		expected = append(expected, "\tout.targetField = target")
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

func isInvalidType(t types.Type) bool {
	return t == nil || t == types.Typ[types.Invalid]
}
