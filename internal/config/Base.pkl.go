// Code generated from Pkl module `gomappergen.Config`. DO NOT EDIT.
package config

type Base interface {
	GetOutput() *Output

	GetConvertFunctions() *[]ConvertFunc
}

var _ Base = BaseImpl{}

type BaseImpl struct {
	Output *Output `pkl:"output"`

	ConvertFunctions *[]ConvertFunc `pkl:"convert_functions"`
}

func (rcv BaseImpl) GetOutput() *Output {
	return rcv.Output
}

func (rcv BaseImpl) GetConvertFunctions() *[]ConvertFunc {
	return rcv.ConvertFunctions
}
