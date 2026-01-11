// Code generated from Pkl module `gomappergen.mapper`. DO NOT EDIT.
package decoratormode

import (
	"encoding"
	"fmt"
)

type DecoratorMode string

const (
	Adaptive DecoratorMode = "adaptive"
	Always   DecoratorMode = "always"
	Never    DecoratorMode = "never"
)

// String returns the string representation of DecoratorMode
func (rcv DecoratorMode) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(DecoratorMode)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for DecoratorMode.
func (rcv *DecoratorMode) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "adaptive":
		*rcv = Adaptive
	case "always":
		*rcv = Always
	case "never":
		*rcv = Never
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid DecoratorMode`, str)
	}
	return nil
}
