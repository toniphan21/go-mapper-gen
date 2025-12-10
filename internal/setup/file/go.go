package file

import (
	"fmt"
	"strings"
)

const DefaultGoModule = "github.com/toniphan21/go-mapper-gen/example"
const DefaultGoVersion = "1.25"

type Go struct {
	Path    string
	Content string
	Lines   []string
}

func (g Go) FilePath() string {
	return g.Path
}

func (g Go) FileContent() []byte {
	if g.Lines == nil {
		return []byte(g.Content)
	}
	return []byte(strings.Join(g.Lines, "\n"))
}

type GoMod struct {
	Module   string
	Version  string
	Requires []string
}

func (g GoMod) GetModule() string {
	if g.Module == "" {
		return DefaultGoModule
	}
	return g.Module
}

func (g GoMod) GetVersion() string {
	if g.Version == "" {
		return DefaultGoVersion
	}
	return g.Version
}

func (g GoMod) FilePath() string {
	return "go.mod"
}

func (g GoMod) FileContent() []byte {
	var lines []string

	lines = append(lines, fmt.Sprintf("module %s\n", g.GetModule()))
	lines = append(lines, fmt.Sprintf("go %s\n", g.GetVersion()))
	for _, v := range g.Requires {
		lines = append(lines, fmt.Sprintf("require %s", v))
	}

	return []byte(strings.Join(lines, "\n"))
}

var _ File = (*Go)(nil)
var _ File = (*GoMod)(nil)
var _ File = (*Pkl)(nil)
