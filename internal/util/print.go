package util

import (
	"fmt"
	"strconv"
	"strings"
)

type printFileOptions struct {
	BorderLeft  bool
	BorderRight bool
	LineNumber  bool
	Color       bool
}

func PrintGeneratedFile(name string, data []byte) {
	printFile(name, data, printFileOptions{})
}

func PrintFile(name string, data []byte) {
	printFile(name, data, printFileOptions{BorderLeft: true, LineNumber: true, Color: true})
}

func PrintDiff(leftName string, leftData []byte, rightName string, rightData []byte) {
	left, leftLength := makePrintable(leftName, leftData, printFileOptions{BorderLeft: true, LineNumber: true, Color: true, BorderRight: true})
	right, _ := makePrintable(rightName, rightData, printFileOptions{BorderLeft: true, LineNumber: true, Color: true})

	ll := len(left)
	rl := len(right)
	l := ll
	if rl > l {
		l = rl
	}

	for i := 0; i < l; i++ {
		if i < ll && i < rl {
			line := left[i] + " " + right[i]
			fmt.Println(line)
			continue
		}

		if i < ll {
			fmt.Println(left[i])
			continue
		}

		line := strings.Repeat(" ", leftLength+1) + right[i]
		fmt.Println(line)
	}
	fmt.Println()
}

func printFile(name string, data []byte, opts printFileOptions) {
	lines, _ := makePrintable(name, data, opts)
	for _, line := range lines {
		fmt.Println(line)
	}
	fmt.Println()
}
func makePrintable(name string, data []byte, opts printFileOptions) ([]string, int) {
	var maxLineLength int
	var out []string

	fileName := name
	if opts.Color {
		fileName = ColorBlue(fileName)
	}

	if !opts.BorderLeft && !opts.LineNumber {
		out = append(out, fmt.Sprintf("--- file: %v", fileName))
		out = append(out, string(data))
		out = append(out, fmt.Sprintf("--- end file: %v", fileName))
		return out, maxLineLength
	}

	lines := strings.Split(string(data), "\n")
	maxLength := 0
	for _, line := range lines {
		ll := lineLength(line)
		if ll > maxLength {
			maxLength = ll
		}
	}

	linesLen := len(lines)
	if linesLen == 0 {
		out = append(out, fmt.Sprintf("%v |", fileName))
		return out, maxLineLength
	}

	fileNameLen := len(name)
	nodLineNumber := len(strconv.Itoa(linesLen))
	for i, line := range lines {
		lineNumber := fmt.Sprintf("%*d", nodLineNumber, i+1)
		if !opts.LineNumber {
			lineNumber = ""
		}

		mll := lineLength(line) + len(lineNumber) + fileNameLen + 7
		if mll > maxLineLength {
			maxLineLength = mll
		}

		borderLeft := " |\t"
		if !opts.BorderLeft {
			borderLeft = ""
			mll -= 4
		}

		borderRight := " |"
		if !opts.BorderRight {
			borderRight = ""
			mll -= 2
		}

		fnPlaceholder := strings.Repeat("·", fileNameLen)

		if opts.Color {
			lineNumber = ColorWhite(lineNumber)
			borderLeft = ColorWhite(borderLeft)
			borderRight = ColorWhite(borderRight)
			fnPlaceholder = ColorWhite(fnPlaceholder)
		}

		ll := lineLength(line)
		switch i {
		case 0:
			fallthrough
		case linesLen - 1:
			if opts.BorderRight {
				out = append(out, fmt.Sprintf("%v %v%v%v%-*s%v", fileName, lineNumber, borderLeft, line, maxLength-ll, "", borderRight))
			} else {
				out = append(out, fmt.Sprintf("%v %v%v%v", fileName, lineNumber, borderLeft, line))
			}

		default:
			if opts.BorderRight {
				out = append(out, fmt.Sprintf("%v %v%v%v%-*s%v", fnPlaceholder, lineNumber, borderLeft, line, maxLength-ll, "", borderRight))
			} else {
				out = append(out, fmt.Sprintf("%v %v%v%v", fnPlaceholder, lineNumber, borderLeft, line))
			}
		}
	}
	return out, maxLineLength
}

func lineLength(line string) int {
	ll := 0
	for _, r := range line {
		if r == '\t' {
			ll += 2
			continue
		}
		ll += 1
	}
	return ll
}
