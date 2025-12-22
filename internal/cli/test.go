package cli

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	gen "github.com/toniphan21/go-mapper-gen"
	"github.com/toniphan21/go-mapper-gen/internal/setup"
	"github.com/toniphan21/go-mapper-gen/internal/setup/file"
	"github.com/toniphan21/go-mapper-gen/internal/util"
)

func runTest(cmd TestCmd, logger *slog.Logger) error {
	gen.UseTestVersion()
	util.SetTabSize(cmd.TabSize)

	var tempDirs []string
	defer func() {
		if len(tempDirs) > 0 {
			logger.Info("deleting temporary directories")
		}

		for _, dir := range tempDirs {
			logger.Debug("\tdeleted temporary directory " + dir)
			_ = os.RemoveAll(dir)
		}
	}()

	for _, inputFile := range cmd.Files {
		logger.Info("running test for " + util.ColorBlue(inputFile))

		stat, err := os.Stat(inputFile)
		if err != nil {
			logger.Error(util.ColorRed(err.Error()))
			continue
		}

		if stat.IsDir() {
			logger.Warn(util.ColorBlue(inputFile) + " is a directory, skipped")
			continue
		}

		content, err := os.ReadFile(inputFile)
		if err != nil {
			logger.Error(util.ColorRed(err.Error()))
			continue
		}

		mdTestCases := gen.Test.ParseMarkdownTestCases(content)
		for _, mdTestCase := range mdTestCases {
			logger.Info("\ttesting " + util.ColorCyan(mdTestCase.Name))
			tempDir, err := os.MkdirTemp("", "mapper-*")
			if err != nil {
				logger.Error("\tError creating temp dir:", slog.Any("error", err))
			}
			logger.Info("\t\tcreated temporary directory at " + util.ColorWhite(tempDir))

			if mdTestCase.GoModFileContent != nil {
				logger.Info("\t\tcopied go.mod file")
				if err := writeTestFile(tempDir, "go.mod", mdTestCase.GoModFileContent); err != nil {
					logger.Error("\t\t" + util.ColorRed(err.Error()))
					continue
				}
				util.PrintFileWithFunction("", mdTestCase.GoModFileContent, func(s string) {
					logger.Debug("\t\t" + s)
				})

				directDependencies, err := gen.Test.ParseDirectDependencies(mdTestCase.GoModFileContent)
				if err != nil {
					logger.Error("\t\t" + util.ColorRed(err.Error()))
					continue
				}

				if len(directDependencies) > 0 {
					if err = setup.ExecuteGoGet(tempDir, directDependencies); err != nil {
						logger.Error("\t\t" + util.ColorRed(err.Error()))
						continue
					}
					logger.Info("\t\tinstalled dependencies")
				}
			}

			if mdTestCase.GoSumFileContent != nil {
				logger.Info("\t\tcopied go.sum file")
				if err := writeTestFile(tempDir, "go.sum", mdTestCase.GoSumFileContent); err != nil {
					logger.Error(util.ColorRed(err.Error()))
					continue
				}
				util.PrintFileWithFunction("", mdTestCase.GoSumFileContent, func(s string) {
					logger.Debug("\t\t" + s)
				})
			}

			setupOk := true

			if mdTestCase.PklDevFileContent != nil {
				pklLibFiles := setup.PklLibFiles()
				for _, v := range pklLibFiles {
					fp := path.Join("_pkl_lib_", v.FilePath())
					logger.Info("\t\tcopied pkl lib file " + v.FilePath() + " to " + fp)

					if err = writeTestFile(tempDir, fp, v.FileContent()); err != nil {
						logger.Error("\t\t" + util.ColorRed(err.Error()))
						setupOk = false
						continue
					}
				}

				configFile := file.MakePklDevConfigFile(
					"./_pkl_lib_/Config.pkl",
					"mapper.pkl",
					[]string{string(mdTestCase.PklDevFileContent)},
				)

				if err = writeTestFile(tempDir, configFile.FilePath(), configFile.FileContent()); err != nil {
					logger.Error("\t\t" + util.ColorRed(err.Error()))
					setupOk = false
				}
				logger.Info("\t\tmade pkl config file " + util.ColorWhite(configFile.FilePath()))
				util.PrintFileWithFunction("", configFile.FileContent(), func(s string) {
					logger.Debug("\t\t" + s)
				})
			}

			if !setupOk {
				continue
			}

			for fn, fc := range mdTestCase.SourceFiles {
				if err = writeTestFile(tempDir, fn, fc); err != nil {
					logger.Error(util.ColorRed(err.Error()))
					setupOk = false
					continue
				}
				logger.Info("\t\tcopied source file " + util.ColorWhite(fn))

				util.PrintFileWithFunction("", fc, func(s string) {
					logger.Debug("\t\t" + s)
				})

				if fn == "go.mod" {
					directDependencies, err := gen.Test.ParseDirectDependencies(fc)
					if err != nil {
						logger.Error("\t\t" + util.ColorRed(err.Error()))
						setupOk = false
						continue
					}

					if len(directDependencies) > 0 {
						if err = setup.ExecuteGoGet(tempDir, directDependencies); err != nil {
							logger.Error("\t\t" + util.ColorRed(err.Error()))
							setupOk = false
							continue
						}
						logger.Info("\t\tinstalled dependencies")
					}
				}
			}

			if !setupOk {
				continue
			}

			logger.Info("\t\tgenerating mapper")
			glogger := gen.NewNoopLogger()
			if cmd.LogGenerate {
				glogger = logger
			}
			err = runGenerate(GenerateCmd{WorkingDir: tempDir}, glogger)
			if err != nil {
				logger.Error(util.ColorRed(err.Error()))
				continue
			}

			isSuccess := true
			for fn, fc := range mdTestCase.GoldenFiles {
				out, err := readTestFile(tempDir, fn)
				if err != nil {
					if errors.Is(err, os.ErrNotExist) {
						logger.Error(util.ColorRed(fmt.Sprintf("\t\texpected golden file %v but file does not exist", fn)))
						isSuccess = false
						continue
					}

					logger.Error(err.Error())
					continue
				}

				if !compareFileContent(out, fc) {
					logger.Error(util.ColorRed("\t\tgolden file content does not match expectation"))
					util.PrintDiffWithFunction("expected", fc, "generated", out, func(s string) {
						logger.Error("\t\t" + s)
					})
					isSuccess = false

					_ = writeTestFile(tempDir, fn+".golden", fc)
					continue
				}

				logger.Debug("\t\tpassed with " + util.ColorYellow("golden file ") + util.ColorWhite(fn))
				util.PrintFileWithFunction("", fc, func(s string) {
					logger.Debug("\t\t" + s)
				})
			}

			if isSuccess {
				tempDirs = append(tempDirs, tempDir)
				logger.Info("\t\tput temporary directory to clean up list")
				if cmd.EmitCode {
					if err := emitTestCode(mdTestCase, tempDir, logger); err != nil {
						logger.Error("\t\t" + util.ColorRed(err.Error()))
					}
				}
				logger.Info(util.ColorGreen("\t\t\u2714 passed"))
				logger.Info("")
				continue
			}

			logger.Info("\t\tsaved test case to file test.md")
			logger.Info("\t\ttemporary directory will not be deleted")
			_ = writeTestFile(tempDir, "test.md", []byte(mdTestCase.Content))

			logger.Info(util.ColorRed("\t\t\u2718 failed"))
			logger.Info("")
		}

	}
	return nil
}

func emitTestCode(testCase gen.MarkdownTestCase, tempDir string, logger *slog.Logger) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	var readme []string
	var examplePath string

	emit := false
	lines := strings.Split(testCase.Content, "\n")
	for _, v := range lines {
		line := strings.TrimSpace(v)
		if strings.HasPrefix(line, "[//]: # (EmitCode:") && strings.HasSuffix(line, ")") {
			line = strings.TrimPrefix(line, "[//]: # (EmitCode:")
			line = strings.TrimSuffix(line, ")")
			examplePath = line
			emit = true
			continue
		}
		readme = append(readme, v)
	}

	if emit {
		// copy all files in temp to
		dst := filepath.Join(wd, examplePath)
		if err = os.RemoveAll(dst); err != nil {
			return err
		}

		if err = copyDir(tempDir, dst); err != nil {
			return err
		}

		_ = writeTestFile(dst, "README.md", []byte(strings.Join(readme, "\n")))
		logger.Info("\t\temit code to " + examplePath)
	}
	return nil
}

func writeTestFile(testDir string, filePath string, fileContent []byte) error {
	fp := filepath.Join(testDir, filePath)
	dir := filepath.Dir(fp)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	return os.WriteFile(fp, fileContent, 0600)
}

func readTestFile(testDir string, filePath string) ([]byte, error) {
	fp := filepath.Join(testDir, filePath)
	return os.ReadFile(fp)
}

func compareFileContent(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for i := 0; i < len(left); i++ {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func copyDir(src string, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		return copyFile(path, targetPath)
	})
}

func copyFile(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer out.Close()

	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer in.Close()

	_, err = io.Copy(out, in)
	return err
}
