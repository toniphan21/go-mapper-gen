package pgtype

import (
	"go/types"

	"github.com/dave/jennifer/jen"
	gen "github.com/toniphan21/go-mapper-gen"
)

var pgtypeFloat8 = gen.TypeInfo{PkgPath: pgtypePkgPath, PkgName: "pgtype", TypeName: "Float8"}
var pointerOfFloat64 = gen.TypeInfo{TypeName: "float64", IsPointer: true}

type float8Converter struct {
	orchestrator gen.GeneratedTypeOrchestrator
}

func (c *float8Converter) Init(parser gen.Parser, config gen.Config) {
	ce := singleValueCodeEmitter{
		Generated:         pgtypeFloat8,
		Target:            pointerOfFloat64,
		PgTypeName:        pgtypeFloat8.TypeName,
		ValuePropertyName: "Float64",
	}
	c.orchestrator = ce.toGeneratedTypeOrchestrator()
}

func (c *float8Converter) Info() gen.ConverterInfo {
	return gen.ConverterInfo{
		Name:                 "built-in lib pgtype/float8Converter",
		ShortForm:            "pgtype.Float8 <-> [T *float64]",
		ShortFormDescription: "pgtype.Float8 to T where T -> *float64 is possible",
	}
}

func (c *float8Converter) CanConvert(ctx gen.LookupContext, targetType, sourceType types.Type) bool {
	return c.orchestrator.CanConvert(c, ctx, targetType, sourceType)
}

func (c *float8Converter) ConvertField(ctx gen.ConverterContext, target, source gen.Symbol, opts gen.ConverterOption) jen.Code {
	return ctx.Run(c, opts, func() jen.Code {
		return c.orchestrator.PerformConvert(c, ctx, target, source, opts)
	})
}

var _ gen.Converter = (*int8Converter)(nil)
