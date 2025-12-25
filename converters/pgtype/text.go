package pgtype

import (
	"go/types"

	"github.com/dave/jennifer/jen"
	gen "github.com/toniphan21/go-mapper-gen"
)

var pgtypeText = gen.TypeInfo{PkgPath: pgtypePkgPath, PkgName: "pgtype", TypeName: "Text"}
var pointerOfString = gen.TypeInfo{TypeName: "string", IsPointer: true}

type textConverter struct {
	orchestrator gen.GeneratedTypeOrchestrator
}

func (c *textConverter) Init(parser gen.Parser, config gen.Config) {
	ce := singleValueCodeEmitter{
		Generated:         pgtypeText,
		Target:            pointerOfString,
		PgTypeName:        pgtypeText.TypeName,
		ValuePropertyName: "String",
	}
	c.orchestrator = ce.toGeneratedTypeOrchestrator()
}

func (c *textConverter) Info() gen.ConverterInfo {
	return gen.ConverterInfo{
		Name:                 "built-in lib pgtype/textConverter",
		ShortForm:            "pgtype.Text <-> [T *string]",
		ShortFormDescription: "pgtype.Text to T where T -> *string is possible",
	}
}

func (c *textConverter) CanConvert(ctx gen.LookupContext, targetType, sourceType types.Type) bool {
	return c.orchestrator.CanConvert(c, ctx, targetType, sourceType)
}

func (c *textConverter) ConvertField(ctx gen.ConverterContext, target, source gen.Symbol, opts gen.ConverterOption) jen.Code {
	return ctx.Run(c, opts, func() jen.Code {
		return c.orchestrator.PerformConvert(c, ctx, target, source, opts)
	})
}

var _ gen.Converter = (*textConverter)(nil)
