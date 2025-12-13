package gomappergen

import "strings"

func replacePlaceholders(template string, vars map[string]string) string {
	result := template
	for k, v := range vars {
		result = strings.ReplaceAll(result, k, v)
	}
	return result
}
