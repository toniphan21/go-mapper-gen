package pgtype

import (
	"go/types"

	"github.com/dave/jennifer/jen"
	gen "github.com/toniphan21/go-mapper-gen"
)

var pgtypeInt8 = gen.TypeInfo{PkgPath: pgtypePkgPath, PkgName: "pgtype", TypeName: "Int8"}
var pointerOfInt64 = gen.TypeInfo{TypeName: "int64", IsPointer: true}

type int8Converter struct {
	orchestrator gen.GeneratedTypeOrchestrator
}

func (c *int8Converter) Init(parser gen.Parser, config gen.Config) {
	ce := singleValueCodeEmitter{
		Generated:         pgtypeInt8,
		Target:            pointerOfInt64,
		PgTypeName:        pgtypeInt8.TypeName,
		ValuePropertyName: "Int64",
	}
	c.orchestrator = ce.toGeneratedTypeOrchestrator()
}

func (c *int8Converter) Info() gen.ConverterInfo {
	return gen.ConverterInfo{
		Name:                 "built-in lib pgtype/int8Converter",
		ShortForm:            "pgtype.Int8 <-> [T *int64]",
		ShortFormDescription: "pgtype.Int8 to T where T -> *int64 is possible",
	}
}

func (c *int8Converter) CanConvert(ctx gen.LookupContext, targetType, sourceType types.Type) bool {
	return c.orchestrator.CanConvert(c, ctx, targetType, sourceType)
}

func (c *int8Converter) ConvertField(ctx gen.ConverterContext, target, source gen.Symbol, opts gen.ConverterOption) jen.Code {
	return ctx.Run(c, opts, func() jen.Code {
		return c.orchestrator.PerformConvert(c, ctx, target, source, opts)
	})
}

var _ gen.Converter = (*int8Converter)(nil)
