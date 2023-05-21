package interpreter

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const (
	headerTemplate = `package main
import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/Abathargh/harlock/pkg/interpreter"
)

func main() {
	fileContent := `

	footerTemplate = `
errs := interpreter.Exec(bytes.NewBufferString(fileContent), os.Stdout, os.Args...)
	if errs != nil {
		for _, err := range errs {
			_, _ = io.WriteString(os.Stderr, fmt.Sprintf("%s\n", err))
		}
	}
}`
)

// Embed generates an executable from a script, by embedding the script
// and the harlock runtime within a binary, returning an error if the
// process fails.
func Embed(filename string) error {
	program, err := buildEmbeddedProgram(filename)
	if err != nil {
		return embedError(err)
	}
	_ = os.Mkdir("./temp", 0775)
	_ = os.WriteFile("./temp/main.go", []byte(program), 0775)
	defer func() { _ = os.RemoveAll("./temp") }()

	modCmd := command("go", "mod", "init", "embedded_harlock")
	if err := modCmd.Run(); err != nil {
		return embedError(err)
	}

	tidyCmd := command("go", "mod", "tidy")
	if err := tidyCmd.Run(); err != nil {
		return embedError(err)
	}

	buildCmd := command("go", "build", "-ldflags", "-s", "-ldflags", "-w")
	if err := buildCmd.Run(); err != nil {
		return embedError(err)
	}

	tmpName := "./temp/embedded_harlock"
	execName := "./" + strings.Split(filename, ".")[0]
	if runtime.GOOS == "windows" {
		tmpName += ".exe"
		execName += ".exe"
	}

	if err := moveFile(tmpName, execName); err != nil {
		return embedError(err)
	}
	return nil
}

func buildEmbeddedProgram(filename string) (string, error) {
	fileContents, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return headerTemplate + "`" + string(fileContents) + "`" + footerTemplate, nil
}

func embedError(err error) error {
	msg := err.Error()
	if strings.HasPrefix(msg, "exec: ") {
		err = errors.New(msg[len("exec: "):])
	}
	return fmt.Errorf("embed error: could not generate an harlock binary (%w)", err)
}

func command(c string, args ...string) *exec.Cmd {
	cmd := exec.Command(c, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = "./temp"
	return cmd
}

func moveFile(inName string, outName string) error {
	if err := copyFile(inName, outName); err != nil {
		return err
	}

	if err := os.Remove(inName); err != nil {
		return err
	}
	return nil
}

func copyFile(inName string, outName string) error {
	dest, err := os.OpenFile(outName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer func() { _ = dest.Close() }()

	src, err := os.Open(inName)
	if err != nil {
		return err
	}

	if _, err = io.Copy(dest, src); err != nil {
		_ = os.Remove(outName)
		return err
	}
	_ = src.Close()
	return nil
}
