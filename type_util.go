package gomappergen

import "go/types"

type typeUtil struct{}

// IsPointerOfType reports whether x is a pointer type whose element type
// is identical to y. It returns true when x is a *T and y is T, according
// to types.Identical, and false otherwise.
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
	return types.Identical(ptr.Elem(), y)
}

func (u *typeUtil) IsInterface(t types.Type) bool {
	_, ok := t.Underlying().(*types.Interface)
	return ok
}

var TypeUtil = &typeUtil{}
