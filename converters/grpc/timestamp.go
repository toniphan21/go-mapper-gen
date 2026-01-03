package grpc

import (
	"go/types"

	"github.com/dave/jennifer/jen"
	gen "github.com/toniphan21/go-mapper-gen"
)

const timestamppbPkgPath = "google.golang.org/protobuf/types/known/timestamppb"

var standardTimeTypeInfo = gen.TypeInfo{PkgPath: "time", PkgName: "time", TypeName: "Time"}
var timestampTypeInfo = gen.TypeInfo{
	PkgPath:   timestamppbPkgPath,
	PkgName:   "timestamp",
	TypeName:  "Timestamp",
	IsPointer: true,
}

type timestampConverter struct {
	orchestrator gen.GeneratedTypeOrchestrator
}

func (c *timestampConverter) Init(parser gen.Parser, config gen.Config) {
	c.orchestrator = gen.GeneratedTypeOrchestrator{
		Generated:                timestampTypeInfo,
		Target:                   standardTimeTypeInfo,
		GeneratedToTarget:        c.timestampToTime,
		GeneratedToTargetToOther: c.timestampToTimeToOther,
		TargetToGenerated:        c.timeToTimestamp,
		OtherToTargetToGenerated: c.otherToTimeToTimestamp,
	}
}

func (c *timestampConverter) Info() gen.ConverterInfo {
	return gen.ConverterInfo{
		Name:                 "built-in lib grpc/timestampConverter",
		ShortForm:            "*timestamppb.Timestamp <-> [T time.Time]",
		ShortFormDescription: "protobuf *timestamppb.Timestamp to T where T -> time.Time is possible",
	}
}

func (c *timestampConverter) CanConvert(ctx gen.LookupContext, targetType, sourceType types.Type) bool {
	return c.orchestrator.CanConvert(c, ctx, targetType, sourceType)
}

func (c *timestampConverter) ConvertField(ctx gen.ConverterContext, target, source gen.Symbol) jen.Code {
	return ctx.Run(c, func() jen.Code {
		return c.orchestrator.PerformConvert(c, ctx, target, source)
	})
}

// timestampToTime handles *timestamppb.Timestamp -> time.Time
func (c *timestampConverter) timestampToTime(ctx gen.ConverterContext, target, source gen.Symbol) jen.Code {
	// we need to check nil then use *timestamppb.Timestamp.AsTime
	return jen.If(source.Expr().Op("!=").Nil()).
		BlockFunc(func(g *jen.Group) {
			g.Add(target.Expr()).Op("=").Add(source.Expr().Dot("AsTime").Params())
		})
}

// timestampToTimeToOther handles *timestamppb.Timestamp -> T, where time.Time -> T is possible
func (c *timestampConverter) timestampToTimeToOther(ctx gen.ConverterContext, target, source gen.Symbol, oc gen.Converter) jen.Code {
	standardTimeType := standardTimeTypeInfo.ToType()

	// first convert *timestamppb.Timestamp to time.Time by checking nil and use *timestamppb.Timestamp.AsTime
	varName := ctx.NextVarName()
	code := jen.Line().Var().Id(varName).Add(gen.GeneratorUtil.TypeToJenCode(standardTimeType)).Line()

	code = code.If(source.Expr().Op("!=").Nil()).
		BlockFunc(func(g *jen.Group) {
			g.Add(jen.Id(varName).Op("=").Add(source.Expr().Dot("AsTime").Params()))
		}).Line()

	// then convert time.Time to T
	sourceSymbol := gen.Symbol{VarName: varName, Type: standardTimeType}
	convertedCode := oc.ConvertField(ctx, target, sourceSymbol)
	if convertedCode == nil {
		return nil
	}
	return code.Add(convertedCode).Line()
}

// timeToTimestamp handles time.Time -> *timestamppb.Timestamp
func (c *timestampConverter) timeToTimestamp(ctx gen.ConverterContext, target, source gen.Symbol) jen.Code {
	// simply use timestamppb.New()
	return target.Expr().Op("=").Qual(timestamppbPkgPath, "New").Params(source.Expr())
}

// otherToTimeToTimestamp handles T -> *timestamppb.Timestamp, where T -> time.Time is possible
func (c *timestampConverter) otherToTimeToTimestamp(ctx gen.ConverterContext, target, source gen.Symbol, oc gen.Converter) jen.Code {
	standardTimeType := standardTimeTypeInfo.ToType()

	// first convert T -> time.Time
	varName := ctx.NextVarName()
	code := jen.Line().Var().Id(varName).Add(gen.GeneratorUtil.TypeToJenCode(standardTimeType)).Line()

	targetSymbol := gen.Symbol{VarName: varName, Type: standardTimeType, Metadata: gen.SymbolMetadata{IsVariable: true, HasZeroValue: true}}
	convertedCode := oc.ConvertField(ctx, targetSymbol, source)
	if convertedCode == nil {
		return nil
	}
	code.Add(convertedCode).Line()

	// then convert time.Time -> *timestamppb.Timestamp by timestamppb.New()
	return code.Add(target.Expr().Op("=").Qual(timestamppbPkgPath, "New").Params(jen.Id(varName)))
}

var _ gen.Converter = (*timestampConverter)(nil)
