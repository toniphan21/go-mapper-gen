package cli

import (
	"log/slog"

	lib "github.com/toniphan21/go-mapper-gen/pkg/cli"
)

func runGenerate(cmd GenerateCmd, logger *slog.Logger) error {
	return lib.Run(lib.RunCommand{
		WorkingDir:                cmd.WorkingDir,
		ConfigFileName:            cmd.ConfigFileName,
		DryRun:                    cmd.DryRun,
		PrintRegisteredConverters: true,
		Logger:                    logger,
	})
}
