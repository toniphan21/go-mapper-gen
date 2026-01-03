
package mapping

import "strings"

func TrimName(in string) string {
	return strings.TrimSpace(in)
}

type helper struct{}

func (h *helper) Trim(in string) string {
	return strings.TrimSpace(in)
}

var Helper = &helper{}
