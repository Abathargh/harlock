package interactive

import (
	"fmt"
	"golang.org/x/term"
	"math"
	"os"
	"os/exec"
)

const (
	OFFSET = 5

	clearScreen        = "\033[2J\033[H"
	startNextLine      = "\033[1E"
	moveBackNCols      = "\033[%dD"
	deleteUntilLineEnd = "\033[0K"
	deleteUntilEnd     = "\033[0J"
	resetLine          = "\033[5G\033[0K\n"
	cleanWholeRow      = "\033[2K\r"
	cleanUpToPrompt    = "\033[5G\033[0J"
	endOfPrevRow       = "\033[1F\033[%dC"
	upNRows            = "\033[%dA"
	downNRows          = "\033[%dB"
	toNthColumn        = "\033[%dG"
	startOfNextRow     = "\033[1E"
	rightByOne         = "\033[1C"
	leftByOne          = "\033[1D"
	moveXY             = "\033[{%d};{%d}H"
)

type Terminal struct {
	line        *Line
	history     *HistoryMgr
	historyLast []rune
}

func NewTerminal(line *Line) *Terminal {
	return &Terminal{
		line:    line,
		history: &HistoryMgr{},
	}
}

func (t *Terminal) Home() {
	t.line.Reset()
	fmt.Print(resetLine)
}

func (t *Terminal) PutRune(r rune) {
	t.line.Character(r)
	t.update(r)
}

func (t *Terminal) Enter() string {
	fmt.Println()
	cmd := t.line.AsString() + "\n"
	t.line.Reset()
	t.history.Push(cmd)
	return cmd
}

func (t *Terminal) MoveRight() {
	t.moveCursorRight()
}

func (t *Terminal) MoveLeft() {
	_ = t.moveCursorLeft(t.line.MoveLeft)
}

func (t *Terminal) Move(x int, y int) {
	fmt.Printf(moveXY, x, y)
}

func (t *Terminal) moveCursorRight() {
	lastPos := t.line.Position()
	if t.line.MoveRight() {
		lastCount, _ := getRowsCount(lastPos)
		currCount, _ := getRowsCount(t.line.Position())
		if lastCount != currCount {
			fmt.Print(startOfNextRow)
		}
		fmt.Print(rightByOne)
	}
}

// returns true if we actually moved, second bool is if we changed line while moving
func (t *Terminal) moveCursorLeft(moveFunc func() bool) bool {
	lastPos := t.line.Position() // get current position before moving
	if moveFunc() {
		// cursor not at the start of the terminal
		lastCount, width := getRowsCount(lastPos)
		currCount, _ := getRowsCount(t.line.Position())

		// did we jump back one row?
		if lastCount != currCount {
			fmt.Printf(endOfPrevRow, width)
			return true
		}
		fmt.Print(leftByOne)
		return true
	}
	// cursor was at the very start of the terminal, no-op
	return false
}

// TODO
func (t *Terminal) Delete() {
	t.line.Delete()
	t.printLine()
}

func (t *Terminal) BackSpace() {
	moved := t.moveCursorLeft(t.line.Backspace)
	startPos := t.line.Position()
	if !moved {
		return
	}

	currRow, width := getRowsCount(startPos)
	totRows := ((t.line.Size() + OFFSET) / width) + 1

	fmt.Print(deleteUntilEnd) // delete current line until the end
	if startPos == t.line.Size() {
		return
	}

	// let us compute the position of the cursor relative
	// to the row the cursor is in
	relativeToRowPos := (startPos + OFFSET) % width

	for idx := currRow; idx < totRows; idx++ {
		rowLen := width
		if idx == currRow {
			rowLen -= relativeToRowPos
		}

		if idx == 0 {
			rowLen -= OFFSET
		}

		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Errored the the following values: ")
				fmt.Printf("startRow: %d\nrowLen: %d\ntotLen: %d\n", startPos, rowLen, t.line.Size())
				fmt.Printf("Complete error: %s\n", r)
			}
		}()

		rowLen = int(math.Min(float64(rowLen), float64(t.line.Size())))
		fmt.Print(t.line.AsStringFromCursor()[startPos:rowLen])
		startPos += rowLen
	}

}

func (t *Terminal) PreviousCmd() {
	cmd := t.history.GetPrevious(t.line.AsString())
	if len(cmd) != 0 {
		t.line.SetBuffer(cmd)
		t.printLine()
	}
}

func (t *Terminal) NextCmd() {
	cmd := t.history.GetNext()
	if len(cmd) != 0 {
		t.line.SetBuffer(cmd)
	}
	t.printLine()
}

func (t *Terminal) ExitIfBufferEmpty() {
	if t.line.Size() != 0 {
		return
	}
	fmt.Printf("\nbye\n")
	cmd := exec.Command("stty", "sane")
	cmd.Stdin = os.Stdin
	_ = cmd.Run()
	os.Exit(1)
}

func (t *Terminal) ClearIfBufferEmpty() bool {
	if t.line.Size() != 0 {
		return false
	}
	t.line.Reset()
	fmt.Print(clearScreen)
	return true
}

// instead of printing every time just advance by one and print char
// go next line and reset in case, clean on enter or history

func (t *Terminal) update(last rune) {
	count, width := getRowsCount(t.line.Size())
	if count != 0 && (t.line.Size()+OFFSET)%width == 0 {
		fmt.Print(startNextLine)
	}

	if t.line.Position() == t.line.Size() {
		fmt.Print(string(last))
	} else {
		fmt.Print(deleteUntilLineEnd)
		bufferTip := t.line.AsStringFromCursor()
		fmt.Print(bufferTip)
		fmt.Printf(moveBackNCols, len(bufferTip)-1)
	}
}

func (t *Terminal) printLine() {
	t.cleanUpToRepl()

	// re-positioning must account row nums
	count, w := getRowsCount(t.line.Size())

	if count != 0 {
		fmt.Printf("%s", t.line.AsString()[count*w-OFFSET:])
	} else {
		fmt.Printf("%s", t.line.AsString())
	}
}

func (t *Terminal) cleanUpToRepl() {
	numRows, _ := getRowsCount(t.line.Size())
	if numRows != 0 {
		fmt.Printf(cleanWholeRow)
		return
	}
	fmt.Print(cleanUpToPrompt)
}

// returns the number of rows used at this moment, together with  the
// terminal width (row size in chars)
func getRowsCount(charNum int) (int, int) {
	w, _, _ := term.GetSize(int(os.Stdout.Fd()))
	totSize := charNum + OFFSET
	return totSize / w, w
}
