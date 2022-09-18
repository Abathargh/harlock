//go:build !interrepl || (interrepl && !linux)

package repl

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/Abathargh/harlock/internal/evaluator"
	"github.com/Abathargh/harlock/internal/lexer"
	"github.com/Abathargh/harlock/internal/object"
	"github.com/Abathargh/harlock/internal/parser"
)

const PROMPT = ">>> "
const FOLLOWING = "... "

func Start(input io.Reader, output io.Writer) {
	scanner := bufio.NewScanner(input)
	env := object.NewEnvironment()

	var buf strings.Builder
	exprStarted := false

	for {
		if !exprStarted {
			_, _ = fmt.Fprintf(output, PROMPT)
		} else {
			_, _ = fmt.Fprintf(output, FOLLOWING)
		}
		if !scanner.Scan() {
			return
		}

		line := strings.TrimSpace(scanner.Text())
		switch {
		case line == "" && !exprStarted:
			continue
		case line == "" && exprStarted:
			exprStarted = false
			if !parseAndEval(output, buf.String(), env) {
				buf.Reset()
				continue
			}
			buf.Reset()
		case line != "" && !exprStarted:
			if !strings.HasSuffix(line, "{") {
				parseAndEval(output, line, env)
				continue
			}
			exprStarted = true
			fallthrough
		case line != "" && exprStarted:
			buf.WriteString(line)
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
