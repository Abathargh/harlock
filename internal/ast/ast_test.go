package ast

import (
	"testing"

	"github.com/Abathargh/harlock/internal/token"
)

func TestString(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&VarStatement{
				Token: token.Token{Type: token.VAR, Literal: "var"},
				Name: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "x"},
					Value: "x",
				},
				Value: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "y"},
					Value: "y",
				},
			},
		},
	}

	if program.String() != "var x = y" {
		t.Errorf("expected 'var x = y', got %q", program.String())
	}
}
