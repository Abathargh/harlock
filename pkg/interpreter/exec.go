package interpreter

import (
	"bufio"
	"io"
	"os"

	"github.com/Abathargh/harlock/internal/object"

	"github.com/Abathargh/harlock/internal/evaluator"
	"github.com/Abathargh/harlock/internal/lexer"
	"github.com/Abathargh/harlock/internal/parser"
)

func Exec(filename string, output io.Writer) []string {
	o, err := os.Open(filename)
	if err != nil {
		return []string{err.Error()}
	}

	env := object.NewEnvironment()
	l := lexer.NewLexer(bufio.NewReader(o))
	p := parser.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		return p.Errors()
	}

	evaluatedProg := evaluator.Eval(program, env)
	if evaluatedProg != nil {
		if _, ok := evaluatedProg.(*object.Error); ok {
			_, _ = io.WriteString(output, evaluatedProg.Inspect())
			_, _ = io.WriteString(output, "\n")
		}
	}
	return nil
}
