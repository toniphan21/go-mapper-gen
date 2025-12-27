package pgtype

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	gen "github.com/toniphan21/go-mapper-gen"
)

func RegisterConverters() {
	gen.RegisterConverter(Converter.Bool)
	gen.RegisterConverter(Converter.Int2)
	gen.RegisterConverter(Converter.Int4)
	gen.RegisterConverter(Converter.Int8)
	gen.RegisterConverter(Converter.Float4)
	gen.RegisterConverter(Converter.Float8)
	gen.RegisterConverter(Converter.Uint32)
	gen.RegisterConverter(Converter.Uint64)
	gen.RegisterConverter(Converter.Text)
	gen.RegisterConverter(Converter.Date)
	gen.RegisterConverter(Converter.Time)
	gen.RegisterConverter(Converter.Timestamp)
	gen.RegisterConverter(Converter.Timestamptz)
}

type converters struct {
	Bool        gen.Converter
	Int2        gen.Converter
	Int4        gen.Converter
	Int8        gen.Converter
	Float4      gen.Converter
	Float8      gen.Converter
	Uint32      gen.Converter
	Uint64      gen.Converter
	Text        gen.Converter
	Date        gen.Converter
	Time        gen.Converter
	Timestamp   gen.Converter
	Timestamptz gen.Converter
}

var Converter = converters{
	Bool: &baseConverter[pgtype.Bool, *bool]{
		ValuePropertyName:    "Bool",
		Name:                 "built-in lib pgtype.Converter.Bool",
		ShortForm:            "pgtype.Bool <-> [T *bool]",
		ShortFormDescription: "pgtype.Bool to T where T -> *bool is possible",
	},
	Date: &baseConverter[pgtype.Date, *time.Time]{
		ValuePropertyName:    "Time",
		Name:                 "built-in lib pgtype.Converter.Date",
		ShortForm:            "pgtype.Date <-> [T *time.Time]",
		ShortFormDescription: "pgtype.Date to T where T -> *time.Time is possible",
	},
	Float4: &baseConverter[pgtype.Float4, *float32]{
		ValuePropertyName:    "Float32",
		Name:                 "built-in lib pgtype.Converter.Float4",
		ShortForm:            "pgtype.Float4 <-> [T *float32]",
		ShortFormDescription: "pgtype.Float4 to T where T -> *float32 is possible",
	},
	Float8: &baseConverter[pgtype.Float8, *float64]{
		ValuePropertyName:    "Float64",
		Name:                 "built-in lib pgtype.Converter.Float8",
		ShortForm:            "pgtype.Float8 <-> [T *float64]",
		ShortFormDescription: "pgtype.Float8 to T where T -> *float64 is possible",
	},
	Int2: &baseConverter[pgtype.Int2, *int16]{
		ValuePropertyName:    "Int16",
		Name:                 "built-in lib pgtype.Converter.Int2",
		ShortForm:            "pgtype.Int2 <-> [T *int16]",
		ShortFormDescription: "pgtype.Int2 to T where T -> *int16 is possible",
	},
	Int4: &baseConverter[pgtype.Int4, *int32]{
		ValuePropertyName:    "Int32",
		Name:                 "built-in lib pgtype.Converter.Int4",
		ShortForm:            "pgtype.Int4 <-> [T *int32]",
		ShortFormDescription: "pgtype.Int4 to T where T -> *int32 is possible",
	},
	Int8: &baseConverter[pgtype.Int8, *int64]{
		ValuePropertyName:    "Int64",
		Name:                 "built-in lib pgtype.Converter.Int8",
		ShortForm:            "pgtype.Int8 <-> [T *int64]",
		ShortFormDescription: "pgtype.Int8 to T where T -> *int64 is possible",
	},
	Uint32: &baseConverter[pgtype.Uint32, *uint32]{
		ValuePropertyName:    "Uint32",
		Name:                 "built-in lib pgtype.Converter.Uint32",
		ShortForm:            "pgtype.Uint32 <-> [T *int32]",
		ShortFormDescription: "pgtype.Uint32 to T where T -> int32 is possible",
	},
	Uint64: &baseConverter[pgtype.Uint64, *uint64]{
		ValuePropertyName:    "Uint64",
		Name:                 "built-in lib pgtype.Converter.Uint64",
		ShortForm:            "pgtype.Uint64 <-> [T *int64]",
		ShortFormDescription: "pgtype.Uint64 to T where T -> int64 is possible",
	},
	Text: &baseConverter[pgtype.Text, *string]{
		ValuePropertyName:    "String",
		Name:                 "built-in lib pgtype.Converter.Text",
		ShortForm:            "pgtype.Text <-> [T *string]",
		ShortFormDescription: "pgtype.Text to T where T -> *string is possible",
	},
	Time: &baseConverter[pgtype.Time, *int64]{
		ValuePropertyName:    "Microseconds",
		Name:                 "built-in lib pgtype.Converter.Time",
		ShortForm:            "pgtype.Time <-> [T *int64]",
		ShortFormDescription: "pgtype.Time to T where T -> *int64 is possible",
	},
	Timestamp: &baseConverter[pgtype.Timestamp, *time.Time]{
		ValuePropertyName:    "Time",
		Name:                 "built-in lib pgtype.Converter.Timestamp",
		ShortForm:            "pgtype.Timestamp <-> [T *time.Time]",
		ShortFormDescription: "pgtype.Timestamp to T where T -> *time.Time is possible",
	},
	Timestamptz: &baseConverter[pgtype.Timestamptz, *time.Time]{
		ValuePropertyName:    "Time",
		Name:                 "built-in lib pgtype.Converter.Timestamptz",
		ShortForm:            "pgtype.Timestamptz <-> [T *time.Time]",
		ShortFormDescription: "pgtype.Timestamptz to T where T -> *time.Time is possible",
	},
}
