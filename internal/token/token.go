package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT = "IDENT"
	INT   = "INT"
	STR   = "STRING"

	ASSIGN  = "="
	PLUS    = "+"
	MINUS   = "-"
	MUL     = "*"
	DIV     = "/"
	MOD     = "%"
	LESS    = "<"
	GREATER = ">"

	NOT = "!"
	OR  = "|"
	XOR = "^"
	AND = "&"
	REV = "~"

	EQUALS    = "=="
	NOTEQUALS = "!="
	LESSEQ    = "<="
	GREATEREQ = ">="

	LSHIFT = "<<"
	RSHIFT = ">>"

	LOGICAND = "&&"
	LOGICOR  = "||"

	COMMA   = ","
	ESCAPE  = "\\"
	NEWLINE = "\n"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"

	FUNCTION = "FUNCTION"
	VAR      = "VAR"
	TRY      = "TRY"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RET      = "RET"
)

var keywords = map[string]TokenType{
	"fun":   FUNCTION,
	"var":   VAR,
	"try":   TRY,
	"true":  TRUE,
	"false": FALSE,
	"if":    IF,
	"else":  ELSE,
	"ret":   RET,
}

func LookupIdentifier(identifier string) TokenType {
	if val, ok := keywords[identifier]; ok {
		return val
	}
	return IDENT
}
