// Code generated from Pkl module `gomappergen.Config`. DO NOT EDIT.
package config

type Base interface {
	GetOutput() *Output
}

var _ Base = BaseImpl{}

type BaseImpl struct {
	Output *Output `pkl:"output"`
}

func (rcv BaseImpl) GetOutput() *Output {
	return rcv.Output
}
