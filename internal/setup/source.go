package setup

import (
	"fmt"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/toniphan21/go-mapper-gen/internal/setup/file"
	"golang.org/x/tools/go/packages"
)

func writeSourceFile(t *testing.T, out string, f file.File) {
	path := filepath.Join(out, f.FilePath())
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("cannot mkdir: %s", err)
	}

	if err := os.WriteFile(path, f.FileContent(), 0644); err != nil {
		t.Fatalf("cannot write file %s: %s", path, err)
	}

}

func SourceCode(t *testing.T, files []file.File, additionalFiles ...file.File) string {
	tmp := t.TempDir()

	for _, f := range files {
		writeSourceFile(t, tmp, f)
	}
	for _, f := range additionalFiles {
		writeSourceFile(t, tmp, f)
	}

	return tmp
}

func LoadDir(dir string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		Fset: token.NewFileSet(),
		Dir:  dir,
	}

	return packages.Load(cfg, "./...")
}

func LoadPkg(pattern string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
	}

	return packages.Load(cfg, pattern)
}

func FindPackageByPkgPath(dir string, pkgPath string) (*packages.Package, error) {
	pkgs, err := LoadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		if pkg.PkgPath == pkgPath {
			return pkg, nil
		}
	}
	return nil, fmt.Errorf("package not found: %s", pkgPath)
}

func FindPackageByName(dir string, name string) (*packages.Package, error) {
	pkgs, err := LoadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		if pkg.Name == name {
			return pkg, nil
		}
	}
	return nil, fmt.Errorf("package not found: %s", name)
}

func RunGoModTidy(t *testing.T, dir string) {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run go mod tidy: %v", err)
	}
}
