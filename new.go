package gomappergen

import (
	"log/slog"

	"golang.org/x/tools/go/packages"
)

type Generator interface {
	Generate(currentPkg *packages.Package, configs []PackageConfig) error
}

type Options struct {
	Parser      Parser
	FileManager FileManager
	Logger      *slog.Logger
}

type OptionFunc func(*Options)

func New(parser Parser, config Config, options ...OptionFunc) Generator {
	o := &Options{
		Parser:      parser,
		FileManager: DefaultFileManager(),
		Logger:      NewNoopLogger(),
	}

	for _, fn := range options {
		fn(o)
	}

	initRegisteredConverters(parser, config, o.Logger)
	return &generatorImpl{
		parser:      o.Parser,
		fileManager: o.FileManager,
		logger:      o.Logger,
	}
}

func WithLogger(logger *slog.Logger) OptionFunc {
	return func(o *Options) {
		o.Logger = logger
	}
}

func WithFileManager(fileManager FileManager) OptionFunc {
	return func(o *Options) {
		o.FileManager = fileManager
	}
}
