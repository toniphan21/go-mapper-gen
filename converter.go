package gomappergen

import (
	"go/types"

	"github.com/dave/jennifer/jen"
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
	//   - generatedVarCount: the next unused variable count (currentVarCount + N).
	Before(file jen.File, target, source Symbol, currentVarCount int) (code jen.Code, generatedVarCount int)

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
	Assign(target, source Symbol) jen.Code
}

// RegisterConverter registers a Converter for a specific type pair.
//
// By registering a converter for (SourceType, TargetType), the generator assumes this logic
// can be used for both directions:
//  1. Forward Mapping: SourceType -> TargetType
//  2. Reverse Mapping: TargetType -> SourceType
//
// The generator will be responsible for flipping the target and source symbols
// (e.g., in.Field, out.Field) when applying the converter in the reverse direction.
//
// Params:
//
//   - sourceType: the types.Type being read from in the forward direction (e.g., string).
//   - targetType: the types.Type being written to in the forward direction (e.g., pgtype.Text).
//   - converter:  the implementation of the Converter interface.
func RegisterConverter(sourceType, targetType types.Type, converter Converter) {
	// ...
}

func init() {
}
