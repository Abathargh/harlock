//go:build linux && interrepl

package repl

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/Abathargh/harlock/internal/evaluator"
	"github.com/Abathargh/harlock/internal/lexer"
	"github.com/Abathargh/harlock/internal/object"
	"github.com/Abathargh/harlock/internal/parser"
	"github.com/eiannone/keyboard"
	"io"
	"os"
	"os/exec"
	"strings"
)

const OFFSET = 5
const PROMPT = ">>> "
const FOLLOWING = "... "

type HistoryMgr struct {
	list []string
	pos  int
	end  int
}

func (mgr *HistoryMgr) Push(cmd string) {
	mgr.list = append(mgr.list, cmd)

}

type Direction uint8

const (
	DirLeft Direction = iota
	DirRight
)

type Line struct {
	buffer []rune
	pos    int
	end    int
}

func (l *Line) Position() int {
	return l.pos
}

func (l *Line) Move(direction Direction) bool {
	if direction == DirLeft && l.pos != 0 {
		l.pos--
		return true
	}

	if direction == DirRight && l.pos != l.end {
		l.pos++
		return true
	}
	return false
}

func (l *Line) Reset() {
	l.buffer = make([]rune, 0)
	l.pos = 0
	l.end = 0
}

func (l *Line) Backspace() {
	if l.pos != 0 {
		l.buffer = append(l.buffer[:l.pos-1], l.buffer[l.pos:]...)
		l.pos--
		l.end--
	}
}

func (l *Line) Canc() {
	if l.pos != l.end {
		l.buffer = append(l.buffer[:l.pos], l.buffer[l.pos+1:]...)
		l.end--
	}
}

func (l *Line) Character(c rune) {
	if l.end == l.pos {
		l.buffer = append(l.buffer, c)
		l.pos++
		l.end++
		return
	}

	if len(l.buffer) == cap(l.buffer) {
		newBuffer := make([]rune, len(l.buffer), cap(l.buffer)*2)
		copy(newBuffer, l.buffer)
		l.buffer = newBuffer
	}

	l.buffer = append(l.buffer[:l.pos+1], l.buffer[l.pos:]...)
	l.buffer[l.pos] = c
	l.pos++
	l.end++
}

func (l *Line) AsString() string {
	return string(l.buffer[:l.end])
}

func PrintLine(line *Line) {
	fmt.Print("\033[5G\033[0K")
	fmt.Printf("%s", line.AsString())
	fmt.Printf("\033[%dG", line.Position()+OFFSET)
}

func Setup(command chan string) {
	line := Line{}
	//historyMgr := HistoryMgr{}
	keysEvents, err := keyboard.GetKeys(10)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = keyboard.Close()
	}()

	for {
		event := <-keysEvents
		if event.Err != nil {
			panic(event.Err)
		}

		if event.Rune == 0 {
			// ctrl char
			switch event.Key {
			case keyboard.KeyCtrlC:
				line.Reset()
				fmt.Print("\033[5G\033[0K")
				command <- "\n"
			case keyboard.KeyCtrlD:
				fmt.Printf("\nbye\n")
				_ = exec.Command("reset").Run()
				os.Exit(1)
			case keyboard.KeyArrowRight:
				if line.Move(DirRight) {
					fmt.Print("\033[1C")
				}
			case keyboard.KeyArrowLeft:
				if line.Move(DirLeft) {
					fmt.Print("\033[1D")
				}
			case keyboard.KeyArrowUp:
				fallthrough
			case keyboard.KeyArrowDown:
				// TODO
			case keyboard.KeyDelete:
				line.Canc()
				PrintLine(&line)
			case keyboard.KeyBackspace:
				fallthrough
			case keyboard.KeyBackspace2:
				line.Backspace()
				PrintLine(&line)
			case keyboard.KeySpace:
				line.Character(' ')
				PrintLine(&line)
			case keyboard.KeyEnter:
				fmt.Println()
				command <- line.AsString() + "\n"
				line.Reset()
			default:
				fmt.Printf("\033[0K\rYou pressed: rune %q, key %X", event.Rune, event.Key)
			}
		} else {
			line.Character(event.Rune)
			PrintLine(&line)
		}
	}
}

func Start(_ io.Reader, output io.Writer) {
	command := make(chan string)
	go Setup(command)

	env := object.NewEnvironment()

	var buf strings.Builder
	exprStarted := false

	for {
		if !exprStarted {
			_, _ = fmt.Fprintf(output, PROMPT)
		} else {
			_, _ = fmt.Fprintf(output, FOLLOWING)
		}

		currCommand := <-command

		currLine := strings.TrimSpace(currCommand)
		switch {
		case currLine == "" && !exprStarted:
			continue
		case currLine == "" && exprStarted:
			exprStarted = false
			if !parseAndEval(output, buf.String(), env) {
				buf.Reset()
				continue
			}
			buf.Reset()
		case currLine != "" && !exprStarted:
			if !strings.HasSuffix(currLine, "{") {
				parseAndEval(output, currLine, env)
				continue
			}
			exprStarted = true
			fallthrough
		case currLine != "" && exprStarted:
			buf.WriteString(currLine)
			buf.WriteString("\n")
		}
	}
}

func parseAndEval(output io.Writer, input string, env *object.Environment) bool {
	l := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(input)))
	p := parser.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		printParserErrors(output, p.Errors())
		return false
	}

	evaluatedProg := evaluator.Eval(program, env)
	if evaluatedProg != nil {
		_, _ = io.WriteString(output, evaluatedProg.Inspect())
		_, _ = io.WriteString(output, "\n")
	}
	return true
}

func printParserErrors(writer io.Writer, errors []string) {
	for _, errorMsg := range errors {
		_, _ = io.WriteString(writer, fmt.Sprintf("%s\n", errorMsg))
	}
}
