package file

import "strings"

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
	var l = []string{
		`amends ".../Config.pkl"`,
		``,
	}
	l = append(l, lines...)

	return &Pkl{
		Path:  "dev/config.pkl",
		Lines: l,
	}
}
