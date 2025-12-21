package grpc

import (
	"go/types"

	"github.com/dave/jennifer/jen"
	gen "github.com/toniphan21/go-mapper-gen"
)

const timestamppbPkgPath = "google.golang.org/protobuf/types/known/timestamppb"

type timestampConverter struct {
	standardLibTimeType types.Type
}

func (c *timestampConverter) Init(parser gen.Parser, config gen.Config) {
	pkg := types.NewPackage("time", "time")
	obj := types.NewTypeName(0, pkg, "Time", nil)

	c.standardLibTimeType = types.NewNamed(obj, new(types.Struct), nil)
}

func (c *timestampConverter) Info() gen.ConverterInfo {
	return gen.ConverterInfo{
		Name:                 "built-in lib grpc/timestampConverter",
		ShortForm:            "*timestamppb.Timestamp <-> [T time.Time]",
		ShortFormDescription: "protobuf timestamppb.Timestamp to T where T <=> time.Time is possible",
	}
}

func (c *timestampConverter) CanConvert(ctx gen.ConverterContext, targetType, sourceType types.Type) bool {
	if c.isTimestampPointer(sourceType) {
		return c.isTime(targetType) || c.isConvertibleFromTime(ctx, targetType)
	}

	if c.isTimestampPointer(targetType) {
		return c.isTime(sourceType) || c.isConvertibleToTime(ctx, sourceType)
	}
	return false
}

func (c *timestampConverter) ConvertField(ctx gen.ConverterContext, target, source gen.Symbol, opts gen.ConverterOption) jen.Code {
	return ctx.Run(c, opts, func() jen.Code {
		switch {
		// time.Time -> *timestamppb.Timestamp
		case c.isTime(source.Type) && c.isTimestampPointer(target.Type):
			// simply use timestamppb.New()
			return target.Expr().Op("=").Qual(timestamppbPkgPath, "New").Params(source.Expr())

		// T -> *timestamppb.Timestamp, where T -> time.Time is possible
		case c.isConvertibleToTime(ctx, source.Type) && c.isTimestampPointer(target.Type):
			oc, ok := ctx.LookUp(c, c.standardLibTimeType, source.Type)
			if !ok {
				return nil
			}

			// first convert T -> time.Time
			varName := ctx.NextVarName()
			code := jen.Line().Var().Id(varName).Add(gen.GeneratorUtil.TypeToJenCode(c.standardLibTimeType)).Line()

			targetSymbol := gen.Symbol{VarName: varName, Type: c.standardLibTimeType}
			convertedCode := oc.ConvertField(ctx, targetSymbol, source, opts)
			if convertedCode == nil {
				return nil
			}
			code.Add(convertedCode).Line()

			// then convert time.Time -> *timestamppb.Timestamp by timestamppb.New()
			return code.Add(target.Expr().Op("=").Qual(timestamppbPkgPath, "New").Params(jen.Id(varName)))

		// *timestamppb.Timestamp -> time.Time
		case c.isTimestampPointer(source.Type) && c.isTime(target.Type):
			// we need to check nil then use *timestamppb.Timestamp.AsTime
			code := jen.If(source.Expr().Op("!=").Nil()).
				BlockFunc(func(g *jen.Group) {
					g.Add(target.Expr()).Op("=").Add(source.Expr().Dot("AsTime").Params())
				}).
				Else().
				BlockFunc(func(g *jen.Group) {
					code := g.Var().Id("zero").Add(gen.GeneratorUtil.TypeToJenCode(target.Type)).Line()
					code = code.Add(target.Expr()).Op("=").Id("zero")
				})
			return code

		// *timestamppb.Timestamp -> T, where T -> time.Time is possible
		case c.isTimestampPointer(source.Type) && c.isConvertibleFromTime(ctx, target.Type):
			oc, ok := ctx.LookUp(c, target.Type, c.standardLibTimeType)
			if !ok {
				return nil
			}

			// first convert *timestamppb.Timestamp to T by checking nil and use *timestamppb.Timestamp.AsTime
			varName := ctx.NextVarName()
			code := jen.Line().Var().Id(varName).Add(gen.GeneratorUtil.TypeToJenCode(c.standardLibTimeType)).Line()

			code = code.If(source.Expr().Op("!=").Nil()).
				BlockFunc(func(g *jen.Group) {
					g.Add(jen.Id(varName).Op("=").Add(source.Expr().Dot("AsTime").Params()))
				}).Line()

			// then convert T to time.Time
			sourceSymbol := gen.Symbol{VarName: varName, Type: c.standardLibTimeType}
			convertedCode := oc.ConvertField(ctx, target, sourceSymbol, opts)
			if convertedCode == nil {
				return nil
			}
			return code.Add(convertedCode).Line()

		default:
			return nil
		}
	})
}

func (c *timestampConverter) isTimestampPointer(t types.Type) bool {
	const typeName = "Timestamp"

	return gen.TypeUtil.IsPointerToNamedType(t, timestamppbPkgPath, typeName)
}

func (c *timestampConverter) isTime(t types.Type) bool {
	const pkgPath = "time"
	const typeName = "Time"

	return gen.TypeUtil.MatchNamedType(t, pkgPath, typeName)
}

func (c *timestampConverter) isConvertibleToTime(ctx gen.ConverterContext, t types.Type) bool {
	_, ok := ctx.LookUp(c, c.standardLibTimeType, t)

	return ok
}

func (c *timestampConverter) isConvertibleFromTime(ctx gen.ConverterContext, t types.Type) bool {
	_, ok := ctx.LookUp(c, t, c.standardLibTimeType)
	return ok
}

var _ gen.Converter = (*timestampConverter)(nil)
