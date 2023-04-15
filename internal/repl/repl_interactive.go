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
	"strings"
)

// This version of the repl includes a "more interactive"
// approach that allows for the usual normal stuff related
// to navigation within a repl, like moving within a line
// that is being written, deleting characters, using the
// up/down keys to navigate the history.
//
// Known problems:
// - Line-wrapping is broken, this is mainly tied to the usage
//   of the ANSI escape characters and the current line management.
//

const PROMPT = ">>> "
const FOLLOWING = "... "

func Setup(command chan string) {
	term := interactive.NewTerminal(interactive.NewLine())
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
				term.Home()
				command <- "\n"
			case keyboard.KeyCtrlD:
				term.ExitIfBufferEmpty()
			case keyboard.KeyCtrlL:
				if cleared := term.ClearIfBufferEmpty(); cleared {
					command <- "\n"
				}
			case keyboard.KeyArrowRight:
				term.MoveRight()
			case keyboard.KeyArrowLeft:
				term.MoveLeft()
			case keyboard.KeyArrowUp:
				term.PreviousCmd()
			case keyboard.KeyArrowDown:
				term.NextCmd()
			case keyboard.KeyDelete:
				term.Delete()
			case keyboard.KeyBackspace:
				fallthrough
			case keyboard.KeyBackspace2:
				term.BackSpace()
			case keyboard.KeySpace:
				term.PutRune(' ')
			case keyboard.KeyEnter:
				cmd := term.Enter()
				command <- cmd
			default:
				fmt.Printf("\033[0K\rYou pressed: rune %q, key %X", event.Rune, event.Key)
			}
		} else {
			term.PutRune(event.Rune)
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
