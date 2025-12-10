package gomappergen

import (
	"fmt"
	"go/types"
	"log/slog"
	"slices"
	"strconv"

	"github.com/dave/jennifer/jen"
	"github.com/toniphan21/go-mapper-gen/internal/util"
)

type Symbol struct {
	VarName   string
	FieldName string
	Type      types.Type
}

// Converter defines the contract for generating custom assignment code between a Source and Target type.
// A single Converter instance handles the logic for a specific type pair, regardless of the direction
// (Source -> Target or Target -> Source). The generator logic will handle reversing the assignment.
type Converter interface {
	// Print called to show what the converter can do, mainly for debugging the generator's lookup logic.
	// This method should return a clear, human-readable string describing the conversion rule.
	Print() string

	// CanConvert dynamically checks if this Converter is capable of handling the conversion
	// between the two provided types. This is the first method called by the generator
	// during lookup to select the appropriate conversion logic.
	//
	// Params:
	//   - targetType: the types.Type being written to (the destination type).
	//   - sourceType: The types.Type being read from (the source type).
	//
	// Returns:
	//   - bool: True if this Converter should be used to generate the assignment code.
	CanConvert(targetType, sourceType types.Type) bool

	// Before generates code that must run before the final assignment line.
	// This is used for complex setup like variable declarations (v0, v1), type parsing, or error checks.
	//
	// Params:
	//   - file: the jen.File instance being generated (used for Qualifiers).
	//   - target: the Symbol representing the destination field (the final receiver of the mapped value).
	//   - source: the Symbol representing the input field (the value being read).
	//   - currentVarCount: the first unused integer index for creating unique local variable names (e.g., 0 for "v0").
	//
	// Returns:
	//   - code: the generated setup code (jen.Code).
	//   - nextVarCount: the next unused variable count (currentVarCount + N).
	Before(file *jen.File, target, source Symbol, currentVarCount int) (code jen.Code, nextVarCount int)

	// Assign generates the final assignment line.
	// This code typically assigns the result from the source/a temporary variable (created in Before)
	// to the target field.
	//
	// Params:
	//   - target: the Symbol representing the destination field (out.FieldName).
	//   - source: the Symbol representing the input field (in.FieldName or a temporary variable like v0).
	//
	// Returns:
	//   jen.Code: The assignment code (e.g., out.Field = v0).
	Assign(file *jen.File, target, source Symbol) jen.Code
}

type registeredConverter struct {
	converter Converter
	priority  int
	builtIn   bool
}

var converters []*registeredConverter

func registerConverter(converter Converter, priority int, isBuiltIn bool) {
	converters = append(converters, &registeredConverter{
		converter: converter,
		priority:  priority,
		builtIn:   isBuiltIn,
	})

	slices.SortFunc(converters, func(a, b *registeredConverter) int {
		return a.priority - b.priority
	})
}

// RegisterConverter registers a custom type Converter with the code generator.
// The generator will iterate through all registered converters, ordered by priority,
// and use the first one whose CanConvert method returns true.
//
// Params:
//   - converter: the implementation of the Converter interface.
//   - priority: an integer defining the lookup order. Lower numbers have higher precedence.
func RegisterConverter(converter Converter, priority int) {
	registerConverter(converter, priority, false)
}

func PrintRegisteredConverters(logger *slog.Logger) {
	for _, v := range converters {
		builtin := ""
		if v.builtIn {
			builtin = util.ColorCyan("[built-in]")
		}

		logger.Info(
			fmt.Sprintf("%s %10s %s",
				util.ColorBlue(fmt.Sprintf("%5s", strconv.Itoa(v.priority))),
				builtin,
				v.converter.Print(),
			),
		)
	}
}

func RegisterAllBuiltinConverters() {
	RegisterBaseConverters()
}

func RegisterBaseConverters() {
	registerConverter(&identicalTypeConverter{}, 0, true)
	registerConverter(&typeToPointerConverter{}, 1, true)
	registerConverter(&pointerToTypeConverter{}, 2, true)
}
