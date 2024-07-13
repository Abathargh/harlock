// Package interpreter implements the public API that
// can be used to interact with the harlock runtime
// from other applications or the CLI tool.
package interpreter

import (
	"bufio"
	"fmt"
	"io"
	"runtime/debug"

	"github.com/Abathargh/harlock/internal/object"

	"github.com/Abathargh/harlock/internal/evaluator"
	"github.com/Abathargh/harlock/internal/lexer"
	"github.com/Abathargh/harlock/internal/parser"
)

// Version represents the current harlock version
var Version = ""

func init() {
	if Version == "" {
		info, ok := debug.ReadBuildInfo()
		if ok {
			Version = info.Main.Version
		}
	}
}

// Exec reads a script from the passed reader, executes it and
// sends the generated output to the passed writer. If the parsing
// phase fails, it returns an array of string containing the parsing
// errors, or nil otherwise.
func Exec(r io.Reader, stderr io.Writer, args ...string) []string {
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
		switch evaluatedProg.(type) {
		case *object.RuntimeError:
			return dumpToSlice(evaluatedProg)
		case *object.Error:
			return dumpToSlice(evaluatedProg)
		}
	}
	return nil
}

func dumpToSlice(evaluatedProg object.Object) []string {
	return []string{
		fmt.Sprintf("%s\n", evaluatedProg.Inspect()),
	}
}
