package file

type File interface {
	FilePath() string

	FileContent() []byte
}

func New(path string, content []byte) File {
	return &fileImpl{
		path:    path,
		content: content,
	}
}

type fileImpl struct {
	path    string
	content []byte
}

func (f *fileImpl) FilePath() string {
	return f.path
}

func (f *fileImpl) FileContent() []byte {
	return f.content
}

var _ File = (*fileImpl)(nil)
