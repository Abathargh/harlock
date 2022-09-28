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
	"github.com/Abathargh/harlock/internal/repl/interactive"
	"github.com/eiannone/keyboard"
	"io"
	"os"
	"os/exec"
	"strings"
)

// This version of the repl includes a "more interactive"
// approach that allows for the usual normal stuff related
// to navigation within a repl, like moving within a line
// that is being written, deleting characters, using the
// up/down keys to navigate the history.
//
// Known problems:
// - The buffer used to manipulate the current line is
//   inefficient and is a best-effort kind of implementation;
// - Line-wrapping is broken, this is mainly tied to the usage
//   of the ANSI escape characters and the current line management.
//
// A nicer approach may consist in getting the width of the terminal
// and managing the buffer with reference to that, but for now I am
// keeping it simple and with fewer dependencies.
//

const PROMPT = ">>> "
const FOLLOWING = "... "

func Setup(command chan string) {
	line := interactive.NewLine()
	historyMgr := interactive.HistoryMgr{}
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
				fmt.Print("\033[5G\033[0K\n")
				command <- "\n"
			case keyboard.KeyCtrlD:
				fmt.Printf("\nbye\n")
				cmd := exec.Command("stty", "sane")
				cmd.Stdin = os.Stdin
				_ = cmd.Run()
				os.Exit(1)
			case keyboard.KeyArrowRight:
				if line.Move(interactive.DirRight) {
					fmt.Print("\033[1C")
				}
			case keyboard.KeyArrowLeft:
				if line.Move(interactive.DirLeft) {
					fmt.Print("\033[1D")
				}
			case keyboard.KeyArrowUp:
				cmd := historyMgr.GetPrevious()
				if len(cmd) != 0 {
					line.SetBuffer(cmd)
					interactive.PrintLine(&line)
				}
			case keyboard.KeyArrowDown:
				cmd := historyMgr.GetNext()
				if len(cmd) != 0 {
					line.SetBuffer(cmd)
				}
				interactive.PrintLine(&line)
			case keyboard.KeyDelete:
				line.Delete()
				interactive.PrintLine(&line)
			case keyboard.KeyBackspace:
				fallthrough
			case keyboard.KeyBackspace2:
				line.Backspace()
				interactive.PrintLine(&line)
			case keyboard.KeySpace:
				line.Character(' ')
				interactive.Update(' ', &line)
			case keyboard.KeyEnter:
				fmt.Println()
				l := line.AsString()
				command <- l + "\n"
				line.Reset()
				historyMgr.Push(l)
			default:
				fmt.Printf("\033[0K\rYou pressed: rune %q, key %X", event.Rune, event.Key)
			}
		} else {
			line.Character(event.Rune)
			interactive.Update(event.Rune, &line)
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
