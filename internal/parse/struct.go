package parse

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

func StructType(pkg *packages.Package, structName string) types.Type {
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		tn, ok := obj.(*types.TypeName)
		if !ok {
			continue
		}

		strct, ok := tn.Type().(*types.Named)
		if !ok {
			continue
		}

		if _, ok = strct.Underlying().(*types.Struct); !ok {
			continue
		}

		if tn.Name() == structName {
			return strct
		}
	}
	return nil
}

func Struct(pkg *packages.Package, structName string) *ast.StructType {
	for _, v := range pkg.Syntax {
		for _, decl := range v.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				tspec := spec.(*ast.TypeSpec)
				if tspec.Name.Name != structName {
					continue
				}

				st, ok := tspec.Type.(*ast.StructType)
				if !ok {
					continue
				}
				return st

			}
		}
	}
	return nil
}

func StructFields(pkg *packages.Package, st *ast.StructType) map[string]types.Type {
	result := make(map[string]types.Type)
	for _, field := range st.Fields.List {
		fieldType := pkg.TypesInfo.TypeOf(field.Type)
		fieldName := ""
		if len(field.Names) > 0 {
			fieldName = field.Names[0].Name
		} else {
			// embedded struct, do not flatten fields
			if ptr, ok := fieldType.(*types.Pointer); ok {
				fieldType = ptr.Elem()
			}

			if named, ok := fieldType.(*types.Named); ok {
				fieldName = named.Obj().Name()
			}
		}
		result[fieldName] = fieldType
	}
	return result
}

func IsInvalidType(t types.Type) bool {
	return t == nil || t == types.Typ[types.Invalid]
}
