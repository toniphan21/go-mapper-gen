package gomappergen

import "testing"

func Test_functionsConverter(t *testing.T) {
	additionalCode := []string{
		`type CustomType struct {`,
		`	ID string`,
		`}`,
		``,
		`func CustomTypeToString(t CustomType) string {`,
		`	return t.ID`,
		`}`,
		``,
		`func StringToCustomType(s string) CustomType {`,
		`	return CustomType {`,
		`		ID: s,`,
		`	}`,
		`}`,
		``,
	}
	config := &Config{
		ConverterFunctions: []ConvertFunctionConfig{
			{
				PackagePath: "github.com/toniphan21/go-mapper-gen/example",
				TypeName:    "CustomTypeToString",
			},
			{
				PackagePath: "github.com/toniphan21/go-mapper-gen/example",
				TypeName:    "StringToCustomType",
			},
		},
	}
	cases := []ConverterTestCase{
		{
			Name:               "convert string to CustomType use StringToCustomType",
			AdditionalCode:     additionalCode,
			Config:             config,
			TargetType:         "CustomType",
			SourceType:         "string",
			ExpectedCanConvert: true,
			ExpectedCode:       []string{"out.targetField = StringToCustomType(in.sourceField)"},
		},

		{
			Name:               "convert CustomType to string use CustomTypeToString",
			AdditionalCode:     additionalCode,
			Config:             config,
			TargetType:         "string",
			SourceType:         "CustomType",
			ExpectedCanConvert: true,
			ExpectedCode:       []string{"out.targetField = CustomTypeToString(in.sourceField)"},
		},

		{
			Name:               "convert *string to CustomType use other converter and StringToCustomType",
			AdditionalCode:     additionalCode,
			Config:             config,
			TargetType:         "CustomType",
			SourceType:         "*string",
			ExpectedCanConvert: true,
			ExpectedCode: []string{
				`var v0 string`,
				`if in.sourceField == nil {`,
				`	var zero string`,
				`	v0 = zero`,
				`} else {`,
				`	v0 = *in.sourceField`,
				`}`,
				`out.targetField = StringToCustomType(v0)`,
			},
		},

		{
			Name:               "convert string to *CustomType use StringToCustomType and other converter",
			AdditionalCode:     additionalCode,
			Config:             config,
			TargetType:         "*CustomType",
			SourceType:         "string",
			ExpectedCanConvert: true,
			ExpectedCode: []string{
				`v0 := StringToCustomType(in.sourceField)`,
				`out.targetField = &v0`,
			},
		},

		{
			Name:               "convert *string to *CustomType use StringToCustomType and other converters",
			AdditionalCode:     additionalCode,
			Config:             config,
			TargetType:         "*CustomType",
			SourceType:         "*string",
			ExpectedCanConvert: true,
			ExpectedCode: []string{
				`var v0 string`,
				`if in.sourceField == nil {`,
				`	var zero string`,
				`	v0 = zero`,
				`} else {`,
				`	v0 = *in.sourceField`,
				`}`,
				`v1 := StringToCustomType(v0)`,
				`out.targetField = &v1`,
			},
		},
		// ---
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			registerConverter(&identicalTypeConverter{}, 0, true)
			registerConverter(&sliceConverter{}, 1, true)
			registerConverter(&typeToPointerConverter{}, 2, true)
			registerConverter(&pointerToTypeConverter{}, 3, true)

			converter := &functionsConverter{}
			Test.RunConverterTestCase(t, tc, converter)
		})
	}
}
