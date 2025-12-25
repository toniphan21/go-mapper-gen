package pgtype

import (
	"go/types"

	"github.com/dave/jennifer/jen"
	gen "github.com/toniphan21/go-mapper-gen"
)

var pgtypeInt2 = gen.TypeInfo{PkgPath: pgtypePkgPath, PkgName: "pgtype", TypeName: "Int2"}
var pointerOfInt16 = gen.TypeInfo{TypeName: "int16", IsPointer: true}

type int2Converter struct {
	orchestrator gen.GeneratedTypeOrchestrator
}

func (c *int2Converter) Init(parser gen.Parser, config gen.Config) {
	ce := singleValueCodeEmitter{
		Generated:         pgtypeInt2,
		Target:            pointerOfInt16,
		PgTypeName:        pgtypeInt2.TypeName,
		ValuePropertyName: "Int16",
	}
	c.orchestrator = ce.toGeneratedTypeOrchestrator()
}

func (c *int2Converter) Info() gen.ConverterInfo {
	return gen.ConverterInfo{
		Name:                 "built-in lib pgtype/int2Converter",
		ShortForm:            "pgtype.Int2 <-> [T *int16]",
		ShortFormDescription: "pgtype.Int2 to T where T -> *int16 is possible",
	}
}

func (c *int2Converter) CanConvert(ctx gen.LookupContext, targetType, sourceType types.Type) bool {
	return c.orchestrator.CanConvert(c, ctx, targetType, sourceType)
}

func (c *int2Converter) ConvertField(ctx gen.ConverterContext, target, source gen.Symbol, opts gen.ConverterOption) jen.Code {
	return ctx.Run(c, opts, func() jen.Code {
		return c.orchestrator.PerformConvert(c, ctx, target, source, opts)
	})
}

var _ gen.Converter = (*int2Converter)(nil)
