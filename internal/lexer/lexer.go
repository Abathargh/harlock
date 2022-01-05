package lexer

import (
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/Abathargh/harlock/internal/token"
)

type Lexer struct {
	input     io.RuneReader
	position  int
	wasPeeked bool
	peeked    rune
	char      rune
}

func NewLexer(input io.RuneReader) *Lexer {
	l := &Lexer{input: input, position: -1}
	l.readRune()
	return l
}

func (lexer *Lexer) NextToken() token.Token {
	var t token.Token
	lexer.skipWhitespace()

	switch lexer.char {
	case '=':
		if lexer.peekRune() == '=' {
			t = token.Token{Type: token.EQUALS, Literal: lexer.buildTwoRuneOperator()}
		} else {
			t = token.Token{Type: token.ASSIGN, Literal: string(lexer.char)}
		}
	case '\'':
		fallthrough
	case '"':
		// TODO customize token for error e.g. token.NODELIMITER and catch
		// TODO it in the parser with custom err msg
		str, err := lexer.readString()
		if err != nil {
			return token.Token{Type: token.ILLEGAL, Literal: string(lexer.char)}
		}
		t = token.Token{Type: token.STR, Literal: str}
	case '+':
		t = token.Token{Type: token.PLUS, Literal: string(lexer.char)}
	case '-':
		t = token.Token{Type: token.MINUS, Literal: string(lexer.char)}
	case '*':
		t = token.Token{Type: token.MUL, Literal: string(lexer.char)}
	case '/':
		peekedRune := lexer.peekRune()
		if peekedRune == '/' {
			lexer.skipComment()
		}
		t = token.Token{Type: token.NEWLINE, Literal: "\n"}
	case '%':
		t = token.Token{Type: token.MOD, Literal: string(lexer.char)}
	case '<':
		peekedRune := lexer.peekRune()
		if peekedRune == '=' {
			t = token.Token{Type: token.LESSEQ, Literal: lexer.buildTwoRuneOperator()}
		} else if peekedRune == '<' {
			t = token.Token{Type: token.LSHIFT, Literal: lexer.buildTwoRuneOperator()}
		} else {
			t = token.Token{Type: token.LESS, Literal: string(lexer.char)}
		}
	case '>':
		peekedRune := lexer.peekRune()
		if peekedRune == '=' {
			t = token.Token{Type: token.GREATEREQ, Literal: lexer.buildTwoRuneOperator()}
		} else if peekedRune == '>' {
			t = token.Token{Type: token.RSHIFT, Literal: lexer.buildTwoRuneOperator()}
		} else {
			t = token.Token{Type: token.GREATER, Literal: string(lexer.char)}
		}
	case '!':
		if lexer.peekRune() == '=' {
			t = token.Token{Type: token.NOTEQUALS, Literal: lexer.buildTwoRuneOperator()}
		} else {
			t = token.Token{Type: token.NOT, Literal: string(lexer.char)}
		}
	case '|':
		if lexer.peekRune() == '|' {
			t = token.Token{Type: token.LOGICOR, Literal: lexer.buildTwoRuneOperator()}
		} else {
			t = token.Token{Type: token.OR, Literal: string(lexer.char)}
		}
	case '^':
		t = token.Token{Type: token.XOR, Literal: string(lexer.char)}
	case '&':
		if lexer.peekRune() == '&' {
			t = token.Token{Type: token.LOGICAND, Literal: lexer.buildTwoRuneOperator()}
		} else {
			t = token.Token{Type: token.AND, Literal: string(lexer.char)}
		}
	case '~':
		t = token.Token{Type: token.REV, Literal: string(lexer.char)}
	case ',':
		t = token.Token{Type: token.COMMA, Literal: string(lexer.char)}
	case ':':
		t = token.Token{Type: token.COLON, Literal: string(lexer.char)}
	case '\n':
		t = token.Token{Type: token.NEWLINE, Literal: string(lexer.char)}
	case '(':
		t = token.Token{Type: token.LPAREN, Literal: string(lexer.char)}
	case ')':
		t = token.Token{Type: token.RPAREN, Literal: string(lexer.char)}
	case '[':
		t = token.Token{Type: token.LBRACK, Literal: string(lexer.char)}
	case ']':
		t = token.Token{Type: token.RBRACK, Literal: string(lexer.char)}
	case '{':
		t = token.Token{Type: token.LBRACE, Literal: string(lexer.char)}
	case '}':
		t = token.Token{Type: token.RBRACE, Literal: string(lexer.char)}
	case 0:
		t = token.Token{Type: token.EOF, Literal: ""}
	default:
		if unicode.IsLetter(lexer.char) || lexer.char == '_' {
			id := lexer.readIdentifier()
			return token.Token{Type: token.LookupIdentifier(id), Literal: id}
		}
		if isDigit(lexer.char) {
			peek := lexer.peekRune()
			if lexer.char == '0' && (peek == 'x' || peek == 'X') {
				return token.Token{Type: token.INT, Literal: lexer.readHexNumber()}
			}
			return token.Token{Type: token.INT, Literal: lexer.readNumber()}
		}
		t = token.Token{Type: token.ILLEGAL, Literal: string(lexer.char)}
	}
	lexer.readRune()
	return t
}

func (lexer *Lexer) readIdentifier() string {
	var buf strings.Builder
	for unicode.IsLetter(lexer.char) || unicode.IsDigit(lexer.char) || lexer.char == '_' {
		buf.WriteRune(lexer.char)
		lexer.readRune()
	}
	return buf.String()
}

func (lexer *Lexer) readNumber() string {
	var buf strings.Builder
	for isDigit(lexer.char) {
		buf.WriteRune(lexer.char)
		lexer.readRune()
	}
	return buf.String()
}

func (lexer *Lexer) readHexNumber() string {
	var buf strings.Builder

	// read the 0x that we know is present
	buf.WriteRune(lexer.char)
	lexer.readRune()
	buf.WriteRune(lexer.char)
	lexer.readRune()

	for isHexDigit(lexer.char) {
		buf.WriteRune(lexer.char)
		lexer.readRune()
	}
	return buf.String()
}

func (lexer *Lexer) readRune() {
	lexer.position++
	if lexer.wasPeeked {
		lexer.wasPeeked = false
		lexer.char = lexer.peeked
		return
	}

	if r, _, err := lexer.input.ReadRune(); err == nil {
		lexer.char = r
		return
	}
	lexer.char = 0
}

func (lexer *Lexer) readString() (string, error) {
	// TODO add char escaping
	var buf strings.Builder
	quoteType := lexer.char
	lexer.readRune()
	for ; lexer.char != quoteType && lexer.char != 0; lexer.readRune() {
		buf.WriteRune(lexer.char)
	}
	if lexer.char == 0 {
		return "", fmt.Errorf("quote delimiter not found")
	}
	return buf.String(), nil
}

func (lexer *Lexer) peekRune() rune {
	if r, _, err := lexer.input.ReadRune(); err == nil {
		lexer.peeked = r
		lexer.wasPeeked = true
		return lexer.peeked
	}
	return 0
}

func (lexer *Lexer) skipWhitespace() {
	for lexer.char == ' ' || lexer.char == '\t' || lexer.char == '\r' {
		lexer.readRune()
	}
}

func (lexer *Lexer) skipComment() {
	for lexer.char != '\n' && lexer.char != 0 {
		lexer.readRune()
	}
}

func (lexer *Lexer) buildTwoRuneOperator() string {
	var buf [2]rune
	buf[0] = lexer.char
	lexer.readRune()
	buf[1] = lexer.char
	return string(buf[:])
}

func isDigit(r rune) bool {
	return unicode.IsDigit(r)
}

func isHexDigit(r rune) bool {
	return isDigit(r) || ('a' <= r && r <= 'f') || ('A' <= r && r <= 'F')
}
