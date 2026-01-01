package gomappergen

import (
	"context"
	"fmt"
	"go/types"
	"log/slog"

	"github.com/dave/jennifer/jen"
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
	Metadata  SymbolMetadata
}

type SymbolMetadata struct {
	IsVariable bool
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
	CanConvert(ctx LookupContext, targetType, sourceType types.Type) bool

	// ConvertField emits the assignment code that writes the converted value from
	// source into target. Returning nil suppresses emission.
	//
	// GeneratorUtil may be used to build complex expression trees.
	ConvertField(ctx ConverterContext, target, source Symbol) jen.Code
}

// ConverterContext provides shared capabilities and state for converters
// during code generation. It embeds context.Context to support cancellation
// and timeouts defined by the generator.
type ConverterContext interface {
	context.Context
	LookupContext

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
	Logger() *slog.Logger

	// Run executes runner if the context has not been cancelled.
	// If the context is done, Run returns nil and runner is not executed.
	// This allows converters to respect generator-defined timeouts without
	// explicitly checking ctx.Done().
	Run(converter Converter, runner func() jen.Code) jen.Code

	// EmitTraceComments indicates whether the converter should emit trace comments
	// for debugging or inspection purposes. It returns false by default.
	EmitTraceComments() bool
}

type converterContext struct {
	context.Context
	jenFile           *jen.File
	parser            Parser
	logger            *slog.Logger
	currentVarCount   int
	lookupContext     *lookupContext
	emitTraceComments bool
}

func (c *converterContext) LookUp(current Converter, targetType, sourceType types.Type) (Converter, error) {
	return c.lookupContext.LookUp(current, targetType, sourceType)
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

func (c *converterContext) Run(converter Converter, runner func() jen.Code) jen.Code {
	select {
	case <-c.Done():
		return nil
	default:
	}

	code := runner()

	var out jen.Code
	if c.EmitTraceComments() {
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

func (c *converterContext) EmitTraceComments() bool {
	return c.emitTraceComments
}

func (c *converterContext) Logger() *slog.Logger {
	return c.logger
}

func (c *converterContext) resetVarCount() {
	c.currentVarCount = 0
}

func (c *converterContext) resetLookupContext() {
	c.lookupContext.converters = nil
}

var _ ConverterContext = (*converterContext)(nil)
var _ LookupContext = (*converterContext)(nil)

// ConverterInfo describes metadata about a Converter.
// It is primarily used for debugging, logging, and trace comments
// in generated code.
type ConverterInfo struct {
	Name                 string
	ShortForm            string
	ShortFormDescription string
}
