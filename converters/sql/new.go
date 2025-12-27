package sql

import (
	"database/sql"
	"time"

	gen "github.com/toniphan21/go-mapper-gen"
)

func RegisterConverters() {
	gen.RegisterConverter(Converter.NullBool)
	gen.RegisterConverter(Converter.NullByte)
	gen.RegisterConverter(Converter.NullFloat64)
	gen.RegisterConverter(Converter.NullInt16)
	gen.RegisterConverter(Converter.NullInt32)
	gen.RegisterConverter(Converter.NullInt64)
	gen.RegisterConverter(Converter.NullString)
	gen.RegisterConverter(Converter.NullTime)
}

type converters struct {
	NullByte    gen.Converter
	NullBool    gen.Converter
	NullFloat64 gen.Converter
	NullInt16   gen.Converter
	NullInt32   gen.Converter
	NullInt64   gen.Converter
	NullString  gen.Converter
	NullTime    gen.Converter
}

var Converter = converters{
	NullByte: &baseConverter[sql.NullByte, *byte]{
		ValuePropertyName:    "Byte",
		Name:                 "built-in lib sql.Converter.NullByte",
		ShortForm:            "sql.NullByte <-> [T *byte]",
		ShortFormDescription: "sql.NullByte to T where T -> *byte is possible",
	},
	NullBool: &baseConverter[sql.NullBool, *bool]{
		ValuePropertyName:    "Bool",
		Name:                 "built-in lib sql.Converter.NullBool",
		ShortForm:            "sql.NullBool <-> [T *bool]",
		ShortFormDescription: "sql.NullBool to T where T -> *bool is possible",
	},
	NullFloat64: &baseConverter[sql.NullFloat64, *float64]{
		ValuePropertyName:    "Float64",
		Name:                 "built-in lib sql.Converter.NullFloat64",
		ShortForm:            "sql.NullFloat64 <-> [T *float64]",
		ShortFormDescription: "sql.NullFloat64 to T where T -> *float64 is possible",
	},
	NullInt16: &baseConverter[sql.NullInt16, *int16]{
		ValuePropertyName:    "Int16",
		Name:                 "built-in lib sql.Converter.NullInt16",
		ShortForm:            "sql.NullInt16 <-> [T *int16]",
		ShortFormDescription: "sql.NullInt16 to T where T -> *int16 is possible",
	},
	NullInt32: &baseConverter[sql.NullInt32, *int32]{
		ValuePropertyName:    "Int32",
		Name:                 "built-in lib sql.Converter.NullInt32",
		ShortForm:            "sql.NullInt32 <-> [T *int32]",
		ShortFormDescription: "sql.NullInt32 to T where T -> *int32 is possible",
	},
	NullInt64: &baseConverter[sql.NullInt64, *int64]{
		ValuePropertyName:    "Int64",
		Name:                 "built-in lib sql.Converter.NullInt64",
		ShortForm:            "sql.NullInt64 <-> [T *int64]",
		ShortFormDescription: "sql.NullInt64 to T where T -> *int64 is possible",
	},
	NullString: &baseConverter[sql.NullString, *string]{
		ValuePropertyName:    "String",
		Name:                 "built-in lib sql.Converter.NullString",
		ShortForm:            "sql.NullString <-> [T *string]",
		ShortFormDescription: "sql.NullString to T where T -> *string is possible",
	},
	NullTime: &baseConverter[sql.NullTime, *time.Time]{
		ValuePropertyName:    "Time",
		Name:                 "built-in lib sql.Converter.NullTime",
		ShortForm:            "sql.NullTime <-> [T *time.Time]",
		ShortFormDescription: "sql.NullTime to T where T -> *time.Time is possible",
	},
}
