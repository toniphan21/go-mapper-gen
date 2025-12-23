// Code generated from Pkl module `gomappergen.Config`. DO NOT EDIT.
package config

type BuiltInConverter struct {
	EnableIdentical bool `pkl:"enable_identical"`

	EnableSlice bool `pkl:"enable_slice"`

	EnableTypeToPointer bool `pkl:"enable_type_to_pointer"`

	EnablePointerToType bool `pkl:"enable_pointer_to_type"`

	EnableNumeric bool `pkl:"enable_numeric"`

	EnableFunctions bool `pkl:"enable_functions"`

	Library BuiltInLibraryConverter `pkl:"library"`
}
