package gomappergen

import "go/types"

type typeUtil struct{}

// IsPointerOfType reports whether x is a pointer type whose element type
// is identical to y. It returns true when x is a *T and y is T, according
// to TypeUtil.IsIdentical, and false otherwise.
//
// Examples:
//
//	*string vs string → true
//	*int vs int       → true
//	*int vs string    → false
//	string vs string  → false (not a pointer)
func (u *typeUtil) IsPointerOfType(x, y types.Type) bool {
	ptr, ok := x.(*types.Pointer)
	if !ok {
		return false
	}
	return u.IsIdentical(ptr.Elem(), y)
}

func (u *typeUtil) IsInterface(t types.Type) bool {
	_, ok := t.Underlying().(*types.Interface)
	return ok
}

func (u *typeUtil) IsSlice(t types.Type) (types.Type, bool) {
	s, ok := t.Underlying().(*types.Slice)
	if !ok {
		return nil, false
	}
	return s.Elem(), true
}

func (u *typeUtil) IsPointerToNamedType(t types.Type, pkgPath, typeName string) bool {
	ptr, ok := t.(*types.Pointer)
	if !ok {
		return false
	}
	return u.MatchNamedType(ptr.Elem(), pkgPath, typeName)
}

func (u *typeUtil) MatchNamedType(t types.Type, pkgPath, typeName string) bool {
	if t == nil {
		return false
	}

	named, ok := t.(*types.Named)
	if !ok {
		return false
	}

	obj := named.Obj()

	actualPkgPath := ""
	if obj.Pkg() != nil {
		actualPkgPath = obj.Pkg().Path()
	}
	return actualPkgPath == pkgPath && obj.Name() == typeName
}

func (u *typeUtil) IsIdentical(t1 types.Type, t2 types.Type) bool {
	n1, ok1 := t1.(*types.Named)
	n2, ok2 := t2.(*types.Named)
	if !ok1 || !ok2 {
		// If they aren't named, they might be Basic (string, int)
		// types.Identical actually works fine for Basic types!
		return types.Identical(t1, t2)
	}

	obj1 := n1.Obj()
	obj2 := n2.Obj()

	path1, path2 := "", ""
	if obj1.Pkg() != nil {
		path1 = obj1.Pkg().Path()
	}
	if obj2.Pkg() != nil {
		path2 = obj2.Pkg().Path()
	}

	return path1 == path2 && obj1.Name() == obj2.Name()
}

func (u *typeUtil) MakeNamedType(pkgPath, pkgName, typeName string) types.Type {
	pkg := types.NewPackage(pkgPath, pkgName)
	obj := types.NewTypeName(0, pkg, typeName, nil)

	return types.NewNamed(obj, new(types.Struct), nil)
}

var TypeUtil = &typeUtil{}
