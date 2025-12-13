package gomappergen

import (
	"fmt"
	"go/types"

	"github.com/dave/jennifer/jen"
)

type genUtil struct {
}

func (g *genUtil) GenerateWithConverterOption(code jen.Code, opt ConverterOption, converterName string) jen.Code {
	if opt.EmitTraceComments {
		out := jen.Comment(fmt.Sprintf("%s generated code start", converterName)).Line()
		out.Add(code).Line()
		out.Add(jen.Comment(fmt.Sprintf("%s generated code end", converterName)))
		return out
	}
	return code
}

func (g *genUtil) TypeToJenCode(t types.Type) jen.Code {
	switch tt := t.(type) {
	case *types.Basic:
		if tt.Kind() == types.Invalid {
			return jen.Id("interface{}")
		}
		if tt.Kind() == types.UnsafePointer {
			return jen.Id("unsafe.Pointer")
		}
		if tt.Name() == "error" {
			return jen.Error()
		}
		return jen.Id(tt.Name())

	case *types.Pointer:
		return jen.Op("*").Add(g.TypeToJenCode(tt.Elem()))

	case *types.Named:
		obj := tt.Obj()
		pkg := obj.Pkg()
		if pkg != nil {
			return jen.Qual(pkg.Path(), obj.Name())
		}
		return jen.Id(obj.Name())

	case *types.Slice:
		return jen.Index().Add(g.TypeToJenCode(tt.Elem()))

	case *types.Array:
		return jen.Index(jen.Lit(tt.Len())).Add(g.TypeToJenCode(tt.Elem()))

	case *types.Map:
		return jen.Map(g.TypeToJenCode(tt.Key())).Add(g.TypeToJenCode(tt.Elem()))

	case *types.Chan:
		elemCode := g.TypeToJenCode(tt.Elem())
		switch tt.Dir() {
		case types.SendRecv:
			return jen.Chan().Add(elemCode)
		case types.SendOnly:
			return jen.Chan().Op("<-").Add(elemCode)
		case types.RecvOnly:
			return jen.Op("<-").Chan().Add(elemCode)
		}

	case *types.Signature:
		fn := jen.Func()
		fn.Params(g.funcGroupFromTuple(tt.Params())...)
		if tt.Results().Len() == 1 {
			fn.Params(g.funcGroupFromTuple(tt.Results())...)
		} else if tt.Results().Len() > 1 {
			fn.Params(g.funcGroupFromTuple(tt.Results())...)
		}
		return fn

	default:
		return jen.Id(tt.String())
	}
	return nil
}

func (g *genUtil) funcGroupFromTuple(t *types.Tuple) []jen.Code {
	var res []jen.Code
	for i := 0; i < t.Len(); i++ {
		v := t.At(i)
		typ := g.TypeToJenCode(v.Type())
		if v.Name() != "" {
			res = append(res, jen.Id(v.Name()).Add(typ))
		} else {
			res = append(res, typ)
		}
	}
	return res
}

var GeneratorUtil = &genUtil{}
