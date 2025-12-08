// Code generated from Pkl module `gomappergen.Config`. DO NOT EDIT.
package gen

type BaseMapper interface {
	Base

	GetInterfaceName() string

	GetImplementationName() string

	GetConstructorName() string

	GetSourceToTargetFunctionName() string

	GetSourceFromTargetFunctionName() string

	GetDecoratorFunctionName() string

	GetTargetPkg() *string

	GetSourcePkg() string

	GetStructs() map[string]MapperStruct

	GetGenerateGoDoc() bool

	GetIgnored() bool
}

var _ BaseMapper = BaseMapperImpl{}

type BaseMapperImpl struct {
	BaseImpl

	InterfaceName string `pkl:"interface_name"`

	ImplementationName string `pkl:"implementation_name"`

	ConstructorName string `pkl:"constructor_name"`

	SourceToTargetFunctionName string `pkl:"source_to_target_function_name"`

	SourceFromTargetFunctionName string `pkl:"source_from_target_function_name"`

	DecoratorFunctionName string `pkl:"decorator_function_name"`

	TargetPkg *string `pkl:"target_pkg"`

	SourcePkg string `pkl:"source_pkg"`

	Structs map[string]MapperStruct `pkl:"structs"`

	GenerateGoDoc bool `pkl:"generate_go_doc"`

	Ignored bool `pkl:"ignored"`
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

func (rcv BaseMapperImpl) GetDecoratorFunctionName() string {
	return rcv.DecoratorFunctionName
}

func (rcv BaseMapperImpl) GetTargetPkg() *string {
	return rcv.TargetPkg
}

func (rcv BaseMapperImpl) GetSourcePkg() string {
	return rcv.SourcePkg
}

func (rcv BaseMapperImpl) GetStructs() map[string]MapperStruct {
	return rcv.Structs
}

func (rcv BaseMapperImpl) GetGenerateGoDoc() bool {
	return rcv.GenerateGoDoc
}

func (rcv BaseMapperImpl) GetIgnored() bool {
	return rcv.Ignored
}
