
package main

import (
	"go/types"
	"log/slog"
	"os"

	"github.com/dave/jennifer/jen"
	gen "github.com/toniphan21/go-mapper-gen"
	gencli "github.com/toniphan21/go-mapper-gen/pkg/cli"
)

func main() {
	// you need to handle args and create a slog.Logger instance, otherwise go-mapper-gen will use default one.
	// ...
	// ...

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	err = gencli.Run(gencli.RunCommand{
		// required, working directory
		WorkingDir: wd,

		// optional, config file name, default: mapper.pkl
		// ConfigFileName: "config.pkl",

		// optional, default: false
		// DryRun: false,

		// optional, default false
		PrintRegisteredConverters: true,

		// optional, if nil use DefaultParser
		// Parser: nil,

		// optional, if nil use DefaultFileManager
		// FileManager: nil,

		// optional, if nil use slog.DefaultLogger
		// Logger: nil

		// optional
		RegisterConverters: func() {
			// you can register your Converters here
			gen.RegisterConverter(&awesomeConverter{})
			slog.Info("registered awesome converter")
		},
	})

	if err != nil {
		// handle error
	}
}

type awesomeConverter struct{}

func (c *awesomeConverter) Init(parser gen.Parser, config gen.Config) {
	// no-op
}

func (c *awesomeConverter) Info() gen.ConverterInfo {
	return gen.ConverterInfo{
		Name:                 "MyAwesomeConverter",
		ShortForm:            "awesome -> legendary",
		ShortFormDescription: "convert awesome to legen... wait for it... dary!",
	}
}

func (c *awesomeConverter) CanConvert(ctx gen.LookupContext, targetType, sourceType types.Type) bool {
	// perform the check that it can convert sourceType to targetType
	return false
}

func (c *awesomeConverter) ConvertField(ctx gen.ConverterContext, target, source gen.Symbol, opts gen.ConverterOption) jen.Code {
	// generate your code here
	return nil
}

var _ gen.Converter = (*awesomeConverter)(nil)
