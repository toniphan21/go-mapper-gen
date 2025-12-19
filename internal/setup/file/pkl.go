package file

import (
	"fmt"
	"strings"
)

type Pkl struct {
	Path    string
	Content string
	Lines   []string
}

func (p *Pkl) FilePath() string {
	return p.Path
}

func (p *Pkl) FileContent() []byte {
	if p.Lines == nil {
		return []byte(p.Content)
	}
	return []byte(strings.Join(p.Lines, "\n"))
}

var _ File = (*Pkl)(nil)

func PklDevConfigFile(lines ...string) File {
	return MakePklDevConfigFile(".../Config.pkl", "dev/config.pkl", lines)
}

func MakePklDevConfigFile(amends string, filePath string, lines []string) File {
	var l = []string{
		fmt.Sprintf(`amends "%s"`, amends),
		``,
	}
	l = append(l, lines...)

	return &Pkl{
		Path:  filePath,
		Lines: l,
	}
}
