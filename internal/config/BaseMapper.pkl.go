// Code generated from Pkl module `gomappergen.Config`. DO NOT EDIT.
package config

type BaseMapper interface {
	Base

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

	GetStructs() map[string]MapperStruct

	GetUseGetterIfAvailable() bool

	GetGenerateGoDoc() bool
}

var _ BaseMapper = BaseMapperImpl{}

// Base configuration for mapper code generation.
//
// This configuration controls *how* mappers are generated (types vs functions),
// naming conventions of generated symbols, decorator behavior, and package layout.
type BaseMapperImpl struct {
	BaseImpl

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
	Structs map[string]MapperStruct `pkl:"structs"`

	// Whether to prefer getter methods over direct field access
	// on source structs by default.
	//
	// Can be overridden per struct.
	UseGetterIfAvailable bool `pkl:"use_getter_if_available"`

	// Whether to generate GoDoc comments for generated code.
	GenerateGoDoc bool `pkl:"generate_go_doc"`
}

// Controls the overall generation strategy.
//
// - "types": Generates an interface, implementation struct, and constructor.
// - "functions": Generates package-level mapping functions only (no interface or struct).
//
// Note: When mode = "functions", type-related naming options are ignored.
func (rcv BaseMapperImpl) GetMode() string {
	return rcv.Mode
}

// Name of the generated mapper interface.
//
// Used only when mode = "types".
// Defaults to a non-exported name to avoid polluting consumer APIs.
func (rcv BaseMapperImpl) GetInterfaceName() string {
	return rcv.InterfaceName
}

// Name of the generated mapper implementation struct.
//
// Used only when mode = "types".
func (rcv BaseMapperImpl) GetImplementationName() string {
	return rcv.ImplementationName
}

// Name of the constructor function for the mapper implementation.
//
// Used only when mode = "types".
func (rcv BaseMapperImpl) GetConstructorName() string {
	return rcv.ConstructorName
}

// Default template for source-to-target mapping function names.
// `{TargetStructName}` will be replaced with the actual target struct name.
//
// Can be overridden per struct.
func (rcv BaseMapperImpl) GetSourceToTargetFunctionName() string {
	return rcv.SourceToTargetFunctionName
}

// Default template for target-to-source mapping function names.
// `{TargetStructName}` will be replaced with the actual target struct name.
//
// Can be overridden per struct.
func (rcv BaseMapperImpl) GetSourceFromTargetFunctionName() string {
	return rcv.SourceFromTargetFunctionName
}

// Controls whether and how decorators are generated.
//
// - "adaptive": Generate decorators only when customization hooks are needed.
// - "always": Always generate decorator types.
// - "never": Never generate decorators.
func (rcv BaseMapperImpl) GetDecoratorMode() string {
	return rcv.DecoratorMode
}

// Name of the generated decorator interface.
//
// Used only when decorator_mode != "never"
func (rcv BaseMapperImpl) GetDecoratorInterfaceName() string {
	return rcv.DecoratorInterfaceName
}

// Name of the no-op decorator implementation. Empty string
// would skip the NoOp implementation.
//
// Used only when decorator_mode != "never"
func (rcv BaseMapperImpl) GetDecoratorNoopName() string {
	return rcv.DecoratorNoopName
}

// Template for the decorator function name.
//
// `{FunctionName}` will be replaced with the decorated function name.
func (rcv BaseMapperImpl) GetDecorateFunctionName() string {
	return rcv.DecorateFunctionName
}

// Default target package for generated mapper code.
//
// Defaults to the current package where generation is invoked.
func (rcv BaseMapperImpl) GetTargetPkg() string {
	return rcv.TargetPkg
}

// Default source package containing input structs.
//
// Can be overridden per struct.
func (rcv BaseMapperImpl) GetSourcePkg() string {
	return rcv.SourcePkg
}

// Mapping definitions for this package.
//
// Each entry describes how a source struct maps to a target struct.
func (rcv BaseMapperImpl) GetStructs() map[string]MapperStruct {
	return rcv.Structs
}

// Whether to prefer getter methods over direct field access
// on source structs by default.
//
// Can be overridden per struct.
func (rcv BaseMapperImpl) GetUseGetterIfAvailable() bool {
	return rcv.UseGetterIfAvailable
}

// Whether to generate GoDoc comments for generated code.
func (rcv BaseMapperImpl) GetGenerateGoDoc() bool {
	return rcv.GenerateGoDoc
}
