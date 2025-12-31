// Code generated from Pkl module `gomappergen.mapper`. DO NOT EDIT.
package mapper

type Package interface {
	GetMode() string

	GetInterfaceName() string

	GetImplementationName() string

	GetConstructorName() string

	GetSourceToTargetFunctionName() string

	GetSourceFromTargetFunctionName() string

	GetDecoratorMode() string

	GetDecoratorInterfaceName() string

	GetDecoratorNoopName() string

	GetDecorateFunctionName() string

	GetTargetPkg() string

	GetSourcePkg() string

	GetStructs() map[string]Struct

	GetUseGetterIfAvailable() bool

	GetGenerateSourceToTarget() bool

	GetGenerateSourceFromTarget() bool

	GetGenerateGoDoc() bool
}

var _ Package = PackageImpl{}

// Base configuration for mapper code generation.
//
// This configuration controls *how* mappers are generated (types vs functions),
// naming conventions of generated symbols, decorator behavior, and package layout.
type PackageImpl struct {
	// Controls the overall generation strategy.
	//
	// - "types": Generates an interface, implementation struct, and constructor.
	// - "functions": Generates package-level mapping functions only (no interface or struct).
	//
	// Note: When mode = "functions", type-related naming options are ignored.
	Mode string `pkl:"mode"`

	// Name of the generated mapper interface.
	//
	// Used only when mode = "types".
	// Defaults to a non-exported name to avoid polluting consumer APIs.
	InterfaceName string `pkl:"interface_name"`

	// Name of the generated mapper implementation struct.
	//
	// Used only when mode = "types".
	ImplementationName string `pkl:"implementation_name"`

	// Name of the constructor function for the mapper implementation.
	//
	// Used only when mode = "types".
	ConstructorName string `pkl:"constructor_name"`

	// Default template for source-to-target mapping function names.
	// `{TargetStructName}` will be replaced with the actual target struct name.
	//
	// Can be overridden per struct.
	SourceToTargetFunctionName string `pkl:"source_to_target_function_name"`

	// Default template for target-to-source mapping function names.
	// `{TargetStructName}` will be replaced with the actual target struct name.
	//
	// Can be overridden per struct.
	SourceFromTargetFunctionName string `pkl:"source_from_target_function_name"`

	// Controls whether and how decorators are generated.
	//
	// - "adaptive": Generate decorators only when customization hooks are needed.
	// - "always": Always generate decorator types.
	// - "never": Never generate decorators.
	DecoratorMode string `pkl:"decorator_mode"`

	// Name of the generated decorator interface.
	//
	// Used only when decorator_mode != "never"
	DecoratorInterfaceName string `pkl:"decorator_interface_name"`

	// Name of the no-op decorator implementation. Empty string
	// would skip the NoOp implementation.
	//
	// Used only when decorator_mode != "never"
	DecoratorNoopName string `pkl:"decorator_noop_name"`

	// Template for the decorator function name.
	//
	// `{FunctionName}` will be replaced with the decorated function name.
	DecorateFunctionName string `pkl:"decorate_function_name"`

	// Default target package for generated mapper code.
	//
	// Defaults to the current package where generation is invoked.
	TargetPkg string `pkl:"target_pkg"`

	// Default source package containing input structs.
	//
	// Can be overridden per struct.
	SourcePkg string `pkl:"source_pkg"`

	// Mapping definitions for this package.
	//
	// Each entry describes how a source struct maps to a target struct.
	Structs map[string]Struct `pkl:"structs"`

	// Whether to prefer getter methods over direct field access
	// on source structs by default.
	//
	// Can be overridden per struct.
	UseGetterIfAvailable bool `pkl:"use_getter_if_available"`

	// Whether to generate source-to-target mapping code.
	// When false, only target-to-source mapping is generated.
	//
	// Can be overridden per struct.
	GenerateSourceToTarget bool `pkl:"generate_source_to_target"`

	// Whether to generate target-to-source mapping code.
	// When false, only source-to-target mapping is generated.
	//
	// Can be overridden per struct.
	GenerateSourceFromTarget bool `pkl:"generate_source_from_target"`

	// Whether to generate GoDoc comments for generated code.
	GenerateGoDoc bool `pkl:"generate_go_doc"`
}

// Controls the overall generation strategy.
//
// - "types": Generates an interface, implementation struct, and constructor.
// - "functions": Generates package-level mapping functions only (no interface or struct).
//
// Note: When mode = "functions", type-related naming options are ignored.
func (rcv PackageImpl) GetMode() string {
	return rcv.Mode
}

// Name of the generated mapper interface.
//
// Used only when mode = "types".
// Defaults to a non-exported name to avoid polluting consumer APIs.
func (rcv PackageImpl) GetInterfaceName() string {
	return rcv.InterfaceName
}

// Name of the generated mapper implementation struct.
//
// Used only when mode = "types".
func (rcv PackageImpl) GetImplementationName() string {
	return rcv.ImplementationName
}

// Name of the constructor function for the mapper implementation.
//
// Used only when mode = "types".
func (rcv PackageImpl) GetConstructorName() string {
	return rcv.ConstructorName
}

// Default template for source-to-target mapping function names.
// `{TargetStructName}` will be replaced with the actual target struct name.
//
// Can be overridden per struct.
func (rcv PackageImpl) GetSourceToTargetFunctionName() string {
	return rcv.SourceToTargetFunctionName
}

// Default template for target-to-source mapping function names.
// `{TargetStructName}` will be replaced with the actual target struct name.
//
// Can be overridden per struct.
func (rcv PackageImpl) GetSourceFromTargetFunctionName() string {
	return rcv.SourceFromTargetFunctionName
}

// Controls whether and how decorators are generated.
//
// - "adaptive": Generate decorators only when customization hooks are needed.
// - "always": Always generate decorator types.
// - "never": Never generate decorators.
func (rcv PackageImpl) GetDecoratorMode() string {
	return rcv.DecoratorMode
}

// Name of the generated decorator interface.
//
// Used only when decorator_mode != "never"
func (rcv PackageImpl) GetDecoratorInterfaceName() string {
	return rcv.DecoratorInterfaceName
}

// Name of the no-op decorator implementation. Empty string
// would skip the NoOp implementation.
//
// Used only when decorator_mode != "never"
func (rcv PackageImpl) GetDecoratorNoopName() string {
	return rcv.DecoratorNoopName
}

// Template for the decorator function name.
//
// `{FunctionName}` will be replaced with the decorated function name.
func (rcv PackageImpl) GetDecorateFunctionName() string {
	return rcv.DecorateFunctionName
}

// Default target package for generated mapper code.
//
// Defaults to the current package where generation is invoked.
func (rcv PackageImpl) GetTargetPkg() string {
	return rcv.TargetPkg
}

// Default source package containing input structs.
//
// Can be overridden per struct.
func (rcv PackageImpl) GetSourcePkg() string {
	return rcv.SourcePkg
}

// Mapping definitions for this package.
//
// Each entry describes how a source struct maps to a target struct.
func (rcv PackageImpl) GetStructs() map[string]Struct {
	return rcv.Structs
}

// Whether to prefer getter methods over direct field access
// on source structs by default.
//
// Can be overridden per struct.
func (rcv PackageImpl) GetUseGetterIfAvailable() bool {
	return rcv.UseGetterIfAvailable
}

// Whether to generate source-to-target mapping code.
// When false, only target-to-source mapping is generated.
//
// Can be overridden per struct.
func (rcv PackageImpl) GetGenerateSourceToTarget() bool {
	return rcv.GenerateSourceToTarget
}

// Whether to generate target-to-source mapping code.
// When false, only source-to-target mapping is generated.
//
// Can be overridden per struct.
func (rcv PackageImpl) GetGenerateSourceFromTarget() bool {
	return rcv.GenerateSourceFromTarget
}

// Whether to generate GoDoc comments for generated code.
func (rcv PackageImpl) GetGenerateGoDoc() bool {
	return rcv.GenerateGoDoc
}
