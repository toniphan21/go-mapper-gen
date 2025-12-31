// Code generated from Pkl module `gomappergen.Config`. DO NOT EDIT.
package pkl

type Package interface {
	BasePackage

	GetPriorities() map[int]Package
}

var _ Package = PackageImpl{}

type PackageImpl struct {
	BasePackageImpl

	Priorities map[int]Package `pkl:"priorities"`
}

func (rcv PackageImpl) GetPriorities() map[int]Package {
	return rcv.Priorities
}
