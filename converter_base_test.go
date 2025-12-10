package gomappergen

import (
	"testing"
)

func Test_typeToPointerConverter(t *testing.T) {
	cases := []ConverterTestCase{
		{Name: "cannot convert string to *bool", SourceType: "string", TargetType: "*bool"},
		{Name: "cannot convert int to *int32", SourceType: "int", TargetType: "*int32"},
		{
			Name:       "cannot convert Interface to *Interface",
			Imports:    map[string]string{"io": "io"},
			SourceType: "io.Reader",
			TargetType: "*io.Reader",
		},

		{
			Name:               "bool to *bool",
			SourceType:         "bool",
			TargetType:         "*bool",
			ExpectedCanConvert: true,
			ExpectedCode:       []string{"out.targetField = &in.sourceField"},
		},
		{
			Name:               "pgtype.Text to *pgtype.Text",
			Imports:            map[string]string{"pgtype": "github.com/jackc/pgx/v5/pgtype"},
			GoModRequires:      []string{"github.com/jackc/pgx/v5 v5.5.4"},
			SourceType:         "pgtype.Text",
			TargetType:         "*pgtype.Text",
			ExpectedCanConvert: true,
			ExpectedCode:       []string{"out.targetField = &in.sourceField"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			converter := &typeToPointerConverter{}
			Test.RunConverterTestCase(t, tc, converter)
		})
	}
}

func Test_pointerToTypeConverter(t *testing.T) {
	cases := []ConverterTestCase{
		{Name: "cannot convert *string to bool", SourceType: "*string", TargetType: "bool"},
		{Name: "cannot convert *int to int32", SourceType: "*int", TargetType: "int32"},
		{Name: "cannot convert *int to int64", SourceType: "*int", TargetType: "int64"},
		{
			Name:       "cannot convert *Interface to Interface",
			Imports:    map[string]string{"io": "io"},
			SourceType: "*io.Reader",
			TargetType: "io.Reader",
		},

		{
			Name:               "*bool to bool",
			SourceType:         "*bool",
			TargetType:         "bool",
			ExpectedCanConvert: true,
			ExpectedCode: []string{
				`if in.sourceField == nil {`,
				`	var v bool`,
				`	out.targetField = v`,
				`} else {`,
				`	out.targetField = *in.sourceField`,
				`}`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			converter := &pointerToTypeConverter{}
			Test.RunConverterTestCase(t, tc, converter)
		})
	}
}
