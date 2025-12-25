package pgtype

import gen "github.com/toniphan21/go-mapper-gen"

const pgtypePkgPath = "github.com/jackc/pgx/v5/pgtype"

func RegisterConverters() {
	gen.RegisterConverter(&boolConverter{}, gen.RegisteredConverterCount())
	gen.RegisterConverter(&int2Converter{}, gen.RegisteredConverterCount())
	gen.RegisterConverter(&int4Converter{}, gen.RegisteredConverterCount())
	gen.RegisterConverter(&int8Converter{}, gen.RegisteredConverterCount())
	gen.RegisterConverter(&float4Converter{}, gen.RegisteredConverterCount())
	gen.RegisterConverter(&float8Converter{}, gen.RegisteredConverterCount())
	gen.RegisterConverter(&uint32Converter{}, gen.RegisteredConverterCount())
	gen.RegisterConverter(&uint64Converter{}, gen.RegisteredConverterCount())
	gen.RegisterConverter(&textConverter{}, gen.RegisteredConverterCount())
}
