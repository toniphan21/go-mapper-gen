package gomappergen

import (
	"context"
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

func newSymbol(varName string, fieldName string, typ types.Type) Symbol {
	if fieldName == "" {
		return Symbol{VarName: varName, FieldName: nil, Type: typ}
	}
	var fn = fieldName
	return Symbol{VarName: varName, FieldName: &fn, Type: typ}
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

func (s Symbol) toGetterSymbol(getterName string) Symbol {
	if s.FieldName == nil || getterName == "" {
		return s
	}
	fieldName := getterName
	fieldName += "()"
	return Symbol{VarName: s.VarName, FieldName: &fieldName, Type: s.Type}
}

// Converter defines the contract for converting a source value or field
// into a target value or field during code generation.
type Converter interface {
	// Init is called once before code generation starts.
	// It allows the converter to initialize internal state, validate assumptions,
	// or prepare data required during code generation. Init must not emit code.
	// If no initialization is required, the implementation may be a no-op.
	Init(parser Parser, config Config)

	// Info returns metadata describing the converter.
	Info() ConverterInfo

	// CanConvert reports whether this converter can convert a value of
	// sourceType into a value of targetType. Implementations typically use
	// TypeUtil to perform type analysis.
	//
	// CanConvert must be pure and must not modify file state.
	CanConvert(ctx ConverterContext, targetType, sourceType types.Type) bool

	// ConvertField emits the assignment code that writes the converted value from
	// source into target. Returning nil suppresses emission.
	//
	// GeneratorUtil may be used to build complex expression trees.
	ConvertField(ctx ConverterContext, target, source Symbol, opts ConverterOption) jen.Code
}

// ConverterContext provides shared capabilities and state for converters
// during code generation. It embeds context.Context to support cancellation
// and timeouts defined by the generator.
type ConverterContext interface {
	context.Context

	// JenFile returns the current jennifer file used for code generation.
	// Converters may append generated code to this file.
	JenFile() *jen.File

	// Parser returns the type parser used to inspect and analyze Go types
	// during conversion.
	Parser() Parser

	// NextVarName returns a unique variable name for use in generated code.
	// It guarantees no name collisions within the current generation scope.
	NextVarName() string

	// Logger returns a slog handler that can be used for logging during
	// code generation.
	Logger() slog.Handler

	// LookUp searches the global converter registry for a converter that
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
	LookUp(current Converter, targetType, sourceType types.Type) (Converter, bool)

	// Run executes runner if the context has not been cancelled.
	// If the context is done, Run returns nil and runner is not executed.
	// This allows converters to respect generator-defined timeouts without
	// explicitly checking ctx.Done().
	Run(converter Converter, opts ConverterOption, runner func() jen.Code) jen.Code
}

type converterContext struct {
	context.Context
	jenFile         *jen.File
	parser          Parser
	logger          slog.Handler
	currentVarCount int
}

func (c *converterContext) JenFile() *jen.File {
	return c.jenFile
}

func (c *converterContext) Parser() Parser {
	return c.parser
}

func (c *converterContext) NextVarName() string {
	v := fmt.Sprintf("v%d", c.currentVarCount)
	c.currentVarCount++
	return v
}

func (c *converterContext) Run(converter Converter, opts ConverterOption, runner func() jen.Code) jen.Code {
	select {
	case <-c.Done():
		return nil
	default:
	}

	code := runner()

	var out jen.Code
	if opts.EmitTraceComments {
		info := converter.Info()
		out = jen.Comment(fmt.Sprintf("%s generated code start", info.Name)).Line().
			Add(code).Line().
			Add(jen.Comment(fmt.Sprintf("%s generated code end", info.Name)))
	} else {
		out = code
	}

	select {
	case <-c.Done():
		return nil
	default:
		return out
	}
}

func (c *converterContext) Logger() slog.Handler {
	return c.logger
}

func (c *converterContext) LookUp(current Converter, targetType, sourceType types.Type) (Converter, bool) {
	for _, reg := range converters {
		if current != nil {
			if current == reg.converter {
				continue
			}

			if normalizeConverterType(current) == reg.typ {
				continue
			}
		}

		if reg.converter.CanConvert(c, targetType, sourceType) {
			return reg.converter, true
		}
	}
	return nil, false
}

func (c *converterContext) resetVarCount() {
	c.currentVarCount = 0
}

var _ ConverterContext = (*converterContext)(nil)

// ConverterInfo describes metadata about a Converter.
// It is primarily used for debugging, logging, and trace comments
// in generated code.
type ConverterInfo struct {
	Name                 string
	Description          string
	ShortForm            string
	ShortFormDescription string
}

type ConverterOption struct {
	EmitTraceComments bool
}

// ---

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

func RegisteredConverterCount() int {
	return len(converters)
}

func PrintRegisteredConverters(logger *slog.Logger) {
	shortFormBuffer := 0
	for _, c := range converters {
		info := c.converter.Info()
		l := len(info.ShortForm)
		if l > shortFormBuffer {
			shortFormBuffer = l
		}
	}

	for _, v := range converters {
		builtin := ""
		if v.builtIn {
			builtin = util.ColorCyan("[built-in]")
		}

		info := v.converter.Info()
		line := fmt.Sprintf("%-*s %v", shortFormBuffer+1, info.ShortForm, info.ShortFormDescription)

		logger.Info(
			fmt.Sprintf("%s %10s %s",
				util.ColorBlue(fmt.Sprintf("%5s", strconv.Itoa(v.priority))),
				builtin,
				line,
			),
		)
	}
}

func ClearAllRegisteredConverters() {
	converters = []*registeredConverter{}
}

func RegisterAllBuiltinConverters() {
	cf := BuiltInConverterConfig{}
	cf.EnableAll()

	RegisterBuiltinConverters(cf)
}

func RegisterBuiltinConverters(config BuiltInConverterConfig) {
	priority := 0
	if config.UseIdentical {
		registerConverter(BuiltinConverters.IdenticalType, priority, true)
		priority++
	}

	if config.UseSlice {
		registerConverter(BuiltinConverters.Slice, priority, true)
		priority++
	}

	if config.UseTypeToPointer {
		registerConverter(BuiltinConverters.TypeToPointer, priority, true)
		priority++
	}

	if config.UsePointerToType {
		registerConverter(BuiltinConverters.PointerToType, priority, true)
		priority++
	}

	if config.UseFunctions {
		registerConverter(BuiltinConverters.Functions, priority, true)
		priority++
	}
}

func InitAllRegisteredConverters(parser Parser, parsedConfig Config) {
	for _, v := range converters {
		v.converter.Init(parser, parsedConfig)
	}
}

func findConverter(ctx ConverterContext, targetType, sourceType types.Type) (Converter, bool) {
	for _, reg := range converters {
		if reg.converter.CanConvert(ctx, targetType, sourceType) {
			return reg.converter, true
		}
	}
	return nil, false
}

type builtinConverters struct {
	IdenticalType Converter
	Slice         Converter
	TypeToPointer Converter
	PointerToType Converter
	Functions     Converter
}

var BuiltinConverters = builtinConverters{
	IdenticalType: &identicalTypeConverter{},
	Slice:         &sliceConverter{},
	TypeToPointer: &typeToPointerConverter{},
	PointerToType: &pointerToTypeConverter{},
	Functions:     &functionsConverter{},
}
