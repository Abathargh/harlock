package repl

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/Abathargh/harlock/internal/evaluator"

	"github.com/Abathargh/harlock/internal/parser"

	"github.com/Abathargh/harlock/internal/lexer"
)

const PROMPT = ">>"

func Start(input io.Reader, output io.Writer) {
	scanner := bufio.NewScanner(input)

	for {
		_, _ = fmt.Fprintf(output, PROMPT)
		if !scanner.Scan() {
			return
		}

		line := scanner.Text() + "\n"
		l := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(line)))
		p := parser.NewParser(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(output, p.Errors())
			continue
		}

		evaluatedProg := evaluator.Eval(program)
		if evaluatedProg != nil {
			_, _ = io.WriteString(output, evaluatedProg.Inspect())
			_, _ = io.WriteString(output, "\n")
		}
	}
}

func printParserErrors(writer io.Writer, errors []string) {
	for _, errorMsg := range errors {
		_, _ = io.WriteString(writer, fmt.Sprintf("\t%s\n", errorMsg))
	}
}
