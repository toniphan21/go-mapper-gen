package gomappergen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/toniphan21/go-mapper-gen/internal/parse"
)

func Test_mapFields(t *testing.T) {
	cases := []struct {
		name     string
		target   []string
		source   []string
		match    FieldNameMatch
		expected map[string]string
	}{
		{
			name:  "number of fields are the same, match ignored case",
			match: FieldNameMatchIgnoreCase,
			target: []string{
				`Id int`,
				`fieldName string`,
			},
			source: []string{
				`ID int`,
				`fieldName string`,
			},
			expected: map[string]string{
				"Id":        "ID",
				"fieldName": "FieldName",
			},
		},

		{
			name:  "number of fields are the same, match exact",
			match: FieldNameMatchExact,
			target: []string{
				`fieldName string`,
				`Id int`,
				`Data []byte`,
				`A bool`,
			},
			source: []string{
				`A bool`,
				`ID int`,
				`fieldName string`,
				`Data []byte`,
			},
			expected: map[string]string{
				"A":    "A",
				"Data": "Data",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code := []string{
				`package test`,
				``,
				`type Target struct {`,
			}
			for _, c := range tc.target {
				code = append(code, c)
			}
			code = append(code, `}`)
			code = append(code, ``)
			code = append(code, `type Source struct {`)
			for _, c := range tc.source {
				code = append(code, c)
			}
			code = append(code, `}`)
			code = append(code, ``)

			gtc := GoldenTestCase{
				Name:    tc.name,
				Package: "github.com/toniphan21/go-mapper-gen/test",
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
		
			pkg, config := Test.SetupGoldenTestCase(t, gtc)
			require.Equal(t, 1, len(config.Structs))

			targetStruct := parse.Struct(pkg, "Target")
			sourceStruct := parse.Struct(pkg, "Source")
			targetFields := parse.StructFields(pkg, targetStruct)
			sourceFields := parse.StructFields(pkg, sourceStruct)

			result := mapFields(targetFields, sourceFields, tc.match)
			fmt.Println(result)
		})
	}
}
