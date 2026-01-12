// Code generated from Pkl module `gomappergen.mapper`. DO NOT EDIT.
package mapper

type BuiltInConverter struct {
	EnableIdentical bool `pkl:"enable_identical"`

	EnableSlice bool `pkl:"enable_slice"`

	EnableTypeToPointer bool `pkl:"enable_type_to_pointer"`

	EnablePointerToType bool `pkl:"enable_pointer_to_type"`

	EnableNumeric bool `pkl:"enable_numeric"`

	EnableFunctions bool `pkl:"enable_functions"`

	EnableFunctionsStrict bool `pkl:"enable_functions_strict"`

	Library BuiltInLibraryConverter `pkl:"library"`
}
