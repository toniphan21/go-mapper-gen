// Code generated from Pkl module `gomappergen.Config`. DO NOT EDIT.
package config

type MapperStruct struct {
	TargetPkg *string `pkl:"target_pkg"`

	TargetStructName *string `pkl:"target_struct_name"`

	SourcePkg *string `pkl:"source_pkg"`

	SourceStructName *string `pkl:"source_struct_name"`

	SourceToTargetFunctionName *string `pkl:"source_to_target_function_name"`

	SourceFromTargetFunctionName *string `pkl:"source_from_target_function_name"`

	DecorateFunctionName *string `pkl:"decorate_function_name"`

	Pointer string `pkl:"pointer"`

	Fields Fields `pkl:"fields"`

	UseGetterIfAvailable *bool `pkl:"use_getter_if_available"`

	GenerateSourceToTarget bool `pkl:"generate_source_to_target"`

	GenerateSourceFromTarget bool `pkl:"generate_source_from_target"`
}
