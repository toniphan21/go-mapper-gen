package gomappergen

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/toniphan21/go-mapper-gen/internal/util"
	"golang.org/x/tools/go/packages"
)

type StructInfo struct {
	Type   types.Type
	Fields map[string]StructFieldInfo
}

type StructFieldInfo struct {
	Name       string
	Getter     *string
	Tag        *string
	Comment    *string
	Doc        *string
	Type       types.Type
	Index      int
	IsExported bool
}

type FuncInfo struct {
	Name        string
	PackagePath string
	Params      []types.Type
	Results     []types.Type
}

type Parser interface {
	SourceDir() string

	SourcePackages() []*packages.Package

	FindStruct(pkgPath string, name string) (StructInfo, bool)

	FindFunction(pkgPath string, name string) (FuncInfo, bool)

	FindVariableMethods(pkgPath string, name string) []FuncInfo
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
		dir:            dir,
		config:         cfg,
		sourcePackages: pkgs,
	}, nil
}

type parserImpl struct {
	dir            string
	config         *packages.Config
	sourcePackages []*packages.Package
}

func (p *parserImpl) SourceDir() string {
	return p.dir
}

func (p *parserImpl) SourcePackages() []*packages.Package {
	return p.sourcePackages
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
		if pkg.PkgPath != pkgPath {
			continue
		}

		if len(pkg.Errors) == 0 {
			return p.findStructFromPkg(pkg, name)
		}

		for _, v := range pkg.Errors {
			fmt.Println(util.ColorYellow(fmt.Sprintf("Warning: %v", v)))
		}
	}
	return StructInfo{}, false
}

func (p *parserImpl) FindFunction(pkgPath string, name string) (FuncInfo, bool) {
	for _, pkg := range p.sourcePackages {
		if pkg.PkgPath == pkgPath {
			return p.findFunctionFromPkg(pkg, name)
		}
	}

	pkgs, err := packages.Load(p.config, pkgPath)
	if err != nil {
		return FuncInfo{}, false
	}

	for _, pkg := range pkgs {
		if pkg.PkgPath == pkgPath && len(pkg.Errors) == 0 {
			return p.findFunctionFromPkg(pkg, name)
		}
	}
	return FuncInfo{}, false
}

func (p *parserImpl) FindVariableMethods(pkgPath string, variableName string) []FuncInfo {
	for _, pkg := range p.sourcePackages {
		if pkg.PkgPath == pkgPath {
			return p.findVariableMethodsFromPkg(pkg, variableName)
		}
	}

	pkgs, err := packages.Load(p.config, pkgPath)
	if err != nil {
		return nil
	}

	for _, pkg := range pkgs {
		if pkg.PkgPath == pkgPath && len(pkg.Errors) == 0 {
			return p.findVariableMethodsFromPkg(pkg, variableName)
		}
	}
	return nil
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

	fields := p.structFields(pkg, structAST, structType)

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

func (p *parserImpl) structFields(pkg *packages.Package, sa *ast.StructType, st types.Type) map[string]StructFieldInfo {
	fields := make(map[string]StructFieldInfo)
	for idx, field := range sa.Fields.List {
		fieldType := pkg.TypesInfo.TypeOf(field.Type)
		fieldName := ""
		isExported := false

		if len(field.Names) > 0 {
			fieldName = field.Names[0].Name
			isExported = field.Names[0].IsExported()
		} else {
			isExported = ast.IsExported(p.getTypeNameFromExpr(field.Type))

			// embedded struct, do not flatten fields
			if ptr, ok := fieldType.(*types.Pointer); ok {
				fieldType = ptr.Elem()
			}

			if named, ok := fieldType.(*types.Named); ok {
				fieldName = named.Obj().Name()
			}
		}

		var getter *string
		if fn := p.findGetterForField(st, "Get"+fieldName, fieldType); fn != nil {
			v := "Get" + fieldName
			getter = &v
		}

		var tag *string
		if field.Tag != nil {
			tag = &field.Tag.Value
		}

		var comment *string
		if field.Comment != nil {
			v := field.Comment.Text()
			comment = &v
		}

		var doc *string
		if field.Doc != nil {
			v := field.Doc.Text()
			doc = &v
		}

		fields[fieldName] = StructFieldInfo{
			Name:       fieldName,
			Getter:     getter,
			Tag:        tag,
			Comment:    comment,
			Doc:        doc,
			Type:       fieldType,
			Index:      idx,
			IsExported: isExported,
		}
	}
	return fields
}

func (p *parserImpl) findGetterForField(t types.Type, name string, returnedType types.Type) *types.Func {
	named, ok := t.(*types.Named)
	if !ok {
		if ptr, ok := t.(*types.Pointer); ok {
			named, _ = ptr.Elem().(*types.Named)
		}
	}
	if named == nil {
		return nil
	}

	for i := 0; i < named.NumMethods(); i++ {
		m := named.Method(i)

		if m.Name() != name {
			continue
		}

		sig := m.Type().(*types.Signature)
		if sig.Params().Len() != 0 {
			continue
		}

		if sig.Results().Len() != 1 {
			continue
		}

		if TypeUtil.IsIdentical(sig.Results().At(0).Type(), returnedType) {
			return m
		}
	}
	return nil
}

func (p *parserImpl) getTypeNameFromExpr(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return t.Sel.Name
	case *ast.StarExpr:
		return p.getTypeNameFromExpr(t.X)
	default:
		return ""
	}
}

func (p *parserImpl) findFunctionFromPkg(pkg *packages.Package, name string) (FuncInfo, bool) {
	scope := pkg.Types.Scope()
	obj := scope.Lookup(name)

	if obj == nil {
		return FuncInfo{}, false
	}

	fn, ok := obj.(*types.Func)
	if !ok {
		return FuncInfo{}, false
	}

	signature, ok := fn.Type().(*types.Signature)
	if !ok {
		return FuncInfo{}, false
	}

	params := p.getTypesFromTuple(signature.Params())
	results := p.getTypesFromTuple(signature.Results())

	return FuncInfo{
		Name:        fn.Name(),
		PackagePath: pkg.PkgPath,
		Params:      params,
		Results:     results,
	}, true
}

func (p *parserImpl) findVariableMethodsFromPkg(pkg *packages.Package, variableName string) []FuncInfo {
	obj := pkg.Types.Scope().Lookup(variableName)
	if obj == nil {
		return nil
	}

	v, ok := obj.(*types.Var)
	if !ok {
		return nil
	}

	methodSet := types.NewMethodSet(v.Type())

	var methods []FuncInfo
	for i := 0; i < methodSet.Len(); i++ {
		methodObj := methodSet.At(i).Obj()

		fn, ok := methodObj.(*types.Func)
		if !ok {
			continue
		}

		sig, ok := fn.Type().(*types.Signature)
		if !ok {
			continue
		}

		methods = append(methods, FuncInfo{
			Name:        fn.Name(),
			PackagePath: pkg.PkgPath,
			Params:      p.getTypesFromTuple(sig.Params()),
			Results:     p.getTypesFromTuple(sig.Results()),
		})
	}

	return methods
}

func (p *parserImpl) getTypesFromTuple(tup *types.Tuple) []types.Type {
	if tup == nil {
		return nil
	}
	var typeList []types.Type
	for i := 0; i < tup.Len(); i++ {
		typeList = append(typeList, tup.At(i).Type())
	}
	return typeList
}

var _ Parser = (*parserImpl)(nil)
