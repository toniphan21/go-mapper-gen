package app

import (
	"log/slog"
	"os"
	"path/filepath"

	gen "github.com/toniphan21/go-mapper-gen"
	"github.com/toniphan21/go-mapper-gen/internal/util"
)

const appName = "go-mapper-gen"
const configFileName = "mapper.pkl"

func RunCLI() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	handler := util.NewSlogHandler(os.Stdout, "info")
	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info(util.ColorGreen(appName) + " " + gen.Version())
	slog.Info(util.ColorGreen(appName) + " is working on directory: " + util.ColorCyan(wd))

	cff := filepath.Join(wd, configFileName)
	slog.Info(util.ColorGreen(appName) + " uses configuration file: " + util.ColorCyan(cff))

	parsedConfig, err := gen.ParseConfig(cff)
	if err != nil {
		slog.Error(util.ColorRed("failed to load configuration file."))
		slog.Error(util.ColorRed(err.Error()))
		os.Exit(1)
	}

	fm := gen.DefaultFileManager()
	parser, err := gen.DefaultParser(wd)
	if err != nil {
		slog.Error(util.ColorRed("failed to parse source code."))
	}

	gen.ClearAllRegisteredConverters()
	gen.RegisterBuiltinConverters(parsedConfig.BuiltInConverters)

	logger.Info(util.ColorGreen(appName) + " is running with registered field converters:")
	gen.PrintRegisteredConverters(logger)

	gen.InitAllRegisteredConverters(parser, *parsedConfig)

	for _, pkg := range parser.SourcePackages() {
		pkgPath := pkg.PkgPath
		configs, have := parsedConfig.Packages[pkgPath]
		if !have {
			continue
		}

		_ = gen.Generate(parser, fm, pkg, configs)
	}

	outs := fm.JenFiles()
	for p, out := range outs {
		_ = os.WriteFile(p, []byte(out.GoString()), 0644)
	}
}
