package sql

import (
	"go/types"

	"github.com/dave/jennifer/jen"
	gen "github.com/toniphan21/go-mapper-gen"
)

const sqlPkgPath = "database/sql"

type baseConverter[T, V any] struct {
	ValuePropertyName string
	ValidPropertyName *string

	Name                 string
	ShortForm            string
	ShortFormDescription string

	generatedType T
	targetType    V
	orchestrator  gen.GeneratedTypeOrchestrator

	targetedType types.Type
	sqlTypeName  string
}

func (b *baseConverter[T, V]) Init(parser gen.Parser, config gen.Config) {
	generated := gen.MakeTypeInfo(b.generatedType)
	target := gen.MakeTypeInfo(b.targetType)
	b.sqlTypeName = generated.TypeName
	b.targetedType = target.ToType()
	b.orchestrator = gen.GeneratedTypeOrchestrator{
		Generated:                generated,
		Target:                   target,
		GeneratedToTarget:        b.sqlTypeToTarget,
		GeneratedToTargetToOther: b.sqlTypeToTargetToOther,
		TargetToGenerated:        b.targetToSQLType,
		OtherToTargetToGenerated: b.otherToTargetToSQLType,
	}
}

func (b *baseConverter[T, V]) Info() gen.ConverterInfo {
	return gen.ConverterInfo{
		Name:                 b.Name,
		ShortForm:            b.ShortForm,
		ShortFormDescription: b.ShortFormDescription,
	}
}

func (b *baseConverter[T, V]) CanConvert(ctx gen.LookupContext, targetType, sourceType types.Type) bool {
	return b.orchestrator.CanConvert(b, ctx, targetType, sourceType)
}

func (b *baseConverter[T, V]) ConvertField(ctx gen.ConverterContext, target, source gen.Symbol) jen.Code {
	return ctx.Run(b, func() jen.Code {
		return b.orchestrator.PerformConvert(b, ctx, target, source)
	})
}

func (b *baseConverter[T, V]) sqlTypeToTarget(ctx gen.ConverterContext, target, source gen.Symbol) jen.Code {
	/** generated code:
	if {source}.[[ValidPropertyName]] {
		{target} = &{source}.[[ValuePropertyName]]
	}
	*/
	validName := "Valid"
	if b.ValidPropertyName != nil {
		validName = *b.ValidPropertyName
	}

	return jen.If(source.Expr().Dot(validName)).BlockFunc(func(g *jen.Group) {
		g.Add(target.Expr()).Op("=").Op("&").Add(source.Expr().Dot(b.ValuePropertyName))
	})
}

func (b *baseConverter[T, V]) sqlTypeToTargetToOther(ctx gen.ConverterContext, target, source gen.Symbol, oc gen.Converter) jen.Code {
	/** generated code:
	var v0 [[Target]]
	if {source}.[[ValidPropertyName]] {
		v0 = &{source}.[[ValuePropertyName]]
	}

	// other converter converts v0 to target
	*/
	validName := "Valid"
	if b.ValidPropertyName != nil {
		validName = *b.ValidPropertyName
	}

	// first convert b.Generated to b.Target hold in a temporary variable
	varName := ctx.NextVarName()
	code := jen.Line().Var().Id(varName).Add(gen.GeneratorUtil.TypeToJenCode(b.targetedType)).Line()

	code = code.If(source.Expr().Dot(validName)).BlockFunc(func(g *jen.Group) {
		g.Add(jen.Id(varName)).Op("=").Op("&").Add(source.Expr().Dot(b.ValuePropertyName))
	}).Line()

	// then call other converter to convert the variable to target
	sourceSymbol := gen.Symbol{VarName: varName, Type: b.targetedType}
	convertedCode := oc.ConvertField(ctx, target, sourceSymbol)
	if convertedCode == nil {
		return nil
	}
	return code.Add(convertedCode).Line()
}

func (b *baseConverter[T, V]) targetToSQLType(ctx gen.ConverterContext, target, source gen.Symbol) jen.Code {
	/** generated code:
	if {source} != nil {
		{target} = sql.[[sqlTypeName]]{[[ValuePropertyName]]: *{source}, [[ValidPropertyName]]: true}
	}
	*/
	validName := "Valid"
	if b.ValidPropertyName != nil {
		validName = *b.ValidPropertyName
	}

	code := jen.If(source.Expr().Op("!=").Nil())
	code = code.BlockFunc(func(g *jen.Group) {
		g.Add(target.Expr()).Op("=").Add(
			jen.Qual(sqlPkgPath, b.sqlTypeName).Values(jen.DictFunc(func(d jen.Dict) {
				d[jen.Id(b.ValuePropertyName)] = jen.Op("*").Add(source.Expr())
				d[jen.Id(validName)] = jen.Lit(true)
			})),
		)
	})

	return code
}

func (b *baseConverter[T, V]) otherToTargetToSQLType(ctx gen.ConverterContext, target, source gen.Symbol, oc gen.Converter) jen.Code {
	/** generated code:
	var v0 [[Target]]
	// other converter converts source to v0

	if v0 != nil {
		{target} = pgtype.[[sqlTypeName]]{[[ValuePropertyName]]: *v0, [[ValidPropertyName]]: true}
	}
	*/
	validName := "Valid"
	if b.ValidPropertyName != nil {
		validName = *b.ValidPropertyName
	}

	// first convert source to [[Target]] hold in a temporary variable
	varName := ctx.NextVarName()
	code := jen.Line().Var().Id(varName).Add(gen.GeneratorUtil.TypeToJenCode(b.targetedType)).Line()
	targetSymbol := gen.Symbol{VarName: varName, Type: b.targetedType, Metadata: gen.SymbolMetadata{IsVariable: true}}
	convertedCode := oc.ConvertField(ctx, targetSymbol, source)
	if convertedCode == nil {
		return nil
	}
	code.Add(convertedCode).Line()

	// then convert variable to target
	code = code.Add(jen.If(jen.Id(varName).Op("!=").Nil()))
	code = code.BlockFunc(func(g *jen.Group) {
		g.Add(target.Expr()).Op("=").Add(
			jen.Qual(sqlPkgPath, b.sqlTypeName).Values(jen.DictFunc(func(d jen.Dict) {
				d[jen.Id(b.ValuePropertyName)] = jen.Op("*").Add(jen.Id(varName))
				d[jen.Id(validName)] = jen.Lit(true)
			})),
		)
	})
	return code
}
