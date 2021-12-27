package repl

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/Abathargh/harlock/internal/lexer"
	"github.com/Abathargh/harlock/internal/token"
)

const PROMPT = ">>"

func Start(input io.Reader, output io.Writer) {
	scanner := bufio.NewScanner(input)

	for {
		_, _ = fmt.Fprintf(output, PROMPT)
		if !scanner.Scan() {
			return
		}

		line := scanner.Text()
		l := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(line)))
		for t := l.NextToken(); t.Type != token.EOF; t = l.NextToken() {
			_, _ = fmt.Fprintf(output, "%+v\n", t)
		}
	}
}
