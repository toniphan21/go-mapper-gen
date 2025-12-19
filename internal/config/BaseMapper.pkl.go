// Code generated from Pkl module `gomappergen.Config`. DO NOT EDIT.
package config

type BaseMapper interface {
	Base

	GetInterfaceName() string

	GetImplementationName() string

	GetConstructorName() string

	GetSourceToTargetFunctionName() string

	GetSourceFromTargetFunctionName() string

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

type BaseMapperImpl struct {
	BaseImpl

	InterfaceName string `pkl:"interface_name"`

	ImplementationName string `pkl:"implementation_name"`

	ConstructorName string `pkl:"constructor_name"`

	SourceToTargetFunctionName string `pkl:"source_to_target_function_name"`

	SourceFromTargetFunctionName string `pkl:"source_from_target_function_name"`

	DecoratorInterfaceName string `pkl:"decorator_interface_name"`

	DecoratorNoopName string `pkl:"decorator_noop_name"`

	DecorateFunctionName string `pkl:"decorate_function_name"`

	TargetPkg string `pkl:"target_pkg"`

	SourcePkg string `pkl:"source_pkg"`

	Structs map[string]MapperStruct `pkl:"structs"`

	UseGetterIfAvailable bool `pkl:"use_getter_if_available"`

	GenerateGoDoc bool `pkl:"generate_go_doc"`
}

func (rcv BaseMapperImpl) GetInterfaceName() string {
	return rcv.InterfaceName
}

func (rcv BaseMapperImpl) GetImplementationName() string {
	return rcv.ImplementationName
}

func (rcv BaseMapperImpl) GetConstructorName() string {
	return rcv.ConstructorName
}

func (rcv BaseMapperImpl) GetSourceToTargetFunctionName() string {
	return rcv.SourceToTargetFunctionName
}

func (rcv BaseMapperImpl) GetSourceFromTargetFunctionName() string {
	return rcv.SourceFromTargetFunctionName
}

func (rcv BaseMapperImpl) GetDecoratorInterfaceName() string {
	return rcv.DecoratorInterfaceName
}

func (rcv BaseMapperImpl) GetDecoratorNoopName() string {
	return rcv.DecoratorNoopName
}

func (rcv BaseMapperImpl) GetDecorateFunctionName() string {
	return rcv.DecorateFunctionName
}

func (rcv BaseMapperImpl) GetTargetPkg() string {
	return rcv.TargetPkg
}

func (rcv BaseMapperImpl) GetSourcePkg() string {
	return rcv.SourcePkg
}

func (rcv BaseMapperImpl) GetStructs() map[string]MapperStruct {
	return rcv.Structs
}

func (rcv BaseMapperImpl) GetUseGetterIfAvailable() bool {
	return rcv.UseGetterIfAvailable
}

func (rcv BaseMapperImpl) GetGenerateGoDoc() bool {
	return rcv.GenerateGoDoc
}
