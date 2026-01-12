// Code generated from Pkl module `gomappergen.mapper`. DO NOT EDIT.
package pointer

import (
	"encoding"
	"fmt"
)

type Pointer string

const (
	None       Pointer = "none"
	SourceOnly Pointer = "source-only"
	TargetOnly Pointer = "target-only"
	Both       Pointer = "both"
)

// String returns the string representation of Pointer
func (rcv Pointer) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(Pointer)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for Pointer.
func (rcv *Pointer) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "none":
		*rcv = None
	case "source-only":
		*rcv = SourceOnly
	case "target-only":
		*rcv = TargetOnly
	case "both":
		*rcv = Both
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid Pointer`, str)
	}
	return nil
}
