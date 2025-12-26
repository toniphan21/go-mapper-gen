package gomappergen

import "testing"

func Test_functionsConverter_UseFunctions(t *testing.T) {
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
				`if in.sourceField != nil {`,
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
				`if in.sourceField != nil {`,
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
			registerBuiltInConverter(&identicalTypeConverter{}, 0)
			registerBuiltInConverter(&sliceConverter{}, 1)
			registerBuiltInConverter(&typeToPointerConverter{}, 2)
			registerBuiltInConverter(&pointerToTypeConverter{}, 3)

			converter := &functionsConverter{}
			Test.RunConverterTestCase(t, tc, converter)
		})
	}
}

func Test_functionsConverter_UseMethodsInVariable(t *testing.T) {
	additionalCode := []string{
		`type CustomType struct {`,
		`	ID string`,
		`}`,
		``,
		`type convertHelper struct{}`,
		``,
		`func (h *convertHelper) CustomTypeToString(t CustomType) string {`,
		`	return t.ID`,
		`}`,
		``,
		`func (h *convertHelper) StringToCustomType(s string) CustomType {`,
		`	return CustomType {`,
		`		ID: s,`,
		`	}`,
		`}`,
		``,
		`var Converters = &convertHelper{}`,
		``,
	}

	config := &Config{
		ConverterFunctions: []ConvertFunctionConfig{
			{
				PackagePath: "github.com/toniphan21/go-mapper-gen/example",
				TypeName:    "Converters",
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
			ExpectedCode:       []string{"out.targetField = Converters.StringToCustomType(in.sourceField)"},
		},

		{
			Name:               "convert CustomType to string use CustomTypeToString",
			AdditionalCode:     additionalCode,
			Config:             config,
			TargetType:         "string",
			SourceType:         "CustomType",
			ExpectedCanConvert: true,
			ExpectedCode:       []string{"out.targetField = Converters.CustomTypeToString(in.sourceField)"},
		},

		{
			Name:               "convert *string to CustomType use other converter and Converters.StringToCustomType",
			AdditionalCode:     additionalCode,
			Config:             config,
			TargetType:         "CustomType",
			SourceType:         "*string",
			ExpectedCanConvert: true,
			ExpectedCode: []string{
				`var v0 string`,
				`if in.sourceField != nil {`,
				`	v0 = *in.sourceField`,
				`}`,
				`out.targetField = Converters.StringToCustomType(v0)`,
			},
		},

		{
			Name:               "convert string to *CustomType use Converters.StringToCustomType and other converter",
			AdditionalCode:     additionalCode,
			Config:             config,
			TargetType:         "*CustomType",
			SourceType:         "string",
			ExpectedCanConvert: true,
			ExpectedCode: []string{
				`v0 := Converters.StringToCustomType(in.sourceField)`,
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
				`if in.sourceField != nil {`,
				`	v0 = *in.sourceField`,
				`}`,
				`v1 := Converters.StringToCustomType(v0)`,
				`out.targetField = &v1`,
			},
		},
		// ---
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			registerBuiltInConverter(&identicalTypeConverter{}, 0)
			registerBuiltInConverter(&sliceConverter{}, 1)
			registerBuiltInConverter(&typeToPointerConverter{}, 2)
			registerBuiltInConverter(&pointerToTypeConverter{}, 3)

			converter := &functionsConverter{}
			Test.RunConverterTestCase(t, tc, converter)
		})
	}
}

func Test_functionsConverter_Use_Functions_And_MethodsInVariable(t *testing.T) {
	additionalCode := []string{
		`type CustomType struct {`,
		`	ID string`,
		`}`,
		``,
		`type convertHelper struct{}`,
		``,
		`func (h *convertHelper) CustomTypeToString(t CustomType) string {`,
		`	return t.ID`,
		`}`,
		``,
		`func StringToCustomType(s string) CustomType {`,
		`	return CustomType {`,
		`		ID: s,`,
		`	}`,
		`}`,
		``,
		`var Converters = &convertHelper{}`,
		``,
	}

	config := &Config{
		ConverterFunctions: []ConvertFunctionConfig{
			{
				PackagePath: "github.com/toniphan21/go-mapper-gen/example",
				TypeName:    "Converters",
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
			ExpectedCode:       []string{"out.targetField = Converters.CustomTypeToString(in.sourceField)"},
		},

		{
			Name:               "convert *string to CustomType use other converter and Converters.StringToCustomType",
			AdditionalCode:     additionalCode,
			Config:             config,
			TargetType:         "CustomType",
			SourceType:         "*string",
			ExpectedCanConvert: true,
			ExpectedCode: []string{
				`var v0 string`,
				`if in.sourceField != nil {`,
				`	v0 = *in.sourceField`,
				`}`,
				`out.targetField = StringToCustomType(v0)`,
			},
		},

		{
			Name:               "convert string to *CustomType use Converters.StringToCustomType and other converter",
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
			Name:               "convert *string to *CustomType use CustomTypeToString and other converters",
			AdditionalCode:     additionalCode,
			Config:             config,
			TargetType:         "*string",
			SourceType:         "*CustomType",
			ExpectedCanConvert: true,
			ExpectedCode: []string{
				`var v0 CustomType`,
				`if in.sourceField != nil {`,
				`	v0 = *in.sourceField`,
				`}`,
				`v1 := Converters.CustomTypeToString(v0)`,
				`out.targetField = &v1`,
			},
		},
		// ---
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			registerBuiltInConverter(&identicalTypeConverter{}, 0)
			registerBuiltInConverter(&sliceConverter{}, 1)
			registerBuiltInConverter(&typeToPointerConverter{}, 2)
			registerBuiltInConverter(&pointerToTypeConverter{}, 3)

			converter := &functionsConverter{}
			Test.RunConverterTestCase(t, tc, converter)
		})
	}
}
