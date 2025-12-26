package pgtype

import gen "github.com/toniphan21/go-mapper-gen"

const pgtypePkgPath = "github.com/jackc/pgx/v5/pgtype"

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
}

type converters struct {
	Bool   gen.Converter
	Int2   gen.Converter
	Int4   gen.Converter
	Int8   gen.Converter
	Float4 gen.Converter
	Float8 gen.Converter
	Uint32 gen.Converter
	Uint64 gen.Converter
	Text   gen.Converter
}

var Converter = converters{
	Bool:   &boolConverter{},
	Int2:   &int2Converter{},
	Int4:   &int4Converter{},
	Int8:   &int8Converter{},
	Float4: &float4Converter{},
	Float8: &float8Converter{},
	Uint32: &uint32Converter{},
	Uint64: &uint64Converter{},
	Text:   &textConverter{},
}
