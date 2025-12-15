package gomappergen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_mapFieldNames(t *testing.T) {
	cases := []struct {
		name     string
		target   []string
		source   []string
		config   FieldConfig
		expected map[string]string
	}{
		{
			name:     "number of fields are the same, match ignored case",
			config:   FieldConfig{NameMatch: NameMatchIgnoreCase},
			target:   []string{`Id`, `fieldName`},
			source:   []string{`ID`, `FieldName`},
			expected: map[string]string{"Id": "ID", "fieldName": "FieldName"},
		},

		{
			name:     "number of fields are the same, match exact",
			config:   FieldConfig{NameMatch: NameMatchExact},
			target:   []string{`Data`, `A`},
			source:   []string{`A`, `Data`},
			expected: map[string]string{"Data": "Data", "A": "A"},
		},

		{
			name:     "number of fields are the same, match exact, miss some fields",
			config:   FieldConfig{NameMatch: NameMatchExact},
			target:   []string{`FieldName`, `Id`, `Data`, `A`},
			source:   []string{`A`, `ID`, `fieldName`, `Data`},
			expected: map[string]string{"FieldName": "", "Id": "", "Data": "Data", "A": "A"},
		},

		{
			name:     "target is shorter mapped all by name match ignore case",
			config:   FieldConfig{NameMatch: NameMatchIgnoreCase},
			target:   []string{`Id`, `Data`, `A`},
			source:   []string{`A`, `ID`, `fieldName`, `Data`},
			expected: map[string]string{"Id": "ID", "Data": "Data", "A": "A"},
		},

		{
			name: "target is shorter mapped all by manual map",
			config: FieldConfig{NameMatch: NameMatchExact, ManualMap: map[string]string{
				"Id": "UserID", "Data": "UserData", "A": "An",
			}},
			target:   []string{`Id`, `Data`, `A`},
			source:   []string{`An`, `UserID`, `fieldName`, `UserData`},
			expected: map[string]string{"Id": "UserID", "Data": "UserData", "A": "An"},
		},

		{
			name: "manual map to unknown source will be ignored",
			config: FieldConfig{NameMatch: NameMatchExact, ManualMap: map[string]string{
				"Id": "UserID", "Data": "UserData", "A": "An",
			}},
			target:   []string{`Id`, `Data`, `A`},
			source:   []string{`An`, `UserID`, `fieldName`},
			expected: map[string]string{"Id": "UserID", "Data": "", "A": "An"},
		},

		{
			name: "manual map will skip name match",
			config: FieldConfig{NameMatch: NameMatchExact, ManualMap: map[string]string{
				"Id": "UserID", "Data": "",
			}},
			target:   []string{`Id`, `Data`, "A"},
			source:   []string{`Data`, `UserID`, `A`},
			expected: map[string]string{"Id": "UserID", "Data": "", "A": "A"},
		},

		{
			name: "target is longer, match name and manual map with unknown source field",
			config: FieldConfig{NameMatch: NameMatchExact, ManualMap: map[string]string{
				"Id": "UserID", "Data": "UserData",
			}},
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
				Name:        tc.name,
				GoModModule: "github.com/toniphan21/go-mapper-gen/test",
				SourceFileContents: map[string][]byte{
					"code.go": []byte(strings.Join(code, "\n")),
				},
				PklFileContent: []byte(strings.Join([]string{
					`packages {`,
					`	["github.com/toniphan21/go-mapper-gen/test"] {`,
					`		source_pkg = "{CurrentPackage}"`,
					`		structs { ["Target"] { source_struct_name = "Source" } }`,
					`	}`,
					`}`,
				}, "\n")),
			}

			RegisterAllBuiltinConverters()

			parser, _, configs := Test.SetupGoldenTestCaseForPackage(t, gtc, pkgPath)
			require.Equal(t, 1, len(configs))
			require.Equal(t, 1, len(configs[0].Structs))

			targetStruct, _ := parser.FindStruct(pkgPath, "Target")
			sourceStruct, _ := parser.FindStruct(pkgPath, "Source")

			result := mapFieldNames(targetStruct.Fields, sourceStruct.Fields, tc.config)
			assert.Equal(t, tc.expected, result)
		})
	}
}
