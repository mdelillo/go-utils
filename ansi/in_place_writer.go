package ansi

import (
	"fmt"
	"io"
	"strings"
)

type InPlaceWriter struct {
	Writer      io.Writer
	LineCount   int
	initialized bool
}

// Init will ensure the terminal has enough lines for output
// The previous output will not be cleared if there are not enough blank lines
func (w *InPlaceWriter) Init() (int, error) {
	totalByteCount := 0

	byteCount, err := fmt.Fprint(w.Writer, strings.Repeat("\n", w.LineCount))
	totalByteCount += byteCount
	if err != nil {
		return totalByteCount, err
	}

	byteCount, err = MoveCursorUp(w.Writer, w.LineCount)
	totalByteCount += byteCount
	if err != nil {
		return totalByteCount, err
	}

	byteCount, err = SaveCursorPosition(w.Writer)
	totalByteCount += byteCount
	if err != nil {
		return totalByteCount, err
	}

	w.initialized = true

	return totalByteCount, nil
}

// Write will clear the terminal before writing
// Call once with all desired output instead of calling multiple times
func (w *InPlaceWriter) Write(bytes []byte) (int, error) {
	totalByteCount := 0

	if !w.initialized {
		byteCount, err := w.Init()
		totalByteCount += byteCount
		if err != nil {
			return totalByteCount, err
		}
	}

	byteCount, err := RestoreCursorPosition(w.Writer)
	totalByteCount += byteCount
	if err != nil {
		return totalByteCount, err
	}

	byteCount, err = ClearAfterCursor(w.Writer)
	totalByteCount += byteCount
	if err != nil {
		return totalByteCount, err
	}

	byteCount, err = w.Writer.Write(bytes)
	totalByteCount += byteCount
	if err != nil {
		return totalByteCount, err
	}

	return totalByteCount, nil
}
