package pgtype

import (
	"go/types"

	"github.com/dave/jennifer/jen"
	gen "github.com/toniphan21/go-mapper-gen"
)

var pgtypeBool = gen.TypeInfo{PkgPath: pgtypePkgPath, PkgName: "pgtype", TypeName: "Bool"}
var pointerOfBool = gen.TypeInfo{TypeName: "bool", IsPointer: true}

type boolConverter struct {
	orchestrator gen.GeneratedTypeOrchestrator
}

func (c *boolConverter) Init(parser gen.Parser, config gen.Config) {
	ce := singleValueCodeEmitter{
		Generated:         pgtypeBool,
		Target:            pointerOfBool,
		PgTypeName:        pgtypeBool.TypeName,
		ValuePropertyName: "Bool",
	}
	c.orchestrator = ce.toGeneratedTypeOrchestrator()
}

func (c *boolConverter) Info() gen.ConverterInfo {
	return gen.ConverterInfo{
		Name:                 "built-in lib pgtype/boolConverter",
		ShortForm:            "pgtype.Bool <-> [T *bool]",
		ShortFormDescription: "pgtype.Bool to T where T -> *bool is possible",
	}
}

func (c *boolConverter) CanConvert(ctx gen.LookupContext, targetType, sourceType types.Type) bool {
	return c.orchestrator.CanConvert(c, ctx, targetType, sourceType)
}

func (c *boolConverter) ConvertField(ctx gen.ConverterContext, target, source gen.Symbol, opts gen.ConverterOption) jen.Code {
	return ctx.Run(c, opts, func() jen.Code {
		return c.orchestrator.PerformConvert(c, ctx, target, source, opts)
	})
}

var _ gen.Converter = (*boolConverter)(nil)
