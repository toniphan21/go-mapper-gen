package app

import (
	"log/slog"
	"os"
	"path/filepath"

	gen "github.com/toniphan21/go-mapper-gen"
	"github.com/toniphan21/go-mapper-gen/internal/util"
)

const appName = "go-mapper-gen"
const binary = "github.com/toniphan21/go-mapper-gen"
const configFileName = "config.pkl"

var mode = "dev"
var version = "v1.0.0"

func RunCLI() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	handler := util.NewSlogHandler(os.Stdout, "info")
	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info(util.ColorGreen(appName) + " " + version)
	slog.Info(util.ColorGreen(appName) + " is working on directory: " + util.ColorCyan(wd))

	cff := filepath.Join(wd, configFileName)
	slog.Info(util.ColorGreen(appName) + " uses configuration file: " + util.ColorCyan(cff))

	cf, err := gen.ParseConfig(cff)
	//if err != nil {
	//	slog.Error(util.ColorRed("failed to load configuration file."))
	//	slog.Error(util.ColorRed(err.Error()))
	//	os.Exit(1)
	//}

	//fmt.Println(cf)
	gen.RegisterAllBuiltinConverters()

	logger.Info(util.ColorGreen(appName) + " is running with registered field converters:")
	gen.PrintRegisteredConverters(logger)

	for _, config := range cf {
		_ = gen.Generate(nil, nil, config)
	}
}
