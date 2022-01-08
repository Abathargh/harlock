package interpreter

import (
	"bufio"
	"io"

	"github.com/Abathargh/harlock/internal/object"

	"github.com/Abathargh/harlock/internal/evaluator"
	"github.com/Abathargh/harlock/internal/lexer"
	"github.com/Abathargh/harlock/internal/parser"
)

func Exec(r io.Reader, output io.Writer) []string {
	env := object.NewEnvironment()
	l := lexer.NewLexer(bufio.NewReader(r))
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
