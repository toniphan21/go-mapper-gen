package gomappergen

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/apple/pkl-go/pkl"
	"github.com/stretchr/testify/assert"
	"github.com/toniphan21/go-mapper-gen/internal/setup"
	"github.com/toniphan21/go-mapper-gen/internal/setup/file"
)

func TestParseConfig_FileNotFound(t *testing.T) {
	cf, err := ParseConfig(path.Join(t.TempDir(), "not-found.pkl"))

	assert.Nil(t, cf)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestParseConfig_Invalid(t *testing.T) {
	cases := []struct {
		name            string
		config          []string
		expectedErrorIs error
	}{
		{
			name: "missing source_pkg",
			config: []string{
				`packages {`,
				`	["github.com/example/repo"] {`,
				`	}`,
				`}`,
			},
			expectedErrorIs: &pkl.EvalError{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := setup.SourceCode(t, setup.PklLibFiles(), file.PklDevConfigFile(tc.config...))
			result, err := ParseConfig(filepath.Join(dir, "dev/config.pkl"))

			assert.ErrorIs(t, err, tc.expectedErrorIs)
			assert.Nil(t, result)
		})
	}
}

type expectedConfig struct {
	Output                 *Output
	InterfaceName          *string
	ImplementationName     *string
	ConstructorName        *string
	DecoratorInterfaceName *string
	GenerateGoDoc          *bool
}

type expectedStruct struct {
	TargetPkgPath            *string
	TargetStructName         string
	SourcePkgPath            string
	SourceStructName         string
	SourceToTargetFuncName   *string
	SourceFromTargetFuncName *string
	DecorateFuncName         *string
	Pointer                  *Pointer
	FieldNameMatch           *FieldNameMatch
	GenerateSourceToTarget   *bool
	GenerateSourceFromTarget *bool
}

func buildConfig(override *expectedConfig, structs ...expectedStruct) Config {
	defaultCf := Config{
		Output:                 Default.Output,
		InterfaceName:          Default.InterfaceName,
		ImplementationName:     Default.ImplementationName,
		ConstructorName:        Default.ConstructorName,
		DecoratorInterfaceName: Default.DecoratorInterfaceName,
		GenerateGoDoc:          true,
		Structs:                nil,
	}

	if len(structs) == 0 {
		return defaultCf
	}

	result := defaultCf
	if override != nil {
		if override.Output != nil {
			result.Output = *override.Output
		}
		if override.InterfaceName != nil {
			result.InterfaceName = *override.InterfaceName
		}
		if override.ImplementationName != nil {
			result.ImplementationName = *override.ImplementationName
		}
		if override.ConstructorName != nil {
			result.ConstructorName = *override.ConstructorName
		}
		if override.DecoratorInterfaceName != nil {
			result.DecoratorInterfaceName = *override.DecoratorInterfaceName
		}
		if override.GenerateGoDoc != nil {
			result.GenerateGoDoc = *override.GenerateGoDoc
		}
	}

	for _, v := range structs {
		defaultStructCf := Struct{
			TargetPkgPath:            Placeholder.CurrentPackage,
			SourceToTargetFuncName:   Default.SourceToTargetFuncName,
			SourceFromTargetFuncName: Default.SourceFromTargetFuncName,
			DecorateFuncName:         Default.DecorateFuncName,
			Pointer:                  PointerNone,
			FieldNameMatch:           FieldNameMatchIgnoreCase,
			GenerateSourceToTarget:   true,
			GenerateSourceFromTarget: true,
		}
		item := defaultStructCf

		if v.TargetPkgPath != nil {
			item.TargetPkgPath = *v.TargetPkgPath
		}

		item.TargetStructName = v.TargetStructName
		item.SourcePkgPath = v.SourcePkgPath
		item.SourceStructName = v.SourceStructName

		if v.SourceToTargetFuncName != nil {
			item.SourceToTargetFuncName = *v.SourceToTargetFuncName
		}
		if v.SourceFromTargetFuncName != nil {
			item.SourceFromTargetFuncName = *v.SourceFromTargetFuncName
		}
		if v.DecorateFuncName != nil {
			item.DecorateFuncName = *v.DecorateFuncName
		}
		if v.Pointer != nil {
			item.Pointer = *v.Pointer
		}
		if v.FieldNameMatch != nil {
			item.FieldNameMatch = *v.FieldNameMatch
		}
		if v.GenerateSourceToTarget != nil {
			item.GenerateSourceToTarget = *v.GenerateSourceToTarget
		}
		if v.GenerateSourceFromTarget != nil {
			item.GenerateSourceFromTarget = *v.GenerateSourceFromTarget
		}
		result.Structs = append(result.Structs, item)
	}
	return result
}

func TestParseConfig(t *testing.T) {
	cases := []struct {
		name     string
		config   []string
		expected map[string]Config
	}{
		{
			name: "empty config",
			config: []string{
				`packages {`,
				`	["github.com/example/repo"] {`,
				`		source_pkg = "{CurrentPackage}/source"`,
				`	}`,
				`}`,
			},
			expected: map[string]Config{},
		},

		{
			name: "minimal config",
			config: []string{
				`packages {`,
				`	["github.com/example/repo"] {`,
				`		source_pkg = "{CurrentPackage}/source"`,
				`		structs {`,
				`			["Target"] {}`,
				`		}`,
				`	}`,
				`}`,
			},
			expected: map[string]Config{
				"github.com/example/repo": buildConfig(nil, expectedStruct{
					TargetStructName: "Target",
					SourceStructName: "Target",
					SourcePkgPath:    "{CurrentPackage}/source",
				}),
			},
		},

		{
			name: "source_struct_name",
			config: []string{
				`packages {`,
				`	["github.com/example/repo"] {`,
				`		source_pkg = "{CurrentPackage}/source"`,
				`		structs {`,
				`			["Target"] { source_struct_name = "Source" }`,
				`		}`,
				`	}`,
				`}`,
			},
			expected: map[string]Config{
				"github.com/example/repo": buildConfig(nil, expectedStruct{
					TargetStructName: "Target",
					SourceStructName: "Source",
					SourcePkgPath:    "{CurrentPackage}/source",
				}),
			},
		},

		{
			name: "override source_pkg in struct",
			config: []string{
				`packages {`,
				`	["github.com/example/repo"] {`,
				`		source_pkg = "{CurrentPackage}/source"`,
				`		structs {`,
				`			["Target"] { source_struct_name = "Source" }`,
				`			["OverrideSourcePkg"] {`,
				`				source_struct_name = "Source"`,
				`				source_pkg = "{CurrentPackage}/another-source"`,
				`			}`,
				`		}`,
				`	}`,
				`}`,
			},
			expected: map[string]Config{
				"github.com/example/repo": buildConfig(nil, expectedStruct{
					TargetStructName: "Target",
					SourceStructName: "Source",
					SourcePkgPath:    "{CurrentPackage}/source",
				}, expectedStruct{
					TargetStructName: "OverrideSourcePkg",
					SourceStructName: "Source",
					SourcePkgPath:    "{CurrentPackage}/another-source",
				}),
			},
		},
		// ---
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := setup.SourceCode(t, setup.PklLibFiles(), file.PklDevConfigFile(tc.config...))
			result, err := ParseConfig(filepath.Join(dir, "dev/config.pkl"))

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}
