package pgtype

import (
	"go/types"

	"github.com/dave/jennifer/jen"
	gen "github.com/toniphan21/go-mapper-gen"
)

var pgtypeInt4 = gen.TypeInfo{PkgPath: pgtypePkgPath, PkgName: "pgtype", TypeName: "Int4"}
var pointerOfInt32 = gen.TypeInfo{TypeName: "int32", IsPointer: true}

type int4Converter struct {
	orchestrator gen.GeneratedTypeOrchestrator
}

func (c *int4Converter) Init(parser gen.Parser, config gen.Config) {
	ce := singleValueCodeEmitter{
		Generated:         pgtypeInt4,
		Target:            pointerOfInt32,
		PgTypeName:        pgtypeInt4.TypeName,
		ValuePropertyName: "Int32",
	}
	c.orchestrator = ce.toGeneratedTypeOrchestrator()
}

func (c *int4Converter) Info() gen.ConverterInfo {
	return gen.ConverterInfo{
		Name:                 "built-in lib pgtype/int4Converter",
		ShortForm:            "pgtype.Int4 <-> [T *int32]",
		ShortFormDescription: "pgtype.Int4 to T where T -> *int32 is possible",
	}
}

func (c *int4Converter) CanConvert(ctx gen.LookupContext, targetType, sourceType types.Type) bool {
	return c.orchestrator.CanConvert(c, ctx, targetType, sourceType)
}

func (c *int4Converter) ConvertField(ctx gen.ConverterContext, target, source gen.Symbol, opts gen.ConverterOption) jen.Code {
	return ctx.Run(c, opts, func() jen.Code {
		return c.orchestrator.PerformConvert(c, ctx, target, source, opts)
	})
}

var _ gen.Converter = (*int4Converter)(nil)
