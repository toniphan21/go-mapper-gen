package gomappergen

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type StructInfo struct {
	Type   types.Type
	Fields map[string]StructFieldInfo
}

type StructFieldInfo struct {
	Type  types.Type
	Index int
}

type Parser interface {
	SourcePackages() []*packages.Package

	FindStruct(pkgPath string, name string) (StructInfo, bool)
}

func DefaultParser(dir string) (Parser, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedDeps | packages.NeedImports,
		Fset: token.NewFileSet(),
		Dir:  dir,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, err
	}

	return &parserImpl{
		config:         cfg,
		sourcePackages: pkgs,
	}, nil
}

type parserImpl struct {
	config         *packages.Config
	sourcePackages []*packages.Package
}

func (p *parserImpl) FindStruct(pkgPath string, name string) (StructInfo, bool) {
	for _, pkg := range p.sourcePackages {
		if pkg.PkgPath == pkgPath {
			return p.findStructFromPkg(pkg, name)
		}
	}

	pkgs, err := packages.Load(p.config, pkgPath)
	if err != nil {
		return StructInfo{}, false
	}

	for _, pkg := range pkgs {
		if pkg.PkgPath == pkgPath && len(pkg.Errors) == 0 {
			return p.findStructFromPkg(pkg, name)
		}
	}
	return StructInfo{}, false
}

func (p *parserImpl) findStructFromPkg(pkg *packages.Package, name string) (StructInfo, bool) {
	structAST := p.findStructAST(pkg, name)
	if structAST == nil {
		return StructInfo{}, false
	}

	structType := p.structType(pkg, name)
	if structType == nil {
		return StructInfo{}, false
	}

	fields := p.structFields(pkg, structAST)

	return StructInfo{
		Type:   structType,
		Fields: fields,
	}, true
}

func (p *parserImpl) structType(pkg *packages.Package, structName string) types.Type {
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

func (p *parserImpl) findStructAST(pkg *packages.Package, structName string) *ast.StructType {
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

func (p *parserImpl) structFields(pkg *packages.Package, st *ast.StructType) map[string]StructFieldInfo {
	fields := make(map[string]StructFieldInfo)
	for idx, field := range st.Fields.List {
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
		fields[fieldName] = StructFieldInfo{
			Type:  fieldType,
			Index: idx,
		}
	}
	return fields
}

func (p *parserImpl) SourcePackages() []*packages.Package {
	return p.sourcePackages
}

var _ Parser = (*parserImpl)(nil)
