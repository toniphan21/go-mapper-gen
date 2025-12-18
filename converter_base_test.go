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
			GoModRequires:      map[string]string{"github.com/jackc/pgx/v5": "v5.5.4"},
			SourceType:         "pgtype.Text",
			TargetType:         "*pgtype.Text",
			ExpectedCanConvert: true,
			ExpectedCode:       []string{"out.targetField = &in.sourceField"},
		},

		{
			Name:               "emit trace comments",
			SourceType:         "bool",
			TargetType:         "*bool",
			ConverterOption:    ConverterOption{EmitTraceComments: true},
			ExpectedCanConvert: true,
			ExpectedCode: []string{
				"// built-in typeToPointerConverter generated code start",
				"out.targetField = &in.sourceField",
				"// built-in typeToPointerConverter generated code end",
			},
		},

		{
			Name:                         "without fieldName in targetSymbol",
			SourceType:                   "bool",
			TargetType:                   "*bool",
			TargetSymbolWithoutFieldName: true,
			ExpectedCanConvert:           true,
			ExpectedCode:                 []string{"target = &in.sourceField"},
		},

		{
			Name:                         "without fieldName in sourceSymbol",
			SourceType:                   "bool",
			TargetType:                   "*bool",
			SourceSymbolWithoutFieldName: true,
			ExpectedCanConvert:           true,
			ExpectedCode:                 []string{"out.targetField = &source"},
		},

		{
			Name:                         "without fieldName in sourceSymbol and targetSymbol",
			SourceType:                   "bool",
			TargetType:                   "*bool",
			TargetSymbolWithoutFieldName: true,
			SourceSymbolWithoutFieldName: true,
			ExpectedCanConvert:           true,
			ExpectedCode:                 []string{"target = &source"},
		},
		// ---
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
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
				`	var zero bool`,
				`	out.targetField = zero`,
				`} else {`,
				`	out.targetField = *in.sourceField`,
				`}`,
			},
		},

		{
			Name:               "emit trace comments",
			SourceType:         "*bool",
			TargetType:         "bool",
			ConverterOption:    ConverterOption{EmitTraceComments: true},
			ExpectedCanConvert: true,
			ExpectedCode: []string{
				"// built-in pointerToTypeConverter generated code start",
				`if in.sourceField == nil {`,
				`	var zero bool`,
				`	out.targetField = zero`,
				`} else {`,
				`	out.targetField = *in.sourceField`,
				`}`,
				"// built-in pointerToTypeConverter generated code end",
			},
		},

		{
			Name:                         "without fieldName in targetSymbol",
			SourceType:                   "*bool",
			TargetType:                   "bool",
			TargetSymbolWithoutFieldName: true,
			ExpectedCanConvert:           true,
			ExpectedCode: []string{
				`if in.sourceField == nil {`,
				`	var zero bool`,
				`	target = zero`,
				`} else {`,
				`	target = *in.sourceField`,
				`}`,
			},
		},

		{
			Name:                         "without fieldName in sourceSymbol",
			SourceType:                   "*bool",
			TargetType:                   "bool",
			SourceSymbolWithoutFieldName: true,
			ExpectedCanConvert:           true,
			ExpectedCode: []string{
				`if source == nil {`,
				`	var zero bool`,
				`	out.targetField = zero`,
				`} else {`,
				`	out.targetField = *source`,
				`}`,
			},
		},

		{
			Name:                         "without fieldName in targetSymbol and sourceSymbol",
			SourceType:                   "*bool",
			TargetType:                   "bool",
			TargetSymbolWithoutFieldName: true,
			SourceSymbolWithoutFieldName: true,
			ExpectedCanConvert:           true,
			ExpectedCode: []string{
				`if source == nil {`,
				`	var zero bool`,
				`	target = zero`,
				`} else {`,
				`	target = *source`,
				`}`,
			},
		},
		// ---
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			converter := &pointerToTypeConverter{}
			Test.RunConverterTestCase(t, tc, converter)
		})
	}
}

func Test_sliceConverter(t *testing.T) {
	cases := []ConverterTestCase{
		{Name: "cannot convert bool to []bool", SourceType: "bool", TargetType: "[]bool"},
		{Name: "cannot convert []int to int", SourceType: "[]int", TargetType: "int"},
		{Name: "cannot convert []int to []int32", SourceType: "[]int", TargetType: "[]int32"},

		{
			Name:               "[]bool to []bool",
			SourceType:         "[]bool",
			TargetType:         "[]bool",
			ExpectedCanConvert: true,
			ExpectedCode: []string{
				`if in.sourceField == nil {`,
				`	out.targetField = nil`,
				`} else {`,
				`	out.targetField = make([]bool, len(in.sourceField))`,
				`	for i, v := range in.sourceField {`,
				`		out.targetField[i] = v`,
				`	}`,
				`}`,
			},
		},

		{
			Name:               "[]string to []*string",
			SourceType:         "[]string",
			TargetType:         "[]*string",
			ExpectedCanConvert: true,
			ExpectedCode: []string{
				`if in.sourceField == nil {`,
				`	out.targetField = nil`,
				`} else {`,
				`	out.targetField = make([]*string, len(in.sourceField))`,
				`	for i, v := range in.sourceField {`,
				`		out.targetField[i] = &v`,
				`	}`,
				`}`,
			},
		},

		{
			Name:               "[]*int to []int",
			SourceType:         "[]*int",
			TargetType:         "[]int",
			ExpectedCanConvert: true,
			ExpectedCode: []string{
				`if in.sourceField == nil {`,
				`	out.targetField = nil`,
				`} else {`,
				`	out.targetField = make([]int, len(in.sourceField))`,
				`	for i, v := range in.sourceField {`,
				`		if v == nil {`,
				`			var zero int`,
				`			out.targetField[i] = zero`,
				`		} else {`,
				`			out.targetField[i] = *v`,
				`		}`,
				`	}`,
				`}`,
			},
		},

		{
			Name:               "emit trace comments",
			SourceType:         "[]*int",
			TargetType:         "[]int",
			ConverterOption:    ConverterOption{EmitTraceComments: true},
			ExpectedCanConvert: true,
			ExpectedCode: []string{
				`// built-in sliceConverter generated code start`,
				`if in.sourceField == nil {`,
				`	out.targetField = nil`,
				`} else {`,
				`	out.targetField = make([]int, len(in.sourceField))`,
				`	for i, v := range in.sourceField {`,
				`		// built-in pointerToTypeConverter generated code start`,
				`		if v == nil {`,
				`			var zero int`,
				`			out.targetField[i] = zero`,
				`		} else {`,
				`			out.targetField[i] = *v`,
				`		}`,
				`		// built-in pointerToTypeConverter generated code end`,
				`	}`,
				`}`,
				`// built-in sliceConverter generated code end`,
			},
		},

		{
			Name:                         "without fieldName in targetSymbol",
			SourceType:                   "[]*int",
			TargetType:                   "[]int",
			TargetSymbolWithoutFieldName: true,
			ExpectedCanConvert:           true,
			ExpectedCode: []string{
				`if in.sourceField == nil {`,
				`	target = nil`,
				`} else {`,
				`	target = make([]int, len(in.sourceField))`,
				`	for i, v := range in.sourceField {`,
				`		if v == nil {`,
				`			var zero int`,
				`			target[i] = zero`,
				`		} else {`,
				`			target[i] = *v`,
				`		}`,
				`	}`,
				`}`,
			},
		},
		// ---
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			converter := &sliceConverter{}
			registerConverter(&identicalTypeConverter{}, 0, true)
			registerConverter(&typeToPointerConverter{}, 1, true)
			registerConverter(&pointerToTypeConverter{}, 2, true)

			Test.RunConverterTestCase(t, tc, converter)
		})
	}
}
