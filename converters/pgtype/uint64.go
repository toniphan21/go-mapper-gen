package pgtype

import (
	"go/types"

	"github.com/dave/jennifer/jen"
	gen "github.com/toniphan21/go-mapper-gen"
)

var pgtypeUint64 = gen.TypeInfo{PkgPath: pgtypePkgPath, PkgName: "pgtype", TypeName: "Uint64"}
var pointerOfUint64 = gen.TypeInfo{TypeName: "uint64", IsPointer: true}

type uint64Converter struct {
	orchestrator gen.GeneratedTypeOrchestrator
}

func (c *uint64Converter) Init(parser gen.Parser, config gen.Config) {
	ce := singleValueCodeEmitter{
		Generated:         pgtypeUint64,
		Target:            pointerOfUint64,
		PgTypeName:        pgtypeUint64.TypeName,
		ValuePropertyName: "Uint64",
	}
	c.orchestrator = ce.toGeneratedTypeOrchestrator()
}

func (c *uint64Converter) Info() gen.ConverterInfo {
	return gen.ConverterInfo{
		Name:                 "built-in lib pgtype/uint64Converter",
		ShortForm:            "pgtype.Uint64 <-> [T *int64]",
		ShortFormDescription: "pgtype.Uint64 to T where T -> int64 is possible",
	}
}

func (c *uint64Converter) CanConvert(ctx gen.LookupContext, targetType, sourceType types.Type) bool {
	return c.orchestrator.CanConvert(c, ctx, targetType, sourceType)
}

func (c *uint64Converter) ConvertField(ctx gen.ConverterContext, target, source gen.Symbol, opts gen.ConverterOption) jen.Code {
	return ctx.Run(c, opts, func() jen.Code {
		return c.orchestrator.PerformConvert(c, ctx, target, source, opts)
	})
}

var _ gen.Converter = (*uint64Converter)(nil)
