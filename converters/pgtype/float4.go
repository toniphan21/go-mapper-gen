package pgtype

import (
	"go/types"

	"github.com/dave/jennifer/jen"
	gen "github.com/toniphan21/go-mapper-gen"
)

var pgtypeFloat4 = gen.TypeInfo{PkgPath: pgtypePkgPath, PkgName: "pgtype", TypeName: "Float4"}
var pointerOfFloat32 = gen.TypeInfo{TypeName: "float32", IsPointer: true}

type float4Converter struct {
	orchestrator gen.GeneratedTypeOrchestrator
}

func (c *float4Converter) Init(parser gen.Parser, config gen.Config) {
	ce := singleValueCodeEmitter{
		Generated:         pgtypeFloat4,
		Target:            pointerOfFloat32,
		PgTypeName:        pgtypeFloat4.TypeName,
		ValuePropertyName: "Float32",
	}
	c.orchestrator = ce.toGeneratedTypeOrchestrator()
}

func (c *float4Converter) Info() gen.ConverterInfo {
	return gen.ConverterInfo{
		Name:                 "built-in lib pgtype/float4Converter",
		ShortForm:            "pgtype.Float4 <-> [T *float32]",
		ShortFormDescription: "pgtype.Float4 to T where T -> *float32 is possible",
	}
}

func (c *float4Converter) CanConvert(ctx gen.LookupContext, targetType, sourceType types.Type) bool {
	return c.orchestrator.CanConvert(c, ctx, targetType, sourceType)
}

func (c *float4Converter) ConvertField(ctx gen.ConverterContext, target, source gen.Symbol, opts gen.ConverterOption) jen.Code {
	return ctx.Run(c, opts, func() jen.Code {
		return c.orchestrator.PerformConvert(c, ctx, target, source, opts)
	})
}

var _ gen.Converter = (*int8Converter)(nil)
