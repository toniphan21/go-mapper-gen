package gomappergen

import (
	"github.com/dave/jennifer/jen"
	"golang.org/x/tools/go/packages"
)

type FileManager interface {
	MakeJenFile(currentPkg *packages.Package, config Config) *jen.File

	JenFiles() map[string]*jen.File
}

func DefaultFileManager() FileManager {
	return &defaultFileManager{files: make(map[string]*jen.File)}
}

type defaultFileManager struct {
	files map[string]*jen.File
}

func (f *defaultFileManager) JenFiles() map[string]*jen.File {
	return f.files
}

func (f *defaultFileManager) MakeJenFile(currentPkg *packages.Package, config Config) *jen.File {
	jf, have := f.files[config.Output.FileName]
	if have {
		return jf
	}

	pkgName := replacePlaceholders(config.Output.PkgName, map[string]string{
		Placeholder.CurrentPackageName: currentPkg.Name,
	})

	f.files[config.Output.FileName] = jen.NewFilePathName(currentPkg.PkgPath, pkgName)
	return f.files[config.Output.FileName]
}

var _ FileManager = (*defaultFileManager)(nil)
