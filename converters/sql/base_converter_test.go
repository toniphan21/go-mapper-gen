package sql

import (
	"fmt"
	"testing"

	gen "github.com/toniphan21/go-mapper-gen"
)

func Test_baseConverter_convertible(t *testing.T) {
	pointerToSqlNull := func(sqlTypeName string, property string) []string {
		var pl, vl int
		p := len(property)
		v := len("Valid")
		if p == v {
			pl = 1
			vl = 1
		}
		if p > v {
			pl = 1
			vl = 1 + p - v
		}
		if p < v {
			vl = 1
			pl = 1 + v - p
		}

		return []string{
			`if in.sourceField != nil {`,
			fmt.Sprintf(`	out.targetField = sql.%s{`, sqlTypeName),
			fmt.Sprintf(`		%s:%-*s*in.sourceField,`, property, pl, " "),
			fmt.Sprintf(`		Valid:%-*strue,`, vl, " "),
			`	}`,
			`}`,
		}
	}

	nonePointerToSqlNull := func(targetType, sqlTypeName string, property string) []string {
		var pl, vl int
		p := len(property)
		v := len("Valid")
		if p == v {
			pl = 1
			vl = 1
		}
		if p > v {
			pl = 1
			vl = 1 + p - v
		}
		if p < v {
			vl = 1
			pl = 1 + v - p
		}

		return []string{
			``,
			fmt.Sprintf(`var v0 %s`, targetType),
			`v0 = &in.sourceField`,
			`if v0 != nil {`,
			fmt.Sprintf(`	out.targetField = sql.%s{`, sqlTypeName),
			fmt.Sprintf(`		%s:%-*s*v0,`, property, pl, " "),
			fmt.Sprintf(`		Valid:%-*strue,`, vl, " "),
			`	}`,
			`}`,
		}
	}

	sqlNullToPointer := func(property string) []string {
		return []string{
			`if in.sourceField.Valid {`,
			fmt.Sprintf(`	out.targetField = &in.sourceField.%s`, property),
			`}`,
		}
	}

	sqlNullToNonePointer := func(targetType, property string) []string {
		return []string{
			``,
			fmt.Sprintf(`var v0 *%s`, targetType),
			`if in.sourceField.Valid {`,
			fmt.Sprintf(`	v0 = &in.sourceField.%s`, property),
			`}`,
			`if v0 != nil {`,
			`	out.targetField = *v0`,
			`} else {`,
			fmt.Sprintf(`	var zero %s`, targetType),
			`	out.targetField = zero`,
			`}`,
			``,
		}
	}

	cases := []struct {
		instance        gen.Converter
		target          string
		source          string
		imports         map[string]string
		expectedImports []string
		expected        []string
		importSQL       bool
		printSetup      bool
	}{
		// --- NullByte
		{
			instance: Converter.NullByte, source: "*byte", target: "sql.NullByte", importSQL: true,
			expected: pointerToSqlNull("NullByte", "Byte"),
		},
		{
			instance: Converter.NullByte, source: "sql.NullByte", target: "*byte",
			expected: sqlNullToPointer("Byte"),
		},
		{
			instance: Converter.NullByte, source: "sql.NullByte", target: "*byte",
			expected: sqlNullToPointer("Byte"),
		},
		{
			instance: Converter.NullByte, source: "sql.NullByte", target: "byte",
			expected: []string{
				``,
				`var v0 *uint8`, // byte is just an alias of uint8
				`if in.sourceField.Valid {`,
				`	v0 = &in.sourceField.Byte`,
				`}`,
				`if v0 != nil {`,
				`	out.targetField = *v0`,
				`} else {`,
				`	var zero byte`,
				`	out.targetField = zero`,
				`}`,
				``,
			},
		},

		// --- NullBool
		{
			instance: Converter.NullBool, source: "*bool", target: "sql.NullBool", importSQL: true,
			expected: pointerToSqlNull("NullBool", "Bool"),
		},
		{
			instance: Converter.NullBool, source: "bool", target: "sql.NullBool", importSQL: true,
			expected: nonePointerToSqlNull("*bool", "NullBool", "Bool"),
		},
		{
			instance: Converter.NullBool, source: "sql.NullBool", target: "*bool",
			expected: sqlNullToPointer("Bool"),
		},
		{
			instance: Converter.NullBool, source: "sql.NullBool", target: "bool",
			expected: sqlNullToNonePointer("bool", "Bool"),
		},

		// --- NullFloat64
		{
			instance: Converter.NullFloat64, source: "*float64", target: "sql.NullFloat64", importSQL: true,
			expected: pointerToSqlNull("NullFloat64", "Float64"),
		},
		{
			instance: Converter.NullFloat64, source: "sql.NullFloat64", target: "*float64",
			expected: sqlNullToPointer("Float64"),
		},
		{
			instance: Converter.NullFloat64, source: "sql.NullFloat64", target: "*float64",
			expected: sqlNullToPointer("Float64"),
		},
		{
			instance: Converter.NullFloat64, source: "sql.NullFloat64", target: "float64",
			expected: sqlNullToNonePointer("float64", "Float64"),
		},

		// --- NullInt16
		{
			instance: Converter.NullInt16, source: "*int16", target: "sql.NullInt16", importSQL: true,
			expected: pointerToSqlNull("NullInt16", "Int16"),
		},
		{
			instance: Converter.NullInt16, source: "sql.NullInt16", target: "*int16",
			expected: sqlNullToPointer("Int16"),
		},
		{
			instance: Converter.NullInt16, source: "sql.NullInt16", target: "*int16",
			expected: sqlNullToPointer("Int16"),
		},
		{
			instance: Converter.NullInt16, source: "sql.NullInt16", target: "int16",
			expected: sqlNullToNonePointer("int16", "Int16"),
		},

		// --- NullInt32
		{
			instance: Converter.NullInt32, source: "*int32", target: "sql.NullInt32", importSQL: true,
			expected: pointerToSqlNull("NullInt32", "Int32"),
		},
		{
			instance: Converter.NullInt32, source: "sql.NullInt32", target: "*int32",
			expected: sqlNullToPointer("Int32"),
		},
		{
			instance: Converter.NullInt32, source: "sql.NullInt32", target: "*int32",
			expected: sqlNullToPointer("Int32"),
		},
		{
			instance: Converter.NullInt32, source: "sql.NullInt32", target: "int32",
			expected: sqlNullToNonePointer("int32", "Int32"),
		},

		// --- NullInt64
		{
			instance: Converter.NullInt64, source: "*int64", target: "sql.NullInt64", importSQL: true,
			expected: pointerToSqlNull("NullInt64", "Int64"),
		},
		{
			instance: Converter.NullInt64, source: "sql.NullInt64", target: "*int64",
			expected: sqlNullToPointer("Int64"),
		},
		{
			instance: Converter.NullInt64, source: "sql.NullInt64", target: "*int64",
			expected: sqlNullToPointer("Int64"),
		},
		{
			instance: Converter.NullInt64, source: "sql.NullInt64", target: "int64",
			expected: sqlNullToNonePointer("int64", "Int64"),
		},

		// --- NullString
		{
			instance: Converter.NullString, source: "*string", target: "sql.NullString", importSQL: true,
			expected: pointerToSqlNull("NullString", "String"),
		},
		{
			instance: Converter.NullString, source: "sql.NullString", target: "*string",
			expected: sqlNullToPointer("String"),
		},
		{
			instance: Converter.NullString, source: "sql.NullString", target: "*string",
			expected: sqlNullToPointer("String"),
		},
		{
			instance: Converter.NullString, source: "sql.NullString", target: "string",
			expected: sqlNullToNonePointer("string", "String"),
		},

		// --- NullTime
		{
			instance: Converter.NullTime, source: "*time.Time", target: "sql.NullTime",
			imports:   map[string]string{"time": "time", "sql": "database/sql"},
			importSQL: true,
			expected:  pointerToSqlNull("NullTime", "Time"),
		},
		{
			instance: Converter.NullTime, source: "sql.NullTime", target: "*time.Time",
			imports:  map[string]string{"time": "time", "sql": "database/sql"},
			expected: sqlNullToPointer("Time"),
		},
		{
			instance: Converter.NullTime, source: "sql.NullTime", target: "*time.Time",
			imports:  map[string]string{"time": "time", "sql": "database/sql"},
			expected: sqlNullToPointer("Time"),
		},
		{
			instance: Converter.NullTime, source: "sql.NullTime", target: "time.Time",
			imports:         map[string]string{"time": "time", "sql": "database/sql"},
			expectedImports: []string{`import "time"`},
			expected:        sqlNullToNonePointer("time.Time", "Time"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.source+" -> "+tc.target, func(t *testing.T) {
			ctc := gen.ConverterTestCase{
				Name:               t.Name(),
				TargetType:         tc.target,
				SourceType:         tc.source,
				ExpectedCanConvert: true,
				ExpectedCode:       tc.expected,
				PrintSetUp:         tc.printSetup,
			}

			if tc.imports == nil {
				ctc.Imports = map[string]string{"sql": "database/sql"}
			} else {
				ctc.Imports = tc.imports
			}

			if tc.expectedImports == nil {
				if tc.importSQL {
					ctc.ExpectedImports = []string{`import "database/sql"`}
				}
			} else {
				ctc.ExpectedImports = tc.expectedImports
			}

			gen.ClearAllRegisteredConverters()
			gen.RegisterConverter(gen.BuiltinConverters.IdenticalType)
			gen.RegisterConverter(gen.BuiltinConverters.Slice)
			gen.RegisterConverter(gen.BuiltinConverters.PointerToType)
			gen.RegisterConverter(gen.BuiltinConverters.TypeToPointer)
			gen.RegisterConverter(gen.BuiltinConverters.Functions)
			gen.RegisterConverter(tc.instance)
			gen.RegisterConverter(gen.BuiltinConverters.Numeric)
			tc.instance.Init(nil, gen.Config{}, nil)

			gen.Test.RunConverterTestCase(t, ctc, tc.instance)
		})
	}
}
