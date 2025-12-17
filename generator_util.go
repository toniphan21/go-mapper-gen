package gomappergen

import (
	"go/types"
	"strconv"
	"strings"

	"github.com/dave/jennifer/jen"
)

const CommentWidth = 80

type genUtil struct {
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

func (g *genUtil) SimpleName(t types.Type) string {
	switch tt := t.(type) {
	case *types.Named:
		return tt.Obj().Name()
	case *types.Basic:
		return tt.Name()
	case *types.Pointer:
		return g.SimpleName(tt.Elem())
	case *types.Slice:
		return "[]" + g.SimpleName(tt.Elem())
	case *types.Array:
		return "[" + strconv.FormatInt(tt.Len(), 10) + "]" + g.SimpleName(tt.Elem())
	case *types.Map:
		return "map[" + g.SimpleName(tt.Key()) + "]" + g.SimpleName(tt.Elem())
	case *types.Chan:
		return "chan " + g.SimpleName(tt.Elem())
	case *types.Interface:
		return "interface{}"
	case *types.Signature:
		return "func"
	default:
		return t.String()
	}
}

func (g *genUtil) WrapComment(comment string) jen.Code {
	lines := g.WrapText(comment, CommentWidth)
	l := len(lines)
	if l == 1 {
		return jen.Comment(lines[0])
	}

	var code = jen.Add()
	for i, line := range lines {
		code = code.Add(jen.Comment(line))
		if i != l-1 {
			code = code.Line()
		}
	}
	return code
}

func (g *genUtil) WrapText(s string, width int) []string {
	var lines []string
	words := strings.Fields(s)

	if len(words) == 0 {
		return []string{""}
	}

	var current string
	for _, word := range words {
		if len(current)+len(word)+1 > width {
			lines = append(lines, strings.TrimSpace(current))
			current = word
		} else {
			if current == "" {
				current = word
			} else {
				current += " " + word
			}
		}
	}
	if current != "" {
		lines = append(lines, current)
	}

	return lines
}

var GeneratorUtil = &genUtil{}
