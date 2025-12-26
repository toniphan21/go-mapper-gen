package gomappergen

import (
	"go/types"
	"strings"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_mapFieldNames(t *testing.T) {
	cases := []struct {
		name     string
		target   []string
		source   []string
		config   FieldConfig
		samePkg  bool
		expected map[string]string
	}{
		{
			name:     "number of fields are the same, match ignored case - same pkg",
			config:   FieldConfig{NameMatch: NameMatchIgnoreCase},
			samePkg:  true,
			target:   []string{`Id`, `fieldName`},
			source:   []string{`ID`, `FieldName`},
			expected: map[string]string{"Id": "ID", "fieldName": "FieldName"},
		},

		{
			name:     "number of fields are the same, match ignored case - skip target export - different pkg",
			config:   FieldConfig{NameMatch: NameMatchIgnoreCase},
			samePkg:  false,
			target:   []string{`Id`, `fieldName`},
			source:   []string{`ID`, `FieldName`},
			expected: map[string]string{"Id": "ID"},
		},

		{
			name:     "number of fields are the same, match ignored case - skip source export - different pkg",
			config:   FieldConfig{NameMatch: NameMatchIgnoreCase},
			samePkg:  false,
			target:   []string{`Id`, `FieldName`},
			source:   []string{`ID`, `fieldName`},
			expected: map[string]string{"Id": "ID", "FieldName": ""},
		},

		{
			name:     "number of fields are the same, match exact",
			config:   FieldConfig{NameMatch: NameMatchExact},
			samePkg:  true,
			target:   []string{`Data`, `A`},
			source:   []string{`A`, `Data`},
			expected: map[string]string{"Data": "Data", "A": "A"},
		},

		{
			name:     "number of fields are the same, match exact, miss some fields",
			config:   FieldConfig{NameMatch: NameMatchExact},
			samePkg:  true,
			target:   []string{`FieldName`, `Id`, `Data`, `A`},
			source:   []string{`A`, `ID`, `fieldName`, `Data`},
			expected: map[string]string{"FieldName": "", "Id": "", "Data": "Data", "A": "A"},
		},

		{
			name:     "target is shorter mapped all by name match ignore case",
			config:   FieldConfig{NameMatch: NameMatchIgnoreCase},
			samePkg:  true,
			target:   []string{`Id`, `Data`, `A`},
			source:   []string{`A`, `ID`, `fieldName`, `Data`},
			expected: map[string]string{"Id": "ID", "Data": "Data", "A": "A"},
		},

		{
			name: "target is shorter mapped all by manual map",
			config: FieldConfig{NameMatch: NameMatchExact, ManualMap: map[string]string{
				"Id": "UserID", "Data": "UserData", "A": "An",
			}},
			samePkg:  true,
			target:   []string{`Id`, `Data`, `A`},
			source:   []string{`An`, `UserID`, `fieldName`, `UserData`},
			expected: map[string]string{"Id": "UserID", "Data": "UserData", "A": "An"},
		},

		{
			name: "manual mapped can map unexported field if same pkg",
			config: FieldConfig{NameMatch: NameMatchExact, ManualMap: map[string]string{
				"Id": "UserID", "Data": "UserData", "A": "a",
			}},
			samePkg:  true,
			target:   []string{`Id`, `Data`, `A`},
			source:   []string{`a`, `UserID`, `fieldName`, `UserData`},
			expected: map[string]string{"Id": "UserID", "Data": "UserData", "A": "a"},
		},

		{
			name: "manual mapped cannot map unexported field if different pkg",
			config: FieldConfig{NameMatch: NameMatchExact, ManualMap: map[string]string{
				"Id": "UserID", "Data": "UserData", "A": "a",
			}},
			samePkg:  false,
			target:   []string{`Id`, `Data`, `A`},
			source:   []string{`a`, `UserID`, `fieldName`, `UserData`},
			expected: map[string]string{"Id": "UserID", "Data": "UserData", "A": ""},
		},

		{
			name: "manual map to unknown source will be ignored",
			config: FieldConfig{NameMatch: NameMatchExact, ManualMap: map[string]string{
				"Id": "UserID", "Data": "UserData", "A": "An",
			}},
			samePkg:  true,
			target:   []string{`Id`, `Data`, `A`},
			source:   []string{`An`, `UserID`, `fieldName`},
			expected: map[string]string{"Id": "UserID", "Data": "", "A": "An"},
		},

		{
			name: "manual map will skip name match",
			config: FieldConfig{NameMatch: NameMatchExact, ManualMap: map[string]string{
				"Id": "UserID", "Data": "",
			}},
			samePkg:  true,
			target:   []string{`Id`, `Data`, "A"},
			source:   []string{`Data`, `UserID`, `A`},
			expected: map[string]string{"Id": "UserID", "Data": "", "A": "A"},
		},

		{
			name: "target is longer, match name and manual map with unknown source field",
			config: FieldConfig{NameMatch: NameMatchExact, ManualMap: map[string]string{
				"Id": "UserID", "Data": "UserData",
			}},
			samePkg:  true,
			target:   []string{`Id`, `Data`, "Profile", "Password", "Email"},
			source:   []string{`Data`, `UserID`, `profile`, "Password"},
			expected: map[string]string{"Id": "UserID", "Data": "", "Profile": "", "Password": "Password", "Email": ""},
		},
		// ---
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code := []string{
				`package test`,
				``,
				`type Target struct {`,
			}
			for _, c := range tc.target {
				code = append(code, c+" string")
			}
			code = append(code, `}`)
			code = append(code, ``)
			code = append(code, `type Source struct {`)
			for _, c := range tc.source {
				code = append(code, c+" string")
			}
			code = append(code, `}`)
			code = append(code, ``)

			pkgPath := "github.com/toniphan21/go-mapper-gen/test"
			gtc := GoldenTestCase{
				Name:             tc.name,
				GoModFileContent: Test.MakeGoModFileContent("github.com/toniphan21/go-mapper-gen/test", nil, nil),
				SourceFiles: map[string][]byte{
					"code.go": []byte(strings.Join(code, "\n")),
				},
				PklDevFileContent: []byte(strings.Join([]string{
					`packages {`,
					`	["github.com/toniphan21/go-mapper-gen/test"] {`,
					`		source_pkg = "{CurrentPackage}"`,
					`		structs { ["Target"] { source_struct_name = "Source" } }`,
					`	}`,
					`}`,
				}, "\n")),
			}

			ClearAllRegisteredConverters()
			RegisterAllBuiltinConverters()

			parser, _, configs := Test.SetupGoldenTestCaseForPackage(t, gtc, pkgPath)
			require.Equal(t, 1, len(configs))
			require.Equal(t, 1, len(configs[0].Structs))

			targetStruct, _ := parser.FindStruct(pkgPath, "Target")
			sourceStruct, _ := parser.FindStruct(pkgPath, "Source")

			result := mapFieldNames(targetStruct.Fields, sourceStruct.Fields, tc.config, tc.samePkg)
			assert.Equal(t, tc.expected, result)
		})
	}
}

type dummyConverter struct {
}

func (d *dummyConverter) Init(parser Parser, config Config) {}

func (d *dummyConverter) Info() ConverterInfo {
	return ConverterInfo{Name: "dummy"}
}

func (d *dummyConverter) CanConvert(ctx LookupContext, targetType, sourceType types.Type) bool {
	return true
}

func (d *dummyConverter) ConvertField(ctx ConverterContext, target, source Symbol, opts ConverterOption) jen.Code {
	return nil
}

var _ Converter = (*dummyConverter)(nil)

func Test_converterReturnNilIsConsiderUnconvertible(t *testing.T) {
	tc := GoldenTestCase{
		Name:             "converter returns nil is considered unconvertible",
		GoModFileContent: Test.MakeGoModFileContent("github.com/toniphan21/go-mapper-gen/test", nil, nil),
		SourceFiles: map[string][]byte{
			"code.go": Test.FileLines(
				`package test`,
				``,
				`type Target struct {`,
				`	ID int`,
				`	Name string`,
				`}`,
				``,
				`type Source struct {`,
				`	ID int`,
				`	Name string`,
				`}`,
			),
		},
		PklDevFileContent: Test.FileLines(
			`packages {`,
			`	["github.com/toniphan21/go-mapper-gen/test"] {`,
			`		source_pkg = "{CurrentPackage}"`,
			``,
			`		structs {`,
			`			["Target"] { source_struct_name = "Source" }`,
			`		}`,
			`	}`,
			`}`,
		),
		GoldenFiles: map[string][]byte{
			Default.Output.FileName: []byte(`// Code generated by github.com/toniphan21/go-mapper-gen - test, DO NOT EDIT.

package test

type iMapper interface {
	// ToTarget converts a Source value into a Target value.
	ToTarget(in Source) Target

	// FromTarget converts a Target value into a Source value.
	FromTarget(in Target) Source
}

type iMapperDecorator interface {
	decorateToTarget(in *Source, out *Target)

	decorateFromTarget(in *Target, out *Source)
}

func new_iMapper(decorator iMapperDecorator) iMapper {
	return &iMapperImpl{decorator: decorator}
}

type iMapperImpl struct {
	decorator iMapperDecorator
}

func (m *iMapperImpl) ToTarget(in Source) Target {
	var out Target

	if m.decorator != nil {
		m.decorator.decorateToTarget(&in, &out)
	}

	return out
}

func (m *iMapperImpl) FromTarget(in Target) Source {
	var out Source

	if m.decorator != nil {
		m.decorator.decorateFromTarget(&in, &out)
	}

	return out
}

type iMapperDecoratorNoOp struct{}

func (d *iMapperDecoratorNoOp) decorateToTarget(in *Source, out *Target) {
	// Fields that could not be converted (no suitable converter found):
	// out.ID =
	// out.Name =
}

func (d *iMapperDecoratorNoOp) decorateFromTarget(in *Target, out *Source) {
	// Fields that could not be converted (no suitable converter found):
	// out.ID =
	// out.Name =
}

var _ iMapper = (*iMapperImpl)(nil)
var _ iMapperDecorator = (*iMapperDecoratorNoOp)(nil)
`),
		},
	}

	Test.RunGoldenTestCase(t, tc, TestWithSetupConverter(func() {
		ClearAllRegisteredConverters()
		RegisterConverter(&dummyConverter{})
	}))
}
