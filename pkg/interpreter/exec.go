// Package interpreter implements the public API that
// can be used to interact with the harlock runtime
// from other applications or the CLI tool.
package interpreter

import (
	"bufio"
	"io"

	"github.com/Abathargh/harlock/internal/object"

	"github.com/Abathargh/harlock/internal/evaluator"
	"github.com/Abathargh/harlock/internal/lexer"
	"github.com/Abathargh/harlock/internal/parser"
)

// Version represents the current harlock version
var Version = ""

// Exec reads a script from the passed reader, executes it and
// sends the generated output to the passed writer. If the parsing
// phase fails, it returns an array of string containing the parsing
// errors, or nil otherwise.
func Exec(r io.Reader, output io.Writer, args ...string) []string {
	env := object.NewEnvironment()
	l := lexer.NewLexer(bufio.NewReader(r))
	p := parser.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		return p.Errors()
	}

	// The interpreter inherits the args from the process call
	argsArray := &object.Array{Elements: make([]object.Object, len(args))}
	for idx, arg := range args {
		argsArray.Elements[idx] = &object.String{Value: arg}
	}
	env.Set("args", argsArray)

	evaluatedProg := evaluator.Eval(program, env)
	if evaluatedProg != nil {
		if _, ok := evaluatedProg.(*object.Error); ok {
			_, _ = io.WriteString(output, evaluatedProg.Inspect())
			_, _ = io.WriteString(output, "\n")
		}
	}
	return nil
}
