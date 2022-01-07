package lexer

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/Abathargh/harlock/internal/token"
)

func TestNextToken(t *testing.T) {
	input := `var test = 12
var test2 = 24
fun f(a, b) {
	var c = try div(a, b)
}
!|&^~-/*<>
if ret false true else
!= == <= >= % >> << && || 0xFF
"long string with text"
'string with single quote'
[1, 2, "ciao"]
{"key": "val"}
file.test()
"\x55\X5a"
"\u0056\u005b"`
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.VAR, "var"},
		{token.IDENT, "test"},
		{token.ASSIGN, "="},
		{token.INT, "12"},
		{token.NEWLINE, "\n"},
		{token.VAR, "var"},
		{token.IDENT, "test2"},
		{token.ASSIGN, "="},
		{token.INT, "24"},
		{token.NEWLINE, "\n"},

		{token.FUNCTION, "fun"},
		{token.IDENT, "f"},
		{token.LPAREN, "("},
		{token.IDENT, "a"},
		{token.COMMA, ","},
		{token.IDENT, "b"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.NEWLINE, "\n"},

		{token.VAR, "var"},
		{token.IDENT, "c"},
		{token.ASSIGN, "="},
		{token.TRY, "try"},
		{token.IDENT, "div"},
		{token.LPAREN, "("},
		{token.IDENT, "a"},
		{token.COMMA, ","},
		{token.IDENT, "b"},
		{token.RPAREN, ")"},
		{token.NEWLINE, "\n"},
		{token.RBRACE, "}"},
		{token.NEWLINE, "\n"},

		{token.NOT, "!"},
		{token.OR, "|"},
		{token.AND, "&"},
		{token.XOR, "^"},
		{token.REV, "~"},
		{token.MINUS, "-"},
		{token.DIV, "/"},
		{token.MUL, "*"},
		{token.LESS, "<"},
		{token.GREATER, ">"},
		{token.NEWLINE, "\n"},

		{token.IF, "if"},
		{token.RET, "ret"},
		{token.FALSE, "false"},
		{token.TRUE, "true"},
		{token.ELSE, "else"},
		{token.NEWLINE, "\n"},

		{token.NOTEQUALS, "!="},
		{token.EQUALS, "=="},
		{token.LESSEQ, "<="},
		{token.GREATEREQ, ">="},
		{token.MOD, "%"},
		{token.RSHIFT, ">>"},
		{token.LSHIFT, "<<"},
		{token.LOGICAND, "&&"},
		{token.LOGICOR, "||"},
		{token.INT, "0xFF"},
		{token.NEWLINE, "\n"},

		{token.STR, "long string with text"},
		{token.NEWLINE, "\n"},
		{token.STR, "string with single quote"},
		{token.NEWLINE, "\n"},

		{token.LBRACK, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.COMMA, ","},
		{token.STR, "ciao"},
		{token.RBRACK, "]"},
		{token.NEWLINE, "\n"},

		{token.LBRACE, "{"},
		{token.STR, "key"},
		{token.COLON, ":"},
		{token.STR, "val"},
		{token.RBRACE, "}"},
		{token.NEWLINE, "\n"},

		{token.IDENT, "file"},
		{token.PERIOD, "."},
		{token.IDENT, "test"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.NEWLINE, "\n"},

		{token.STR, "UZ"},
		{token.NEWLINE, "\n"},

		{token.STR, "V["},
		{token.EOF, ""},
	}

	lexer := NewLexer(bufio.NewReader(bytes.NewBufferString(input)))

	for idx, testCase := range tests {
		tok := lexer.NextToken()
		if tok.Type != testCase.expectedType {
			t.Fatalf("Expected %q, got %q for token #%d", testCase.expectedType, tok.Type, idx)
		}

		if tok.Literal != testCase.expectedLiteral {
			t.Fatalf("Expected %q, got %q for token #%d", testCase.expectedLiteral, tok.Literal, idx)
		}
	}
}
