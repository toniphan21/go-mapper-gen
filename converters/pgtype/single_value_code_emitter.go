package pgtype

import (
	"github.com/dave/jennifer/jen"
	gen "github.com/toniphan21/go-mapper-gen"
)

type singleValueCodeEmitter struct {
	Generated         gen.TypeInfo
	Target            gen.TypeInfo
	PgTypeName        string
	ValuePropertyName string
	ValidPropertyName *string
}

func (e *singleValueCodeEmitter) toGeneratedTypeOrchestrator() gen.GeneratedTypeOrchestrator {
	return gen.GeneratedTypeOrchestrator{
		Generated:                e.Generated,
		Target:                   e.Target,
		GeneratedToTarget:        e.pgtypeToTarget,
		GeneratedToTargetToOther: e.pgtypeToTargetToOther,
		TargetToGenerated:        e.targetToPgType,
		OtherToTargetToGenerated: e.otherToTargetToPGType,
	}
}

func (e *singleValueCodeEmitter) pgtypeToTarget(ctx gen.ConverterContext, target, source gen.Symbol, opts gen.ConverterOption) jen.Code {
	/** generated code:
	if {source}.[[ValidPropertyName]] {
		{target} = &{source}.[[ValuePropertyName]]
	}
	*/
	validName := "Valid"
	if e.ValidPropertyName != nil {
		validName = *e.ValidPropertyName
	}

	return jen.If(source.Expr().Dot(validName)).BlockFunc(func(g *jen.Group) {
		g.Add(target.Expr()).Op("=").Op("&").Add(source.Expr().Dot(e.ValuePropertyName))
	})
}

func (e *singleValueCodeEmitter) pgtypeToTargetToOther(ctx gen.ConverterContext, target, source gen.Symbol, oc gen.Converter, opts gen.ConverterOption) jen.Code {
	/** generated code:
	var v0 [[Target]]
	if {source}.[[ValidPropertyName]] {
		v0 = &{source}.[[ValuePropertyName]]
	}

	// other converter converts v0 to target
	*/
	supportedType := e.Target.ToType()
	validName := "Valid"
	if e.ValidPropertyName != nil {
		validName = *e.ValidPropertyName
	}

	// first convert pgtype.Int2 to *int16 hold in a temporary variable
	varName := ctx.NextVarName()
	code := jen.Line().Var().Id(varName).Add(gen.GeneratorUtil.TypeToJenCode(supportedType)).Line()

	code = code.If(source.Expr().Dot(validName)).BlockFunc(func(g *jen.Group) {
		g.Add(jen.Id(varName)).Op("=").Op("&").Add(source.Expr().Dot(e.ValuePropertyName))
	}).Line()

	// then call other converter to convert the variable to target
	sourceSymbol := gen.Symbol{VarName: varName, Type: supportedType}
	convertedCode := oc.ConvertField(ctx, target, sourceSymbol, opts)
	if convertedCode == nil {
		return nil
	}
	return code.Add(convertedCode).Line()
}

func (e *singleValueCodeEmitter) targetToPgType(ctx gen.ConverterContext, target, source gen.Symbol, opts gen.ConverterOption) jen.Code {
	/** generated code:
	if {source} != nil {
		{target} = pgtype.[[PgTypeName]]{[[ValuePropertyName]]: *{source}, [[ValidPropertyName]]: true}
	}
	*/
	validName := "Valid"
	if e.ValidPropertyName != nil {
		validName = *e.ValidPropertyName
	}

	code := jen.If(source.Expr().Op("!=").Nil())
	code = code.BlockFunc(func(g *jen.Group) {
		g.Add(target.Expr()).Op("=").Add(
			jen.Qual(pgtypePkgPath, e.PgTypeName).Values(jen.DictFunc(func(d jen.Dict) {
				d[jen.Id(e.ValuePropertyName)] = jen.Op("*").Add(source.Expr())
				d[jen.Id(validName)] = jen.Lit(true)
			})),
		)
	})

	return code
}

func (e *singleValueCodeEmitter) otherToTargetToPGType(ctx gen.ConverterContext, target, source gen.Symbol, oc gen.Converter, opts gen.ConverterOption) jen.Code {
	/** generated code:
	var v0 [[Target]]
	// other converter converts source to v0

	if v0 != nil {
		{target} = pgtype.[[PgTypeName]]{[[ValuePropertyName]]: *v0, [[ValidPropertyName]]: true}
	}
	*/
	supportedType := e.Target.ToType()
	validName := "Valid"
	if e.ValidPropertyName != nil {
		validName = *e.ValidPropertyName
	}

	// first convert source to [[Target]] hold in a temporary variable
	varName := ctx.NextVarName()
	code := jen.Line().Var().Id(varName).Add(gen.GeneratorUtil.TypeToJenCode(supportedType)).Line()
	targetSymbol := gen.Symbol{VarName: varName, Type: supportedType, Metadata: gen.SymbolMetadata{IsVariable: true}}
	convertedCode := oc.ConvertField(ctx, targetSymbol, source, opts)
	if convertedCode == nil {
		return nil
	}
	code.Add(convertedCode).Line()

	// then convert variable to target
	code = code.Add(jen.If(jen.Id(varName).Op("!=").Nil()))
	code = code.BlockFunc(func(g *jen.Group) {
		g.Add(target.Expr()).Op("=").Add(
			jen.Qual(pgtypePkgPath, e.PgTypeName).Values(jen.DictFunc(func(d jen.Dict) {
				d[jen.Id(e.ValuePropertyName)] = jen.Op("*").Add(jen.Id(varName))
				d[jen.Id(validName)] = jen.Lit(true)
			})),
		)
	})
	return code
}
