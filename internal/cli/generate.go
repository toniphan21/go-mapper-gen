package cli

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	gen "github.com/toniphan21/go-mapper-gen"
	"github.com/toniphan21/go-mapper-gen/converters/grpc"
	"github.com/toniphan21/go-mapper-gen/converters/pgtype"
	"github.com/toniphan21/go-mapper-gen/converters/sql"
	"github.com/toniphan21/go-mapper-gen/internal/util"
)

const appName = "go-mapper-gen"

func runGenerate(cmd GenerateCmd, logger *slog.Logger) error {
	if cmd.DryRun {
		logger.Info(util.ColorGreen(appName) + " " + gen.Version() + " in DRY mode")
	} else {
		logger.Info(util.ColorGreen(appName) + " " + gen.Version())
	}
	logger.Info(util.ColorGreen(appName) + " is working on directory: " + util.ColorCyan(cmd.WorkingDir))

	cff := filepath.Join(cmd.WorkingDir, defaultConfigFileName)
	logger.Info(util.ColorGreen(appName) + " uses configuration file: " + util.ColorCyan(cff))

	parsedConfig, err := gen.ParseConfig(cff)
	if err != nil {
		logger.Error(util.ColorRed("failed to load configuration file."))
		return err
	}

	fm := gen.DefaultFileManager()
	parser, err := gen.DefaultParser(cmd.WorkingDir)
	if err != nil {
		logger.Error(util.ColorRed("failed to parse source code."))
		return err
	}

	gen.ClearAllRegisteredConverters()
	gen.RegisterBuiltinConverters(parsedConfig.BuiltInConverters)
	loadLibraryConverters(parsedConfig.LibraryConverters)

	generator := gen.New(parser, *parsedConfig, gen.WithLogger(logger), gen.WithFileManager(fm))

	logger.Info(util.ColorGreen(appName) + " is running with registered field converters:")
	gen.PrintRegisteredConverters(logger)

	logger.Info(util.ColorGreen(appName) + " initiated successfully")

	for _, pkg := range parser.SourcePackages() {
		pkgPath := pkg.PkgPath
		configs, have := parsedConfig.Packages[pkgPath]
		if !have {
			logger.Debug(fmt.Sprintf("package %s has no config, skipped", util.ColorCyan(pkgPath)))
			continue
		}

		logger.Info(util.ColorGreen(appName)+" is generating for package %s", util.ColorCyan(pkgPath))
		err = generator.Generate(pkg, configs)
		if err != nil {
			logger.Error(fmt.Sprintf("cannot generate for package %s: %s", util.ColorCyan(pkgPath), util.ColorRed(err.Error())))
		}
	}

	outs := fm.JenFiles()
	if cmd.DryRun {
		logger.Info(util.ColorGreen(appName) + " is printing generated file content")
		for p, out := range outs {
			rp := filepath.Join(cmd.WorkingDir, p)
			logger.Info(util.ColorGreen(appName) + " will generated to file " + util.ColorBlue(rp))
			util.PrintFileWithFunction(p, []byte(out.GoString()), func(l string) {
				logger.Info(l)
			})
		}
	} else {
		logger.Info(util.ColorGreen(appName) + " is saving generated file to disk")
		for p, out := range outs {
			rp := filepath.Join(cmd.WorkingDir, p)
			_ = os.WriteFile(rp, []byte(out.GoString()), 0644)
			logger.Info(util.ColorGreen(appName) + " saved to file " + util.ColorBlue(rp))
		}
	}

	logger.Info("")
	return nil
}

func loadLibraryConverters(cf gen.LibraryConverterConfig) {
	if cf.UseGRPC {
		grpc.RegisterConverters()
	}
	if cf.UsePGType {
		pgtype.RegisterConverters()
	}
	if cf.UseSQL {
		sql.RegisterConverters()
	}
}
