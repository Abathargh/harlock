package lexer

const (
	invalidHex    = LexError("invalid hex escape, expected \\xXX, where X is an hex digit (0-9 a-f)")
	invalidUni    = LexError("invalid unicode escape, expected \\xUUUU, where U is an hex digit (0-9 a-f)")
	invalidEsc    = LexError("invalid escape")
	invalidString = LexError("quote delimiter not found at the end of the string")
)

type LexError string

func (le LexError) Error() string {
	return string(le)
}
