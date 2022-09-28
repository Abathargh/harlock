package interactive

import (
	"fmt"
	"golang.org/x/term"
	"os"
)

const (
	OFFSET = 5

	nLinesUpOffsetCol = "\033[%dA\033[5G\033[0J"
	startNextLine     = "\033[1E"
)

// instead of printing every time just advance by one and print char
// go next line and reset in case, clean on enter or history

func Update(last rune, line *Line) {
	count, width := getRowsCount(line.Size())
	if count != 0 && line.Size()+OFFSET%width == 0 {
		fmt.Print(startNextLine)
	}
	fmt.Print(string(last))
}

func PrintLine(line *Line) {
	CleanUpToRepl(line)
	fmt.Printf("%s", line.AsString())

	// re-positioning must account row nums
	count, w := getRowsCount(line.Size())
	if count != 0 {
		fmt.Printf("\033[%dB", count)
	}
	fmt.Printf("\033[%dG", line.Position()%w+OFFSET)
}

func CleanUpToRepl(line *Line) {
	numRows, _ := getRowsCount(line.Size())
	if numRows != 0 {
		fmt.Printf("\033[%dA", numRows)
	}
	fmt.Print("\033[5G\033[0J")
}

func getRowsCount(charNum int) (int, int) {
	w, _, _ := term.GetSize(int(os.Stdout.Fd()))
	totSize := charNum + OFFSET
	return totSize / w, w
}
