package gomappergen

import (
	"context"
	"fmt"
	"go/types"
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

type ConverterTestCase struct {
	Name                         string
	AdditionalCode               []string
	Config                       *Config
	Imports                      map[string]string
	GoModGoVersion               string
	GoModRequires                map[string]string
	GoModModule                  string
	TargetType                   string
	SourceType                   string
	EmitTraceComments            bool
	TargetSymbolWithoutFieldName bool
	SourceSymbolWithoutFieldName bool
	TargetSymbolMetadata         SymbolMetadata
	SourceSymbolMetadata         SymbolMetadata
	ExpectedCanConvert           bool
	ExpectedImports              []string
	ExpectedCode                 []string
	PrintSetUp                   bool
}

type converterTest struct {
}

func (h *converterTest) RunConverterTestCase(t *testing.T, tc ConverterTestCase, converter Converter) {
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

	for _, l := range tc.AdditionalCode {
		codeFile.Lines = append(codeFile.Lines, l)
	}

	dir := setup.SourceCode(t, []file.File{goMod, codeFile})
	if len(tc.GoModRequires) != 0 {
		setup.RunGoModTidy(t, dir)
	}

	parser, err := DefaultParser(dir)
	require.NoError(t, err)

	if tc.Config != nil {
		converter.Init(parser, *tc.Config)
	}

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

	require.False(t, h.isInvalidType(targetFieldInfo.Type), "Target type is invalid, use Imports and GoModRequires to import the package")
	require.False(t, h.isInvalidType(sourceFieldInfo.Type), "Source type is invalid, use Imports and GoModRequires to import the package")

	jf := jen.NewFilePathName(goMod.GetModule(), "test")
	ctx := &converterContext{
		Context:           context.Background(),
		lookupContext:     newLookupContext(),
		jenFile:           jf,
		parser:            parser,
		logger:            NewNoopLogger(),
		emitTraceComments: tc.EmitTraceComments,
	}

	if tc.PrintSetUp {
		util.PrintFile(goMod.FilePath(), goMod.FileContent())
		util.PrintFile(codeFile.FilePath(), codeFile.FileContent())
	}

	if !tc.ExpectedCanConvert {
		var expected = converter.CanConvert(ctx, targetFieldInfo.Type, sourceFieldInfo.Type)
		assert.Equal(t, false, expected, "CanConvert should returns false but returns true")
		return
	}

	targetField := "targetField"
	sourceField := "sourceField"
	targetSymbol := Symbol{VarName: "out", Type: targetFieldInfo.Type, Metadata: tc.TargetSymbolMetadata}
	sourceSymbol := Symbol{VarName: "in", Type: sourceFieldInfo.Type, Metadata: tc.SourceSymbolMetadata}
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

	var blocks []jen.Code
	if tc.TargetSymbolWithoutFieldName {
		blocks = append(blocks, jen.Var().Id("target").Add(GeneratorUtil.TypeToJenCode(targetFieldInfo.Type)))
	}
	if tc.SourceSymbolWithoutFieldName {
		blocks = append(blocks, jen.Id("source").Op(":=").Id("in").Dot("sourceField"))
	}

	c := converter.ConvertField(ctx, targetSymbol, sourceSymbol)
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
		if v != "" {
			expected = append(expected, "\t"+v)
		} else {
			expected = append(expected, v)
		}
	}
	if tc.TargetSymbolWithoutFieldName {
		expected = append(expected, "\tout.targetField = target")
	}
	expected = append(expected, "}")
	expected = append(expected, "")

	output := jf.GoString()
	assert.Equal(t, strings.Join(expected, "\n"), output)

	if tc.PrintSetUp {
		util.PrintDiff("expected", []byte(strings.Join(expected, "\n")), "generated", []byte(output))
	}
}

func (h *converterTest) isInvalidType(t types.Type) bool {
	return t == nil || t == types.Typ[types.Invalid]
}
