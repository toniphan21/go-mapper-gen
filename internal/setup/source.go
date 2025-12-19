package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/toniphan21/go-mapper-gen/internal/setup/file"
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

func RunGoModTidy(t *testing.T, dir string) {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run go mod tidy: %v", err)
	}
}

func RunGoGet(t *testing.T, dir string, requires map[string]string) {
	err := ExecuteGoGet(dir, requires)
	if err != nil {
		t.Fatalf("Failed to run go mod tidy: %v", err)
	}
}

func ExecuteGoGet(dir string, requires map[string]string) error {
	for pkg, version := range requires {
		cmd := exec.Command("go", "get", fmt.Sprintf("%v@%v", pkg, version))
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
