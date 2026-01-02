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
const defaultConfigFileName = "mapper.pkl"

type RunCommand struct {
	WorkingDir                string
	ConfigFileName            string
	DryRun                    bool
	PrintRegisteredConverters bool
	Parser                    gen.Parser
	FileManager               gen.FileManager
	FieldInterceptorProvider  gen.FieldInterceptorProvider
	Logger                    *slog.Logger
	RegisterConverters        func()
}

func Run(cmd RunCommand) error {
	logger := cmd.Logger
	if logger == nil {
		logger = slog.Default()
	}

	if cmd.DryRun {
		logger.Info(util.ColorGreen(appName) + " " + gen.Version() + " in DRY mode")
	} else {
		logger.Info(util.ColorGreen(appName) + " " + gen.Version())
	}
	logger.Info(util.ColorGreen(appName) + " is working on directory: " + util.ColorCyan(cmd.WorkingDir))

	cfn := cmd.ConfigFileName
	if cfn == "" {
		cfn = defaultConfigFileName
	}
	cff := filepath.Join(cmd.WorkingDir, cfn)
	logger.Info(util.ColorGreen(appName) + " uses configuration file: " + util.ColorCyan(cff))

	fip := cmd.FieldInterceptorProvider
	if fip == nil {
		fip = gen.DefaultFieldInterceptorProvider()
	}

	parsedConfig, err := gen.ParseConfig(cff, fip)
	if err != nil {
		logger.Error(util.ColorRed("failed to load configuration file."))
		return err
	}

	fm := cmd.FileManager
	if fm == nil {
		fm = gen.DefaultFileManager()
	}

	parser := cmd.Parser
	if parser == nil {
		parser, err = gen.DefaultParser(cmd.WorkingDir)
		if err != nil {
			logger.Error(util.ColorRed("failed to parse source code."))
			return err
		}
	}

	gen.ClearAllRegisteredConverters()
	gen.RegisterBuiltinConverters(parsedConfig.BuiltInConverters)
	loadLibraryConverters(parsedConfig.LibraryConverters)
	if cmd.RegisterConverters != nil {
		cmd.RegisterConverters()
	}

	generator := gen.New(parser, *parsedConfig, gen.WithLogger(logger), gen.WithFileManager(fm))

	if cmd.PrintRegisteredConverters {
		logger.Info(util.ColorGreen(appName) + " is running with registered field converters:")
		gen.PrintRegisteredConverters(logger)
	}

	logger.Info(util.ColorGreen(appName) + " initiated successfully")

	for _, pkg := range parser.SourcePackages() {
		pkgPath := pkg.PkgPath
		configs, have := parsedConfig.Packages[pkgPath]
		if !have {
			logger.Debug(fmt.Sprintf("package %s has no config, skipped", util.ColorCyan(pkgPath)))
			continue
		}

		logger.Info(util.ColorGreen(appName) + " is generating for package " + util.ColorCyan(pkgPath))
		err = generator.Generate(pkg, configs)
		if err != nil {
			logger.Error(fmt.Sprintf("cannot generate for package %s: %s", util.ColorCyan(pkgPath), util.ColorRed(err.Error())))
		}
	}

	outs := fm.JenFiles()
	if len(outs) == 0 {
		logger.Info(util.ColorGreen(appName) + " generated nothing")
	} else {
		if cmd.DryRun {
			logger.Info(util.ColorGreen(appName) + " is printing generated file content")
			for p, out := range outs {
				rp := filepath.Join(cmd.WorkingDir, p)
				logger.Info(util.ColorGreen(appName) + " will save to file " + util.ColorBlue(rp))
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
