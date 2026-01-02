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
	return MakePklDevConfigFile(".../Config.pkl", []string{`.../set.pkl`}, "dev/config.pkl", lines)
}

func MakePklDevConfigFile(amends string, imports []string, filePath string, lines []string) File {
	var l = []string{
		fmt.Sprintf(`amends "%s"`, amends),
		``,
	}
	if len(imports) > 0 {
		for _, i := range imports {
			l = append(l, fmt.Sprintf(`import "%s"`, i))
		}
		l = append(l, "")
	}
	l = append(l, lines...)

	return &Pkl{
		Path:  filePath,
		Lines: l,
	}
}
