// Code generated from Pkl module `gomappergen.Config`. DO NOT EDIT.
package pkl

import "github.com/toniphan21/go-mapper-gen/pkg/pkl/mapper"

type BasePackage interface {
	mapper.Package

	GetOutput() *Output
}

var _ BasePackage = BasePackageImpl{}

type BasePackageImpl struct {
	mapper.PackageImpl

	Output *Output `pkl:"output"`
}

func (rcv BasePackageImpl) GetOutput() *Output {
	return rcv.Output
}
