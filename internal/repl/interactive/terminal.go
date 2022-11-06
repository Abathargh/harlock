package interactive

import (
	"fmt"
	"golang.org/x/term"
	"os"
	"os/exec"
)

const (
	OFFSET = 5

	clearScreen    = "\033[2J\033[H"
	startNextLine  = "\033[1E"
	moveBackNCols  = "\033[%dD"
	deleteUntilEnd = "\033[0K"
	resetLine      = "\033[5G\033[0K\n"
	rightByOne     = "\033[1C"
	leftByOne      = "\033[1D"
)

func ClearScreen() {
	fmt.Print(clearScreen)
}

func Home() {
	fmt.Print(resetLine)
}

func MoveRight() {
	fmt.Print(rightByOne)
}

func MoveLeft() {
	fmt.Print(leftByOne)
}

func Exit() {
	fmt.Printf("\nbye\n")
	cmd := exec.Command("stty", "sane")
	cmd.Stdin = os.Stdin
	_ = cmd.Run()
}

// instead of printing every time just advance by one and print char
// go next line and reset in case, clean on enter or history

func Update(line *Line, last rune) {
	count, width := getRowsCount(line.Size())
	if count != 0 && line.Size()+OFFSET%width == 0 {
		fmt.Print(startNextLine)
	}

	if line.Position() == line.Size() {
		fmt.Print(string(last))
	} else {
		fmt.Print(deleteUntilEnd)
		bufferTip := line.AsStringFromCursor()
		fmt.Print(bufferTip)
		fmt.Printf(moveBackNCols, len(bufferTip)-1)
	}
}

func PrintLine(line *Line) {
	cleanUpToRepl(line)
	fmt.Printf("%s", line.AsString())

	// re-positioning must account row nums
	count, w := getRowsCount(line.Size())
	if count != 0 {
		fmt.Printf("\033[%dB", count)
	}
	fmt.Printf("\033[%dG", line.Position()%w+OFFSET)
}

func cleanUpToRepl(line *Line) {
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
