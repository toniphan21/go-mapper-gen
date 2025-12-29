## [WIP] Use go-mapper-gen as library

Given that you already have a project, first we need to add `go-mapper-gen` as a dependency in `go.mod`

```go.mod
module github.com/toniphan21/awesome

go 1.25

require github.com/toniphan21/go-mapper-gen v0.2.0
```

The `go.sum` is:

```go.sum
github.com/toniphan21/go-mapper-gen v0.2.0 h1:UM30nxsnnwbJz4IbqRiSz0s9rXBfn+xf9zG/COZL0f0=
github.com/toniphan21/go-mapper-gen v0.2.0/go.mod h1:xq9nD8FltQ5wLE81cEH/aXCukKNwA5//11LzamzLMEU=
```

Then set up a command to use as a generator

```go
// file: cmd/generator/main.go

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

	// get current working directory
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
	// perform the check for sourceType/targetType
	return false
}

func (c *awesomeConverter) ConvertField(ctx gen.ConverterContext, target, source gen.Symbol, opts gen.ConverterOption) jen.Code {
	// generate your code here using github.com/dave/jennifer
	return nil
}

var _ gen.Converter = (*awesomeConverter)(nil)
```
