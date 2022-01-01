package main

import (
	"fmt"
	"io"
	"os"

	"github.com/Abathargh/harlock/pkg/interpreter"

	"github.com/Abathargh/harlock/internal/repl"
)

func main() {
	switch len(os.Args) {
	case 1:
		fmt.Println("Harlock programming language REPL")
		repl.Start(os.Stdin, os.Stdout)
	case 2:
		errs := interpreter.Exec(os.Args[1], os.Stdout)
		if errs != nil {
			for _, err := range errs {
				_, _ = io.WriteString(os.Stderr, err)
			}
		}
	}
}
