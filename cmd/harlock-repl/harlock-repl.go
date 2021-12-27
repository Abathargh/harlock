package main

import (
	"fmt"
	"os"

	"github.com/Abathargh/harlock/internal/repl"
)

func main() {
	fmt.Println("Harlock programming language REPL")
	repl.Start(os.Stdin, os.Stdout)
}
