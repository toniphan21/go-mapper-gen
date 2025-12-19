package gomappergen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/toniphan21/go-mapper-gen/internal/setup"
	"github.com/toniphan21/go-mapper-gen/internal/setup/file"
	"golang.org/x/mod/modfile"
)

type testHelper struct {
	*goldenTest
	*converterTest
	*bddTest
}

func (h *testHelper) Parse(t *testing.T, files []file.File) (Parser, error) {
	dir := setup.SourceCode(t, files)
	for _, f := range files {
		if f.FilePath() == "go.mod" {
			directDependencies, err := h.ParseDirectDependencies(f.FileContent())
			if err != nil {
				return nil, err
			}

			if len(directDependencies) > 0 {
				setup.RunGoGet(t, dir, directDependencies)
			}
		}
	}
	return DefaultParser(dir)
}

func (h *testHelper) ParseDirectDependencies(goModContent []byte) (map[string]string, error) {
	f, err := modfile.Parse("go.mod", goModContent, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse go.mod: %w", err)
	}

	dependencies := make(map[string]string)

	for _, req := range f.Require {
		if !req.Indirect {
			dependencies[req.Mod.Path] = req.Mod.Version
		}
	}
	return dependencies, nil
}

func (h *testHelper) ContentFromTestData(elem ...string) []byte {
	elems := []string{"testdata"}
	elems = append(elems, elem...)
	c, err := os.ReadFile(filepath.Join(elems...))
	if err != nil {
		panic(err)
	}
	return c
}

func (h *testHelper) FileLines(lines ...string) []byte {
	return []byte(strings.Join(lines, "\n"))
}

func (h *testHelper) MakeGoModFileContent(module string, version *string, requires map[string]string) []byte {
	v := ""
	if version != nil {
		v = *version
	}
	goMod := &file.GoMod{
		Module:   module,
		Version:  v,
		Requires: requires,
	}
	return goMod.FileContent()
}

var Test = &testHelper{
	goldenTest:    &goldenTest{},
	converterTest: &converterTest{},
	bddTest:       &bddTest{},
}
