// Code generated from Pkl module `gomappergen.mapper`. DO NOT EDIT.
package mapper

// Configuration for mapping between a specific source struct
// and a specific target struct.
//
// Values defined here override corresponding values in `BaseMapper`
// when explicitly set. Unset (null) values fall back to the base
// configuration defaults.
type Struct struct {
	// Target package for the generated mapping code.
	//
	// Overrides package level target_pkg when set.
	TargetPkg *string `pkl:"target_pkg"`

	// Name of the target struct.
	//
	// If not set, the target struct name is inferred from context.
	TargetStructName *string `pkl:"target_struct_name"`

	// Source package containing the source struct.
	//
	// Overrides package level source_pkg when set.
	SourcePkg *string `pkl:"source_pkg"`

	// Name of the source struct.
	//
	// If not set, the source struct name is inferred from context.
	SourceStructName *string `pkl:"source_struct_name"`

	// Template for the source-to-target mapping function name.
	//
	// Overrides package level source_to_target_function_name when set.
	SourceToTargetFunctionName *string `pkl:"source_to_target_function_name"`

	// Template for the target-to-source mapping function name.
	//
	// Overrides package level source_from_target_function_name when set.
	SourceFromTargetFunctionName *string `pkl:"source_from_target_function_name"`

	// Template for the decorator function name.
	//
	// Overrides package level decorate_function_name when set.
	DecorateFunctionName *string `pkl:"decorate_function_name"`

	// Controls pointer usage in generated mapping code.
	//
	// - "none": Neither source nor target is treated as a pointer.
	// - "source-only": Source struct is passed as a pointer.
	// - "target-only": Target struct is returned or populated as a pointer.
	// - "both": Both source and target are pointers.
	Pointer string `pkl:"pointer"`

	// Field-level mapping configuration.
	//
	// Controls how fields are matched and transformed between
	// source and target structs.
	Fields Fields `pkl:"fields"`

	// Whether to prefer getter methods over direct field access
	// on the source struct.
	//
	// Overrides package level use_getter_if_available when set.
	UseGetterIfAvailable *bool `pkl:"use_getter_if_available"`

	// Whether to generate source-to-target mapping code.
	// When false, only target-to-source mapping is generated.
	//
	// Overrides package level generate_source_to_target when set.
	GenerateSourceToTarget *bool `pkl:"generate_source_to_target"`

	// Whether to generate target-to-source mapping code.
	// When false, only source-to-target mapping is generated.
	//
	// Overrides package level generate_source_from_target when set.
	GenerateSourceFromTarget *bool `pkl:"generate_source_from_target"`
}
