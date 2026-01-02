package cli

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	gen "github.com/toniphan21/go-mapper-gen"
	"github.com/toniphan21/go-mapper-gen/internal/setup"
	"github.com/toniphan21/go-mapper-gen/internal/setup/file"
	"github.com/toniphan21/go-mapper-gen/internal/util"
	"github.com/toniphan21/go-mapper-gen/pkl"
)

const maxCountToHideSetup = 10

func runTest(cmd TestCmd, execPath string, logger *slog.Logger) error {
	runner := &testRunner{
		logger:   logger,
		execPath: execPath,
	}
	return runner.Run(cmd)
}

type testFile struct {
	fileName  string
	testCases []gen.MarkdownTestCase
}

type failedTest struct {
	fileName string
	testName string
}

func (t *failedTest) makeRunCmd(execPath string) string {
	cmd := execPath
	if strings.Index(execPath, "go-build") != -1 {
		cmd = "go run ./cmd/go-mapper-gen"
	}
	cmd += " test " + t.fileName + " -n='" + strings.ToLower(t.testName) + "'"
	return cmd
}

type testRunner struct {
	disableLog bool
	execPath   string
	logger     *slog.Logger
}

func (r *testRunner) logError(msg string, args ...any) {
	r.logger.Error(msg, args...)
}

func (r *testRunner) logWarn(msg string, args ...any) {
	r.logger.Warn(msg, args...)
}

func (r *testRunner) logInfo(msg string, args ...any) {
	if r.disableLog {
		return
	}
	r.logger.Info(msg, args...)
}

func (r *testRunner) logDebug(msg string, args ...any) {
	if r.disableLog {
		return
	}
	r.logger.Debug(msg, args...)
}

func (r *testRunner) print(msg string) {
	r.logger.Info(msg)
}

func (r *testRunner) Run(cmd TestCmd) error {
	gen.UseTestVersion()
	util.SetTabSize(cmd.TabSize)

	var total, passed, failed int
	var failedTests []failedTest
	var tempDirs []string
	defer func() {
		if len(tempDirs) > 0 {
			r.logInfo("deleting temporary directories")
			for _, dir := range tempDirs {
				r.logDebug("\tdeleted temporary directory " + dir)
				_ = os.RemoveAll(dir)
			}
			r.logInfo("")
		}

		if total == passed {
			r.print(util.ColorGreen(fmt.Sprintf("Result: passed all %d total tests", passed)))
			r.print("")
		} else {
			r.print(util.ColorRedBold(fmt.Sprintf("Result: %d failed, passed %d/%d total tests.", failed, passed, total)))
			r.print("")
			r.print("Run failed test command:")
			r.print("")
			for _, ft := range failedTests {
				r.print("\t" + ft.makeRunCmd(r.execPath))
			}
			r.print("")
		}
	}()

	testFiles, count := r.makeTestFiles(cmd)
	showSetup := false
	if count < maxCountToHideSetup {
		showSetup = true
	} else {
		showSetup = cmd.ShowSetup
	}

	if !showSetup {
		r.disableLog = true
	}

	total = count

	for _, tf := range testFiles {
		if r.disableLog {
			r.print(util.ColorBlue(tf.fileName))
		}
		r.logInfo("running test for " + util.ColorBlue(tf.fileName))

		for _, tc := range tf.testCases {
			r.logInfo("\ttesting " + util.ColorCyan(tc.Name))
			tempDir, err := os.MkdirTemp("", "mapper-*")
			if err != nil {
				r.logError("\tError creating temp dir:", slog.Any("error", err))
			}
			r.logInfo("\t\tcreated temporary directory at " + util.ColorWhite(tempDir))

			isSuccess := r.runTestCase(cmd, tc, tempDir)
			if isSuccess {
				tempDirs = append(tempDirs, tempDir)
				if r.disableLog {
					r.print(util.ColorGreen("\t\u2714 passed ") + tc.Name)
				}
				passed++
				continue
			}

			failedTests = append(failedTests, failedTest{
				fileName: tf.fileName,
				testName: tc.Name,
			})
			failed++
			if r.disableLog {
				r.print(util.ColorRed("\t\u2718 failed ") + tc.Name)
			}
		}

		if r.disableLog {
			r.print("")
		}
	}

	return nil
}

func (r *testRunner) makeTestFiles(cmd TestCmd) ([]testFile, int) {
	var count int
	var result []testFile
	for _, inputFile := range cmd.Files {
		stat, err := os.Stat(inputFile)
		if err != nil {
			r.logError(util.ColorRed(err.Error()))
			continue
		}

		if stat.IsDir() {
			r.logWarn(util.ColorBlue(inputFile) + " is a directory, skipped")
			continue
		}

		content, err := os.ReadFile(inputFile)
		if err != nil {
			r.logError(util.ColorRed(err.Error()))
			continue
		}

		var matched []gen.MarkdownTestCase
		tcs := gen.Test.ParseMarkdownTestCases(content)
		if strings.TrimSpace(cmd.Name) != "" {
			for _, v := range tcs {
				if v.IsNameMatch(cmd.Name) {
					matched = append(matched, v)
				}
			}
		} else {
			matched = tcs
		}

		count = count + len(matched)
		result = append(result, testFile{
			fileName:  inputFile,
			testCases: matched,
		})
	}
	return result, count
}

func (r *testRunner) runTestCase(cmd TestCmd, mdTestCase gen.MarkdownTestCase, tempDir string) bool {
	if mdTestCase.GoModFileContent != nil {
		r.logInfo("\t\tcopied go.mod file")
		if err := r.writeTestFile(tempDir, "go.mod", mdTestCase.GoModFileContent); err != nil {
			r.logError("\t\t" + util.ColorRed(err.Error()))
			return false
		}
		util.PrintFileWithFunction("", mdTestCase.GoModFileContent, func(s string) {
			r.logDebug("\t\t" + s)
		})

		directDependencies, err := gen.Test.ParseDirectDependencies(mdTestCase.GoModFileContent)
		if err != nil {
			r.logError("\t\t" + util.ColorRed(err.Error()))
			return false
		}

		if len(directDependencies) > 0 {
			if err = setup.ExecuteGoGet(tempDir, directDependencies); err != nil {
				r.logError("\t\t" + util.ColorRed(err.Error()))
				return false
			}
			r.logInfo("\t\tinstalled dependencies")
		}
	}

	if mdTestCase.GoSumFileContent != nil {
		r.logInfo("\t\tcopied go.sum file")
		if err := r.writeTestFile(tempDir, "go.sum", mdTestCase.GoSumFileContent); err != nil {
			r.logError(util.ColorRed(err.Error()))
			return false
		}
		util.PrintFileWithFunction("", mdTestCase.GoSumFileContent, func(s string) {
			r.logDebug("\t\t" + s)
		})
	}

	setupOk := true

	if mdTestCase.PklDevFileContent != nil {
		configFile := file.MakePklDevConfigFile(
			pkl.AmendsPath(cmd.PklBaseURL),
			pkl.ImportPaths(cmd.PklBaseURL),
			"mapper.pkl",
			[]string{string(mdTestCase.PklDevFileContent)},
		)

		if err := r.writeTestFile(tempDir, configFile.FilePath(), configFile.FileContent()); err != nil {
			r.logError("\t\t" + util.ColorRed(err.Error()))
			setupOk = false
		}
		r.logInfo("\t\tmade pkl config file " + util.ColorWhite(configFile.FilePath()))
		util.PrintFileWithFunction("", configFile.FileContent(), func(s string) {
			r.logDebug("\t\t" + s)
		})
	}

	if !setupOk {
		return false
	}

	for fn, fc := range mdTestCase.SourceFiles {
		if err := r.writeTestFile(tempDir, fn, fc); err != nil {
			r.logError(util.ColorRed(err.Error()))
			setupOk = false
			continue
		}
		r.logInfo("\t\tcopied source file " + util.ColorWhite(fn))

		util.PrintFileWithFunction("", fc, func(s string) {
			r.logDebug("\t\t" + s)
		})

		if fn == "go.mod" {
			directDependencies, err := gen.Test.ParseDirectDependencies(fc)
			if err != nil {
				r.logError("\t\t" + util.ColorRed(err.Error()))
				setupOk = false
				continue
			}

			if len(directDependencies) > 0 {
				if err = setup.ExecuteGoGet(tempDir, directDependencies); err != nil {
					r.logError("\t\t" + util.ColorRed(err.Error()))
					setupOk = false
					continue
				}
				r.logInfo("\t\tinstalled dependencies")
			}
		}
	}

	if !setupOk {
		return false
	}

	r.logInfo("\t\tgenerating mapper")
	glogger := gen.NewNoopLogger()
	if cmd.LogGenerate {
		glogger = r.logger
	}
	err := runGenerate(GenerateCmd{WorkingDir: tempDir}, glogger)
	if err != nil {
		r.logError(util.ColorRed(err.Error()))
		return false
	}

	isSuccess := true
	for fn, fc := range mdTestCase.GoldenFiles {
		out, err := r.readTestFile(tempDir, fn)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				r.logInfo(util.ColorRed(fmt.Sprintf("\t\texpected golden file %v but file does not exist", fn)))
				isSuccess = false
				continue
			}

			r.logError(err.Error())
			continue
		}

		if !r.compareFileContent(out, fc) {
			r.logInfo(util.ColorRed("\t\tgolden file content does not match expectation"))
			util.PrintDiffWithFunction("expected", fc, "generated", out, func(s string) {
				r.logInfo("\t\t" + s)
			})
			isSuccess = false

			_ = r.writeTestFile(tempDir, fn+".golden", fc)
			continue
		}

		r.logDebug("\t\tpassed with " + util.ColorYellow("golden file ") + util.ColorWhite(fn))
		util.PrintFileWithFunction("", fc, func(s string) {
			r.logDebug("\t\t" + s)
		})
	}

	if isSuccess {
		r.logInfo("\t\tput temporary directory to clean up list")
		if cmd.EmitCode {
			if err := r.emitTestCode(mdTestCase, tempDir); err != nil {
				r.logError("\t\t" + util.ColorRed(err.Error()))
			}
		}
		r.logInfo(util.ColorGreen("\t\t\u2714 passed"))
		r.logInfo("")
		return true
	}

	r.logInfo("\t\tsaved test case to file test.md")
	r.logInfo("\t\ttemporary directory will not be deleted")
	_ = r.writeTestFile(tempDir, "test.md", []byte(mdTestCase.Content))

	r.logInfo(util.ColorRed("\t\t\u2718 failed"))
	r.logInfo("")
	return false
}

func (r *testRunner) emitTestCode(testCase gen.MarkdownTestCase, tempDir string) error {
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

		if err = r.copyDir(tempDir, dst); err != nil {
			return err
		}

		_ = r.writeTestFile(dst, "README.md", []byte(strings.Join(readme, "\n")))
		r.logInfo("\t\temit code to " + examplePath)
	}
	return nil
}

func (r *testRunner) writeTestFile(testDir string, filePath string, fileContent []byte) error {
	fp := filepath.Join(testDir, filePath)
	dir := filepath.Dir(fp)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	return os.WriteFile(fp, fileContent, 0600)
}

func (r *testRunner) readTestFile(testDir string, filePath string) ([]byte, error) {
	fp := filepath.Join(testDir, filePath)
	return os.ReadFile(fp)
}

func (r *testRunner) compareFileContent(left, right []byte) bool {
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

func (r *testRunner) copyDir(src string, dst string) error {
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

		return r.copyFile(path, targetPath)
	})
}

func (r *testRunner) copyFile(srcFile, dstFile string) error {
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
