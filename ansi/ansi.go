package ansi

import (
	"fmt"
	"io"
)

func MoveCursorUp(writer io.Writer, lines int) (int, error) {
	return fmt.Fprintf(writer, "\033[%dA", lines)
}

func SaveCursorPosition(writer io.Writer) (int, error) {
	return fmt.Fprint(writer, "\033[s")
}

func RestoreCursorPosition(writer io.Writer) (int, error) {
	return fmt.Fprint(writer, "\033[u")
}

func ClearAfterCursor(writer io.Writer) (int, error) {
	return fmt.Fprint(writer, "\033[J")
}
