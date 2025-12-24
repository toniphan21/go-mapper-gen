package util

import "fmt"

func ColorBlack(text string) string {
	return fmt.Sprintf("\u001B[0;30m%s\u001B[0m", text)
}

func ColorRed(text string) string {
	return fmt.Sprintf("\u001B[0;31m%s\u001B[0m", text)
}

func ColorRedBold(text string) string {
	return fmt.Sprintf("\u001B[1;31m%s\u001B[0m", text)
}

func ColorGreen(text string) string {
	return fmt.Sprintf("\u001B[0;32m%s\u001B[0m", text)
}

func ColorYellow(text string) string {
	return fmt.Sprintf("\u001B[0;33m%s\u001B[0m", text)
}

func ColorBlue(text string) string {
	return fmt.Sprintf("\u001B[0;34m%s\u001B[0m", text)
}

func ColorPurple(text string) string {
	return fmt.Sprintf("\u001B[0;35m%s\u001B[0m", text)
}

func ColorCyan(text string) string {
	return fmt.Sprintf("\u001B[0;36m%s\u001B[0m", text)
}

func ColorWhite(text string) string {
	return fmt.Sprintf("\u001B[0;37m%s\u001B[0m", text)
}
