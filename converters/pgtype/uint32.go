package pgtype

import (
	"go/types"

	"github.com/dave/jennifer/jen"
	gen "github.com/toniphan21/go-mapper-gen"
)

var pgtypeUint32 = gen.TypeInfo{PkgPath: pgtypePkgPath, PkgName: "pgtype", TypeName: "Uint32"}
var pointerOfUint32 = gen.TypeInfo{TypeName: "uint32", IsPointer: true}

type uint32Converter struct {
	orchestrator gen.GeneratedTypeOrchestrator
}

func (c *uint32Converter) Init(parser gen.Parser, config gen.Config) {
	ce := singleValueCodeEmitter{
		Generated:         pgtypeUint32,
		Target:            pointerOfUint32,
		PgTypeName:        pgtypeUint32.TypeName,
		ValuePropertyName: "Uint32",
	}
	c.orchestrator = ce.toGeneratedTypeOrchestrator()
}

func (c *uint32Converter) Info() gen.ConverterInfo {
	return gen.ConverterInfo{
		Name:                 "built-in lib pgtype/uint32Converter",
		ShortForm:            "pgtype.Uint32 <-> [T *int32]",
		ShortFormDescription: "pgtype.Uint32 to T where T -> int32 is possible",
	}
}

func (c *uint32Converter) CanConvert(ctx gen.LookupContext, targetType, sourceType types.Type) bool {
	return c.orchestrator.CanConvert(c, ctx, targetType, sourceType)
}

func (c *uint32Converter) ConvertField(ctx gen.ConverterContext, target, source gen.Symbol, opts gen.ConverterOption) jen.Code {
	return ctx.Run(c, opts, func() jen.Code {
		return c.orchestrator.PerformConvert(c, ctx, target, source, opts)
	})
}

var _ gen.Converter = (*uint32Converter)(nil)
