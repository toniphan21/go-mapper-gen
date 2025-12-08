//go:generate rm -rf internal/config/internal/gen
//go:generate pkl-gen-go pkl/Config.pkl
package gomappergen

type Output struct {
	PkgName      string
	FileName     string
	TestFileName string
}

type Config struct {
	Output             Output
	InterfaceName      string
	ImplementationName string
	ConstructorName    string
	Structs            []Struct
	GenerateGoDoc      bool
	Ignored            bool
}

type Struct struct {
	TargetPkgPath    string
	TargetStructName string
	SourcePkgPath    string
	SourceStructName string

	SourceToTargetFuncName   string
	SourceFromTargetFuncName string
	DecorateFuncName         string

	Pointer        Pointer
	FieldNameMatch FieldNameMatch

	GenerateSourceToTarget   bool
	GenerateSourceFromTarget bool
}

type Pointer int

const (
	PointerNone Pointer = iota
	PointerSourceOnly
	PointerTargetOnly
	PointerBoth
)

type FieldNameMatch int

const (
	FieldNameMatchIgnoreCase FieldNameMatch = iota
	FieldNameMatchExact
)
