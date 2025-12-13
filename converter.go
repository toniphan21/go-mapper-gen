package gomappergen

import (
	"fmt"
	"go/types"
	"log/slog"
	"reflect"
	"slices"
	"strconv"

	"github.com/dave/jennifer/jen"
	"github.com/toniphan21/go-mapper-gen/internal/util"
)

// Symbol represents a reference to either a variable or a field expression
// used during code generation.
//
// VarName is the name of the local variable (e.g. "src").
// FieldName is optional. If nil, the Symbol refers to the entire variable
// (e.g. "src"). If non-nil, the Symbol refers to a field of that variable
// (e.g. "src.Field"). This allows converters to handle both top-level
// assignments and assignments targeting specific struct fields.
//
// Type is the Go type of the referenced value. Converters use this when
// determining whether they can perform a conversion.
//
// Examples:
//
//	Symbol{VarName: "src", FieldName: nil}                // "src"
//	Symbol{VarName: "dst", FieldName: &("Name")}          // "dst.Name"
type Symbol struct {
	VarName   string
	FieldName *string
	Type      types.Type
}

type ConverterOption struct {
	EmitTraceComments bool
}

func (s Symbol) Expr() *jen.Statement {
	if s.FieldName == nil {
		return jen.Id(s.VarName)
	}
	return jen.Id(s.VarName).Dot(*s.FieldName)
}

func (s Symbol) ToIndexedSymbol(idx string) Symbol {
	if s.FieldName == nil {
		return Symbol{VarName: s.VarName + "[" + idx + "]", FieldName: nil, Type: s.Type}
	}
	fieldName := *s.FieldName
	fieldName += "[" + idx + "]"
	return Symbol{VarName: s.VarName, FieldName: &fieldName, Type: s.Type}
}

// Converter defines a pluggable conversion rule used by the mapper generator.
// Implementations decide whether they support converting between two Go types
// and, if so, generate the appropriate code snippets using jennifer.
//
// A converter participates in three phases:
//
//  1. CanConvert:
//     Determines whether this converter supports converting from sourceType
//     to targetType. It should NOT generate code or perform side effects.
//
//  2. Before:
//     Generates optional "preamble" code that must appear before the final
//     assignment. This is commonly used to declare temporary variables or
//     prepare intermediate values. The returned jen.Code may be nil, in which
//     case the caller simply emits nothing.
//
//     currentVarCount is supplied by the caller and represents the current
//     counter for temporary variable allocation. The returned nextVarCount
//     must be the updated counter after any temp variables introduced by this
//     converter. This keeps temp variable naming consistent across the whole
//     generated file.
//
//  3. Assign:
//     Generates the actual assignment code that writes the converted value
//     from the source Symbol into the target Symbol. As with Before, the
//     returned jen.Code may be nil, meaning the converter chooses not to emit
//     anything.
//
// Converters must not assume they are the only converter in the system.
// Priority or selection logic is handled externally by the caller.
type Converter interface {
	// Print returns a human-readable name or label for this converter.
	// This is typically used for debugging or logging and has no impact on
	// generation semantics.
	Print() string

	// CanConvert reports whether this converter can convert a value of
	// sourceType into a value of targetType. Implementations typically use
	// TypeUtil to perform type analysis.
	//
	// CanConvert must be pure and must not modify file state.
	CanConvert(targetType, sourceType types.Type) bool

	// Before emits optional pre-assignment code. If the returned jen.Code is
	// nil, the caller simply emits nothing.
	//
	// file is the jennifer file being built.
	// target and source describe the left-hand and right-hand expressions.
	// currentVarCount tracks temporary variable allocation.
	//
	// nextVarCount must be returned to keep temp variable numbering consistent.
	// GeneratorUtil may be used to simplify emission of helpers or temps.
	Before(file *jen.File, target, source Symbol, currentVarCount int, opt ConverterOption) (code jen.Code, nextVarCount int)

	// Assign emits the assignment code that writes the converted value from
	// source into target. Returning nil suppresses emission.
	//
	// GeneratorUtil may be used to build complex expression trees.
	Assign(file *jen.File, target, source Symbol, opt ConverterOption) jen.Code
}

type registeredConverter struct {
	converter Converter
	typ       reflect.Type
	priority  int
	builtIn   bool
}

func normalizeConverterType(v Converter) reflect.Type {
	if v == nil {
		return nil
	}
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

var converters []*registeredConverter

func registerConverter(converter Converter, priority int, isBuiltIn bool) {
	converters = append(converters, &registeredConverter{
		converter: converter,
		typ:       normalizeConverterType(converter),
		priority:  priority,
		builtIn:   isBuiltIn,
	})

	slices.SortFunc(converters, func(a, b *registeredConverter) int {
		return a.priority - b.priority
	})
}

// RegisterConverter adds a new Converter to the global converter registry.
//
// Converters are selected by the mapper generator based on:
//  1. Whether they report true from CanConvert(...)
//  2. Their assigned priority
//
// The priority determines the ordering between converters that can handle the
// same source and target types. A lower numeric value represents a higher
// priority. For example, a converter with priority 0 will be chosen before a
// converter with priority 10, assuming both return true from CanConvert.
//
// Registering a converter does not trigger any generation work. It merely
// appends the converter to the internal registry so that the code generator
// can discover it later.
//
// Converters should be stateless and safe for repeated reuse. The registry
// is typically read during generator initialization or when resolving the
// appropriate converter for a given assignment.
//
// Example:
//
//	RegisterConverter(&StringToIntConverter{}, 5)
//
// Passing the same converter instance multiple times is allowed but generally
// discouraged unless intentional.
func RegisterConverter(converter Converter, priority int) {
	registerConverter(converter, priority, false)
}

// LookUpConverter searches the global converter registry for a converter that
// can convert a value of sourceType to targetType, excluding the provided
// currentConverter (if non-nil).
//
// This helper is intended for converter implementations that need to
// delegate or reuse existing conversion rules. A common use-case is a
// SliceConverter that converts []T -> []V by looking up a converter for
// T -> V and then generating per-element conversion code.
//
// Selection rules (implementation contract):
//  1. The registry is scanned for converters c where c.CanConvert(targetType, sourceType)
//     returns true.
//  2. The currentConverter parameter is excluded from consideration to avoid
//     trivial self-selection (if currentConverter == nil, no exclusion occurs).
//  3. From the remaining candidates, the converter with the highest priority
//     (your package's ordering rule: lower numeric value = higher priority)
//     is chosen. If multiple converters share the same priority, the selection
//     must be deterministic (for example: registration order or stable sorting).
//
// Return value:
//   - (Converter, true) if a matching converter was found.
//   - (nil, false) if no converter in the registry can perform the conversion.
func LookUpConverter(current Converter, targetType, sourceType types.Type) (Converter, bool) {
	for _, reg := range converters {
		if current != nil {
			if current == reg.converter {
				continue
			}

			if normalizeConverterType(current) == reg.typ {
				continue
			}
		}

		if reg.converter.CanConvert(targetType, sourceType) {
			return reg.converter, true
		}
	}
	return nil, false
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
	registerConverter(&sliceConverter{}, 1, true)
	registerConverter(&typeToPointerConverter{}, 2, true)
	registerConverter(&pointerToTypeConverter{}, 3, true)
}
