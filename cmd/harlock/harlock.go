package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/Abathargh/harlock/internal/repl"
	"github.com/Abathargh/harlock/pkg/interpreter"
)

const (
	nameMessage = "usage: harlock [flags] [filename] [args]"
	helpMessage = `Execute an harlock script or start a REPL session. 
If the optional filename argument is passed, it must be a valid 
name for an existing file, the contents of which will be executed. 
If a filename is passed, a number of additional args can be passed, 
that will be available within the instance of the execution.
If no file is passed, the interpreter starts in interactive-mode.

Flags:`

	helpUsage    = "show this help message"
	versionUsage = "prints the version for this build"
	embedUsage   = "embeds the input script into an executable " +
		"containing the interpreter runtime"
)

func main() {
	fs := flag.NewFlagSet("harlock", flag.ExitOnError)
	help := fs.Bool("help", false, helpUsage)
	version := fs.Bool("version", false, versionUsage)
	embed := fs.String("embed", "", embedUsage)

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

		errs := interpreter.Exec(f, os.Stderr, fs.Args()...)
		if errs != nil {
			for _, err := range errs {
				_, _ = io.WriteString(os.Stderr, fmt.Sprintf("%s\n", err))
			}
		}
	}
}
