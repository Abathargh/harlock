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
	embed := fs.String("embed", "", "embeds the input script into an executable "+
		"containing the interpreter runtime")

	if err := fs.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	switch {
	case *help:
		fmt.Printf("%s\n", nameMessage)
		fmt.Printf("%s\n", helpMessage)
		fs.PrintDefaults()
		return
	case *version:
		fmt.Printf("Harlock %s\n", interpreter.Version)
		return
	case *embed != "":
		if err := interpreter.Embed(*embed); err != nil {
			_, _ = io.WriteString(os.Stderr, err.Error()+"\n")
			return
		}
	case len(fs.Args()) == 0:
		fmt.Printf("Harlock %s - %s on %s\n", interpreter.Version, runtime.GOARCH, runtime.GOOS)
		repl.Start(os.Stdin, os.Stdout)
	case len(fs.Args()) > 0:
		f, err := os.Open(fs.Arg(0))
		if err != nil {
			_, _ = io.WriteString(os.Stderr, err.Error()+"\n")
		}

		errs := interpreter.Exec(f, os.Stdout, fs.Args()...)
		if errs != nil {
			for _, err := range errs {
				_, _ = io.WriteString(os.Stderr, fmt.Sprintf("%s\n", err))
			}
		}
	}
}
