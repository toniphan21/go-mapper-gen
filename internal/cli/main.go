package cli

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	gomappergen "github.com/toniphan21/go-mapper-gen"
	"github.com/toniphan21/go-mapper-gen/internal/util"
)

const defaultConfigFileName = "mapper.pkl"

type VersionCmd struct{}

type GenerateCmd struct {
	WorkingDir string `arg:"-w,--working-dir" help:"Base directory" default:"." placeholder:"DIR"`

	ConfigFileName string `arg:"-c,--config" help:"Config file name" default:"mapper.pkl" placeholder:"NAME"`
}

type TestCmd struct {
	Files       []string `arg:"positional" help:"Markdown files for BDD tests" placeholder:"FILE"`
	Name        string   `arg:"-n,--name" help:"Run test which has matched name (case insensitive)" default:""`
	TabSize     int      `arg:"-t,--tab-size" help:"Number of spaces to use in tab size" default:"8"`
	LogGenerate bool     `arg:"-l,--log-generate" help:"Log in generate code process" default:"false"`
	EmitCode    bool     `arg:"-e,--emit-code" help:"Emit to code if the test passed" default:"false"`
	ShowSetup   bool     `arg:"-s,--show-setup" help:"Show test setup step, if you are running less than 10 test cases it will be show by default"`
}

type Args struct {
	Test     *TestCmd     `arg:"subcommand:test" help:"Run BDD tests using markdown files"`
	Generate *GenerateCmd `arg:"subcommand:generate" help:"Generate code from configuration"`
	Version  *VersionCmd  `arg:"subcommand:version" help:"Print version information and exit"`

	NoColor bool `arg:"--no-color" help:"Disable colors" default:"false"`
	Verbose bool `arg:"-v,--verbose" help:"Enable verbose logging"`
}

func Run(args Args) {
	execPath := os.Args[0]

	level := "info"
	if args.Verbose {
		level = "debug"
	}

	handler := util.NewSlogHandler(os.Stdout, level)
	logger := slog.New(handler)

	handleError := func(err error) {
		if err != nil {
			logger.Error(util.ColorRed(err.Error()))
			return
		}
		logger.Debug(fmt.Sprintf("Total LookUp hits %d", gomappergen.LookUpTotalHits))
		logger.Debug("")
		logger.Info(util.ColorGreen("done"))
	}

	switch {
	case args.Version != nil:
		fmt.Println(gomappergen.Version())

	case args.Test != nil:
		var inputs []string

		for _, pattern := range args.Test.Files {
			matches, err := filepath.Glob(pattern)

			if err != nil || len(matches) == 0 {
				inputs = append(inputs, pattern)
			} else {
				inputs = append(inputs, matches...)
			}
		}
		handleError(runTest(*args.Test, execPath, logger))

	case args.Generate != nil:
		absPath, err := filepath.Abs(args.Generate.WorkingDir)
		if err != nil {
			panic(err)
			return
		}

		handleError(runGenerate(GenerateCmd{
			WorkingDir:     absPath,
			ConfigFileName: defaultConfigFileName,
		}, logger))

	default:
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		handleError(runGenerate(GenerateCmd{
			WorkingDir:     wd,
			ConfigFileName: defaultConfigFileName,
		}, logger))
	}
}
