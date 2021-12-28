package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"

	"github.com/Abathargh/harlock/internal/ast"
	"github.com/Abathargh/harlock/internal/lexer"
)

func TestVarStatements(t *testing.T) {
	input := `var x = 5
var y = 10
var test = 2000
`
	lex := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(input)))
	p := NewParser(lex)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if program == nil {
		t.Fatalf("ParseProgram returned nil")
	}

	if len(program.Statements) != 3 {
		t.Fatalf("Expected 3 statements, got %d", len(program.Statements))
	}

	expectedIdentifiers := []string{"x", "y", "test"}
	for idx, testCase := range expectedIdentifiers {
		if !testLetStatement(t, program.Statements[idx], testCase) {
			return
		}
	}
}

func TestReturnStatement(t *testing.T) {
	input := `ret 5
ret 10
ret 2000
`
	lex := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(input)))
	p := NewParser(lex)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("Expected 3 statements, got %d", len(program.Statements))
	}

	for _, statement := range program.Statements {
		returnStatement, ok := statement.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("Expected the statement to have ReturnStatement type, got %T", statement)
			continue
		}
		if returnStatement.TokenLiteral() != "ret" {
			t.Errorf("Expected token literal to be 'ret', got %s", returnStatement.TokenLiteral())
		}
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := `test`

	lex := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(input)))
	p := NewParser(lex)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statements, got %d", len(program.Statements))
	}

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Errorf("Expected the statement to have ExpressionStatement type, got %T", program.Statements[0])
	}

	expression := statement.Expression.(*ast.Identifier)
	if !ok {
		t.Errorf("Expected the expession to have *Identifier type, got %T", statement.Expression)
	}

	if expression.Value != "test" {
		t.Errorf("Expected expression literal to be \"test\", got %q", expression.Value)
	}

	if expression.TokenLiteral() != "test" {
		t.Errorf("Expected token literal to be \"test\", got %q", expression.TokenLiteral())
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := `15`

	lex := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(input)))
	p := NewParser(lex)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statements, got %d", len(program.Statements))
	}

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Errorf("Expected the statement to have ExpressionStatement type, got %T", program.Statements[0])
	}

	literal := statement.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("Expected the expession to have *Identifier type, got %T", statement.Expression)
	}

	if literal.Value != 15 {
		t.Errorf("Expected expression literal to be \"15\", got %q", literal.Value)
	}

	if literal.TokenLiteral() != "15" {
		t.Errorf("Expected token literal to be \"15\", got %q", literal.TokenLiteral())
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		prefixOp string
		intValue int64
	}{
		{"!10", "!", 10},
		{"-15", "-", 15},
		{"~20", "~", 20},
	}

	for _, testCase := range tests {
		lex := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(testCase.input)))
		p := NewParser(lex)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("Expected 1 statements, got %d", len(program.Statements))
		}

		statement, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("Expected the statement to have ExpressionStatement type, got %T", program.Statements[0])
		}

		prefixExpression, ok := statement.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Errorf("Expected the expression to have PrefixExpression type, got %T", statement.Expression)
		}

		if prefixExpression.Operator != testCase.prefixOp {
			t.Errorf("Expected operator %q got %q", testCase.prefixOp, prefixExpression.Operator)
		}

		if !testIntegerLiteral(t, prefixExpression.RightExpression, testCase.intValue) {
			return
		}
	}
}

func TestParsingInfixOperators(t *testing.T) {
	tests := []struct {
		input        string
		leftOperand  int64
		operator     string
		rightOperand int64
	}{
		{"5 + 5", 5, "+", 5},
		{"5 - 5", 5, "-", 5},
		{"5 * 5", 5, "*", 5},
		{"5 / 5", 5, "/", 5},
		{"5 % 5", 5, "%", 5},
		{"5 < 5", 5, "<", 5},
		{"5 > 5", 5, ">", 5},
		{"5 <= 5", 5, "<=", 5},
		{"5 >= 5", 5, ">=", 5},
		{"5 == 5", 5, "==", 5},
		{"5 != 5", 5, "!=", 5},
		{"5 | 5", 5, "|", 5},
		{"5 & 5", 5, "&", 5},
		{"5 ^ 5", 5, "^", 5},
		{"5 >> 5", 5, ">>", 5},
		{"5 << 5", 5, "<<", 5},
	}

	for _, testCase := range tests {
		lex := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(testCase.input)))
		p := NewParser(lex)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("Expected 1 statements, got %d", len(program.Statements))
		}

		statement, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("Expected the statement to have ExpressionStatement type, got %T", program.Statements[0])
		}

		infixExpression, ok := statement.Expression.(*ast.InfixExpression)
		if !ok {
			t.Errorf("Expected the expression to have InfixExpression type, got %T", statement.Expression)
		}

		if !testIntegerLiteral(t, infixExpression.LeftExpression, testCase.leftOperand) {
			return
		}

		if infixExpression.Operator != testCase.operator {
			t.Fatalf("expected operator %q, got %q", testCase.operator, infixExpression.Operator)
		}

		if !testIntegerLiteral(t, infixExpression.RightExpression, testCase.rightOperand) {
			return
		}
	}
}

func testIntegerLiteral(t *testing.T, rightExpression ast.Expression, integerValue int64) bool {
	integerExprValue, ok := rightExpression.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("Expected expression to have IntegerLiteral type, got %T", rightExpression)
		return false
	}

	if integerExprValue.Value != integerValue {
		t.Errorf("Expected value %q got %q", integerValue, integerExprValue.Value)
		return false
	}

	if integerExprValue.TokenLiteral() != fmt.Sprintf("%d", integerValue) {
		t.Errorf("Expected token literal %q got %q", integerValue, integerExprValue.TokenLiteral())
		return false
	}
	return true
}

func testIdentifier(t *testing.T, expression ast.Expression, value string) bool {
	identifier, ok := expression.(*ast.Identifier)
	if !ok {
		t.Errorf("expected Identifier type, got %T", expression)
		return false
	}

	if identifier.Value != value {
		t.Errorf("expected identifier value %q, got %q", value, identifier.Value)
		return false
	}

	if identifier.TokenLiteral() != value {
		t.Errorf("expected identifier value %q, got %q", value, identifier.TokenLiteral())
		return false
	}
	return true
}

func testLiteralExpression(t *testing.T, expression ast.Expression, expected interface{}) bool {
	switch value := expected.(type) {
	case int:
		return testIntegerLiteral(t, expression, int64(value))
	case int64:
		return testIntegerLiteral(t, expression, value)
	case string:
		return testIdentifier(t, expression, value)
	}
	t.Errorf("unhandled type for the passed expression, got %T", expression)
	return false
}

func testInfixExpression(t *testing.T, expression ast.Expression, left interface{},
	operator string, right interface{}) bool {
	operatorExpression, ok := expression.(*ast.InfixExpression)
	if !ok {
		t.Errorf("expected InfixExpression type, got %T", expression)
		return false
	}

	if !testLiteralExpression(t, operatorExpression.LeftExpression, left) {
		return false
	}

	if operatorExpression.Operator != operator {
		t.Errorf("expected operator %q2, got %q", operator, operatorExpression.Operator)
		return false
	}

	if !testLiteralExpression(t, operatorExpression.RightExpression, right) {
		return false
	}
	return true
}

func testLetStatement(t *testing.T, statement ast.Statement, name string) bool {
	if statement.TokenLiteral() != "var" {
		t.Errorf("Expected var, got %s", statement.TokenLiteral())
		return false
	}

	varStatement, ok := statement.(*ast.VarStatement)
	if !ok {
		t.Errorf("Expected the statement to have VarStatement type, got %T", statement)
		return false
	}

	if varStatement.Name.Value != name {
		t.Errorf("Expected name of the variable to be %s, got %s", name, varStatement.Name.Value)
		return false
	}

	if varStatement.Name.TokenLiteral() != name {
		t.Errorf("Expected token literal to be %s, got %s", name, varStatement.Name.TokenLiteral())
		return false
	}
	return true
}

func checkParserErrors(t *testing.T, parser *Parser) {
	errors := parser.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("got %d parser errors", len(errors))
	for _, errMsg := range errors {
		t.Errorf("parser error: %s", errMsg)
	}
	t.FailNow()
}
