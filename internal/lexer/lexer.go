package lexer

import (
	"io"
	"strings"
	"unicode"

	"github.com/Abathargh/harlock/internal/token"
)

type Lexer struct {
	input        io.RuneReader
	position     int
	readPosition int
	char         rune
}

func NewLexer(input io.RuneReader) *Lexer {
	l := &Lexer{input: input}
	l.readRune()
	return l
}

func (lexer *Lexer) NextToken() token.Token {
	var t token.Token
	switch lexer.char {
	case '=':
		t = token.Token{Type: token.ASSIGN, Literal: string(lexer.char)}
	case '+':
		t = token.Token{Type: token.PLUS, Literal: string(lexer.char)}
	case ',':
		t = token.Token{Type: token.COMMA, Literal: string(lexer.char)}
	case '\n':
		t = token.Token{Type: token.NEWLINE, Literal: string(lexer.char)}
	case '(':
		t = token.Token{Type: token.LPAREN, Literal: string(lexer.char)}
	case ')':
		t = token.Token{Type: token.RPAREN, Literal: string(lexer.char)}
	case '{':
		t = token.Token{Type: token.LBRACE, Literal: string(lexer.char)}
	case '}':
		t = token.Token{Type: token.RBRACE, Literal: string(lexer.char)}
	case 0:
		t = token.Token{Type: token.EOF, Literal: ""}
	default:
		if unicode.IsLetter(lexer.char) {
			return token.Token{Literal: lexer.readIdentifier()}
		}
	}
	lexer.readRune()
	return t
}

func (lexer *Lexer) readIdentifier() string {
	var buf strings.Builder
	for unicode.IsLetter(lexer.char) {
		buf.WriteRune(lexer.char)
		lexer.readRune()
	}
	return buf.String()
}

func (lexer *Lexer) readRune() {
	if r, _, err := lexer.input.ReadRune(); err != nil {
		lexer.char = 0
	} else {
		lexer.char = r
	}
	lexer.position = lexer.readPosition
	lexer.readPosition++
}
