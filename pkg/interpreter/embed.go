package interpreter

import (
	"io"
	"io/ioutil"
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
		return err
	}
	_ = os.Mkdir("tmp", 666)
	_ = os.WriteFile("tmp/main.go", []byte(program), 666)
	defer func() { _ = os.RemoveAll("./tmp") }()

	modCmd := command("go", "mod", "init", "embedded_harlock")
	if err := modCmd.Run(); err != nil {
		return err
	}

	tidyCmd := command("go", "mod", "tidy")
	if err := tidyCmd.Run(); err != nil {
		return err
	}

	buildCmd := command("go", "build", "-ldflags", "-s", "-ldflags", "-w")
	if err := buildCmd.Run(); err != nil {
		return err
	}

	tmpName := "./tmp/embedded_harlock"
	execName := "./" + strings.Split(filename, ".")[0]
	if runtime.GOOS == "windows" {
		tmpName += ".exe"
		execName += ".exe"
	}

	if err := moveFile(tmpName, execName); err != nil {
		return err
	}
	return nil
}

func buildEmbeddedProgram(filename string) (string, error) {
	fileContents, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return headerTemplate + "`" + string(fileContents) + "`" + footerTemplate, nil
}

func command(c string, args ...string) *exec.Cmd {
	cmd := exec.Command(c, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = "./tmp"
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
