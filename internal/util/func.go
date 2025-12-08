package util

import "strings"

const CommentWidth = 80

func WrapComment(s string) []string {
	return WrapText(s, CommentWidth)
}

func WrapText(s string, width int) []string {
	var lines []string
	words := strings.Fields(s)

	if len(words) == 0 {
		return []string{""}
	}

	var current string
	for _, word := range words {
		if len(current)+len(word)+1 > width {
			lines = append(lines, strings.TrimSpace(current))
			current = word
		} else {
			if current == "" {
				current = word
			} else {
				current += " " + word
			}
		}
	}
	if current != "" {
		lines = append(lines, current)
	}

	return lines
}
