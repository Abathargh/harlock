package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/Abathargh/harlock/pkg/interpreter"

	"github.com/Abathargh/harlock/internal/repl"
)

const (
	Version     = ""
	nameMessage = "usage: harlock [flags] [filename]"
	helpMessage = `Execute an harlock script or start a REPL session. 
If the optional filename argument is passed, it must be a valid 
name for an existing file, the contents of which will be executed. 
If no file is passed, the interpreter starts in interactive-mode.

Flags:`
)

func main() {
	fs := flag.NewFlagSet("harlock", flag.ExitOnError)
	help := fs.Bool("help", false, "show this help message")
	version := fs.Bool("version", false, "prints the version for this build")
	err := fs.Parse(os.Args[1:])
	if err != nil {
		panic(err)
	}

	if *help {
		fmt.Printf("%s\n", nameMessage)
		fmt.Printf("%s\n", helpMessage)
		fs.PrintDefaults()
		return
	}

	if *version {
		fmt.Printf("Harlock %s\n", Version)
		return
	}

	switch len(os.Args) {
	case 1:
		fmt.Printf("Hsarlock %s - %s on %s\n", Version, runtime.GOARCH, runtime.GOOS)
		repl.Start(os.Stdin, os.Stdout)
	case 2:
		f, err := os.Open(os.Args[1])
		if err != nil {
			_, _ = io.WriteString(os.Stderr, err.Error()+"\n")
		}

		errs := interpreter.Exec(f, os.Stdout)
		if errs != nil {
			for _, err := range errs {
				_, _ = io.WriteString(os.Stderr, fmt.Sprintf("%s\n", err))
			}
		}
	}
}
