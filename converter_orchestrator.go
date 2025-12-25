package gomappergen

import (
	"go/types"

	"github.com/dave/jennifer/jen"
)

type orchestrator interface {
	CanConvert(c Converter, ctx LookupContext, targetType, sourceType types.Type) bool

	PerformConvert(c Converter, ctx ConverterContext, target, source Symbol, opts ConverterOption) jen.Code
}

type TypeInfo struct {
	PkgPath   string
	PkgName   string
	TypeName  string
	IsPointer bool
}

func (ti TypeInfo) ToType() types.Type {
	if ti.PkgPath == "" && ti.PkgName == "" {
		// basic type, no package
		basic := ti.findBasic(ti.TypeName)
		var typ types.Type
		if basic == nil {
			obj := types.NewTypeName(0, nil, ti.TypeName, nil)
			typ = types.NewNamed(obj, types.Typ[types.Invalid], nil)
		} else {
			typ = basic
		}

		if !ti.IsPointer {
			return typ
		}
		return types.NewPointer(typ)
	}

	// standard library or type in package
	pkg := types.NewPackage(ti.PkgPath, ti.PkgName)
	obj := types.NewTypeName(0, pkg, ti.TypeName, nil)

	typ := types.NewNamed(obj, types.Typ[types.Invalid], nil)
	if !ti.IsPointer {
		return typ
	}
	return types.NewPointer(typ)
}

func (ti TypeInfo) findBasic(name string) types.Type {
	obj := types.Universe.Lookup(name)
	if obj != nil {
		return obj.Type()
	}
	return nil
}

type standardRouting struct {
	route           standardRoute
	beforeConverter Converter
	afterConverter  Converter
}

type standardRoute int

const (
	standardRouteNone = iota

	standardRouteSourceToTarget
	standardRouteOtherToSourceToTarget
	standardRouteSourceToTargetToOther
	standardRouteOtherToSourceToTargetToOther

	standardRouteTargetToSource
	standardRouteOtherToTargetToSource
	standardRouteTargetToSourceToOther
	standardRouteOtherToTargetToSourceToOther
)

// StandardConversionOrchestrator provide a framework to a converter want to convert A -> B.
// it provides 8 routes check and invoke function to emit code.
//
// For example: Given you want to convert A -> B, there are 8 standard routes:
//  1. SourceToTarget				A -> B
//  2. OtherToSourceToTarget		T -> A -> B if T -> A possible
//  3. SourceToTargetToOther		A -> B -> T if B -> T possible
//  4. OtherToSourceToTargetToOther	T -> A -> B -> V if T -> A and B -> V possible
//
// flipped cases:
//
//  5. TargetToSource				B -> A
//  6. OtherToTargetToSource		T -> B -> A if T -> B possible
//  7. TargetToSourceToOther		B -> A -> T if A -> T possible
//  8. OtherToTargetToSourceToOther	T -> B -> A -> V if T -> B and A -> V possible
//
// Usage: In your Converter.Init():
//
// ```
//
//	c.orchestrator = gen.StandardConversionOrchestrator{
//			Source:                			A's type info,
//			Target:                			B's type info,
//			SourceToTarget:        			func to emit code when match route 1) A -> B
//			OtherToSourceToTarget: 			nil if you don't want route 2)
//			SourceToTargetToOther: 			func to emit code when match route 3) A -> B -> T
//			OtherToSourceToTargetToOther:	nil if you don't want route 4)
//			TargetToSource:        			func to emit code when match route 5) B -> A
//			OtherToTargetToSource: 			func to emit code when match route 6) T -> B -> A
//			TargetToSourceToOther: 			nil if you don't want route 7)
//			OtherToTargetToSourceToOther:	nil if you don't want route 8)
//	}
//
// ```
// Then in your Converter.CanConvert()
// ```
//
//	return c.orchestrator.CanConvert(c, ctx, targetType, sourceType)
//
// ```
// and in your Converter.ConvertField()
//
// ```
//
//	return ctx.Run(c, opts, func() jen.Code {
//		return c.orchestrator.PerformConvert(c, ctx, target, source, opts)
//	})
//
// ```
// There is a simpler version of StandardConversionOrchestrator called GeneratedTypeOrchestrator
// which one side is fixed, it means you only have 4 cases.
type StandardConversionOrchestrator struct {
	Source TypeInfo
	Target TypeInfo

	SourceToTarget               func(ctx ConverterContext, target, source Symbol, opts ConverterOption) jen.Code
	OtherToSourceToTarget        func(ctx ConverterContext, target, source Symbol, otherToSource Converter, opts ConverterOption) jen.Code
	SourceToTargetToOther        func(ctx ConverterContext, target, source Symbol, targetToOther Converter, opts ConverterOption) jen.Code
	OtherToSourceToTargetToOther func(ctx ConverterContext, target, source Symbol, otherToSource, targetToOther Converter, opts ConverterOption) jen.Code

	TargetToSource               func(ctx ConverterContext, target, source Symbol, opts ConverterOption) jen.Code
	OtherToTargetToSource        func(ctx ConverterContext, target, source Symbol, otherToTarget Converter, opts ConverterOption) jen.Code
	TargetToSourceToOther        func(ctx ConverterContext, target, source Symbol, sourceToOther Converter, opts ConverterOption) jen.Code
	OtherToTargetToSourceToOther func(ctx ConverterContext, target, source Symbol, otherToTarget, sourceToOther Converter, opts ConverterOption) jen.Code
}

func (o *StandardConversionOrchestrator) routing(c Converter, ctx LookupContext, source types.Type, target types.Type) standardRouting {
	expectedSource := o.Source.ToType()
	expectedTarget := o.Target.ToType()
	// expectedSource = A; expectedTarget = B; source = a; target = b

	// normal cases:
	// SourceToTarget: 					A => B			Aa => Bb		(check A == a & B == b)
	// OtherToSourceToTarget: 			T->A => B		a->A => Bb		(check B == b & a -> A possible)
	// SourceToTargetToOther:			A => B->T		Aa => B->b		(check A == a & B -> b possible)
	// OtherToSourceToTargetToOther:	T->A => B->V	a->A => B->b 	(check a -> A & B -> b possible)

	// flipped cases: (flipped cases actually flipped expectedSource <-> expectedTarget)
	// TargetToSource:					B => A				Ba => Ab		(check B == a & A == b)
	// OtherToTargetToSource:			T->B => A			a->B => Ab		(check A == b & a -> B possible)
	// TargetToSourceToOther:			B => A->T			Ba => A->b 		(check B == a & A -> b possible)
	// OtherToTargetToSourceToOther:	T->B => A->V		B->a => A->b 	(check B -> a & A -> b possible)

	if o.SourceToTarget != nil {
		if o.isTypesMatch(expectedSource, source) && o.isTypesMatch(expectedTarget, target) {
			return standardRouting{route: standardRouteSourceToTarget}
		}
	}

	if o.OtherToSourceToTarget != nil {
		if o.isTypesMatch(expectedTarget, target) {
			// ctx.LookUp is so expensive, it has to check last
			bc := o.findConverter(c, ctx, source, expectedSource)
			if bc != nil {
				return standardRouting{route: standardRouteOtherToSourceToTarget, beforeConverter: bc}
			}
		}
	}

	if o.SourceToTargetToOther != nil {
		if o.isTypesMatch(expectedSource, source) {
			ac := o.findConverter(c, ctx, expectedTarget, target)
			if ac != nil {
				return standardRouting{route: standardRouteSourceToTargetToOther, afterConverter: ac}
			}
		}
	}

	if o.OtherToSourceToTargetToOther != nil {
		ac := o.findConverter(c, ctx, source, expectedSource)
		if ac != nil {
			bc := o.findConverter(c, ctx, expectedTarget, target)
			if bc != nil {
				return standardRouting{route: standardRouteOtherToSourceToTargetToOther, beforeConverter: bc, afterConverter: ac}
			}
		}
	}

	// --- flip target -> source

	if o.TargetToSource != nil {
		if o.isTypesMatch(expectedTarget, source) && o.isTypesMatch(expectedSource, target) {
			return standardRouting{route: standardRouteTargetToSource}
		}
	}

	if o.OtherToTargetToSource != nil {
		if o.isTypesMatch(expectedSource, target) {
			bc := o.findConverter(c, ctx, source, expectedTarget)
			if bc != nil {
				return standardRouting{route: standardRouteOtherToTargetToSource, beforeConverter: bc}
			}
		}
	}

	if o.TargetToSourceToOther != nil {
		if o.isTypesMatch(expectedTarget, source) {
			ac := o.findConverter(c, ctx, expectedSource, target)
			if ac != nil {
				return standardRouting{route: standardRouteSourceToTargetToOther, afterConverter: ac}
			}
		}
	}

	if o.OtherToTargetToSourceToOther != nil {
		ac := o.findConverter(c, ctx, source, expectedTarget)
		if ac != nil {
			bc := o.findConverter(c, ctx, expectedSource, target)
			if bc != nil {
				return standardRouting{route: standardRouteOtherToTargetToSourceToOther, beforeConverter: bc, afterConverter: ac}
			}
		}
	}
	return standardRouting{route: standardRouteNone}
}

func (o *StandardConversionOrchestrator) PerformConvert(c Converter, ctx ConverterContext, target, source Symbol, opts ConverterOption) jen.Code {
	routing := o.routing(c, ctx, source.Type, target.Type)
	switch routing.route {
	case standardRouteSourceToTarget:
		return o.SourceToTarget(ctx, target, source, opts)

	case standardRouteOtherToSourceToTarget:
		return o.OtherToSourceToTarget(ctx, target, source, routing.beforeConverter, opts)

	case standardRouteSourceToTargetToOther:
		return o.SourceToTargetToOther(ctx, target, source, routing.afterConverter, opts)

	case standardRouteOtherToSourceToTargetToOther:
		return o.OtherToSourceToTargetToOther(ctx, target, source, routing.beforeConverter, routing.afterConverter, opts)

	case standardRouteTargetToSource:
		return o.TargetToSource(ctx, target, source, opts)

	case standardRouteOtherToTargetToSource:
		return o.OtherToTargetToSource(ctx, target, source, routing.beforeConverter, opts)

	case standardRouteTargetToSourceToOther:
		return o.TargetToSourceToOther(ctx, target, source, routing.afterConverter, opts)

	case standardRouteOtherToTargetToSourceToOther:
		return o.OtherToTargetToSourceToOther(ctx, target, source, routing.beforeConverter, routing.afterConverter, opts)

	default:
		return nil
	}
}

func (o *StandardConversionOrchestrator) CanConvert(c Converter, ctx LookupContext, targetType, sourceType types.Type) bool {
	return o.routing(c, ctx, sourceType, targetType).route != standardRouteNone
}

func (o *StandardConversionOrchestrator) isTypesMatch(a, b types.Type) bool {
	return TypeUtil.IsIdentical(a, b)
}

func (o *StandardConversionOrchestrator) findConverter(c Converter, l LookupContext, from, to types.Type) Converter {
	converter, _ := l.LookUp(c, to, from)

	return converter
}

var _ orchestrator = (*StandardConversionOrchestrator)(nil)

// GeneratedTypeOrchestrator is simpler version of StandardConversionOrchestrator which one side
// is fixed. It provides 4 ways of conversion
//
// For example: Given you want to convert A -> B, A is a generated type which usually provided
// via a generated tool such as grpc or sqlc. GeneratedTypeOrchestrator provides 4 routing ways:
//
//  1. GeneratedToTarget			A -> B
//  2. GeneratedToTargetToOther		A -> B -> T if B -> T possible
//  3. TargetToGenerated			B -> A
//  4. OtherToTargetToGenerated		T -> B -> A if T -> B possible
//
// Usage: In your Converter.Init():
//
// ```
//
//	 	// similar to StandardConversionOrchestrator, you can simply set nil if you don't want to match the route
//		c.orchestrator = gen.GeneratedTypeOrchestrator{
//				Generated:					Generated type, aka. A's type info,
//				Target:                		B's type info,
//				GeneratedToTarget:			func to emit code when match route 1) A -> B
//				GeneratedToTargetToOther:	func to emit code when match route 2) A -> B -> T
//				TargetToGenerated:			func to emit code when match route 3) B -> A
//				OtherToTargetToGenerated:	func to emit code when match route 4) T -> B -> A
//		}
//
// ```
// Then in your Converter.CanConvert()
// ```
//
//	return c.orchestrator.CanConvert(c, ctx, targetType, sourceType)
//
// ```
// and in your Converter.ConvertField()
//
// ```
//
//	return ctx.Run(c, opts, func() jen.Code {
//		return c.orchestrator.PerformConvert(c, ctx, target, source, opts)
//	})
//
// ```
// The GeneratedTypeOrchestrator actually just a wrapper of StandardConversionOrchestrator.
type GeneratedTypeOrchestrator struct {
	wrapee *StandardConversionOrchestrator

	Generated TypeInfo
	Target    TypeInfo

	GeneratedToTarget        func(ctx ConverterContext, target, source Symbol, opts ConverterOption) jen.Code
	GeneratedToTargetToOther func(ctx ConverterContext, target, source Symbol, targetToOther Converter, opts ConverterOption) jen.Code
	TargetToGenerated        func(ctx ConverterContext, target, source Symbol, opts ConverterOption) jen.Code
	OtherToTargetToGenerated func(ctx ConverterContext, target, source Symbol, otherToTarget Converter, opts ConverterOption) jen.Code
}

func (o *GeneratedTypeOrchestrator) toStandardConversionOrchestrator() *StandardConversionOrchestrator {
	if o.wrapee == nil {
		o.wrapee = &StandardConversionOrchestrator{
			Source:                o.Generated,
			Target:                o.Target,
			SourceToTarget:        o.GeneratedToTarget,
			SourceToTargetToOther: o.GeneratedToTargetToOther,
			TargetToSource:        o.TargetToGenerated,
			OtherToTargetToSource: o.OtherToTargetToGenerated,
		}
	}
	return o.wrapee
}

func (o *GeneratedTypeOrchestrator) CanConvert(c Converter, ctx LookupContext, targetType, sourceType types.Type) bool {
	return o.toStandardConversionOrchestrator().CanConvert(c, ctx, targetType, sourceType)
}

func (o *GeneratedTypeOrchestrator) PerformConvert(c Converter, ctx ConverterContext, target, source Symbol, opts ConverterOption) jen.Code {
	return o.toStandardConversionOrchestrator().PerformConvert(c, ctx, target, source, opts)
}

var _ orchestrator = (*GeneratedTypeOrchestrator)(nil)
