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
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      any
	}{
		{"var x = 5", "x", 5},
		{"var test = true", "test", true},
		{"var test2 = y", "test2", "y"},
		{"var test2 = y", "test2", "y"},
	}
	for _, testCase := range tests {
		lex := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(testCase.input)))
		p := NewParser(lex)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("Expected 1 statements, got %d", len(program.Statements))
		}

		statement := program.Statements[0]
		if !testVarStatement(t, statement, testCase.expectedIdentifier) {
			return
		}

		value := statement.(*ast.VarStatement).Value
		if !testLiteralExpression(t, value, testCase.expectedValue) {
			return
		}
	}
}

func TestReturnStatement(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue any
	}{
		{"ret 5\n", 5},
		{"ret true\n", true},
		{"ret false\n", false},
		{"ret a\n", "a"},
	}

	for _, testCase := range tests {

		lex := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(testCase.input)))
		p := NewParser(lex)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("Expected 1 statements, got %d", len(program.Statements))
		}

		returnStatement, ok := program.Statements[0].(*ast.ReturnStatement)
		if !ok {
			t.Errorf("Expected the statement to have ReturnStatement type, got %T", program.Statements[0])
			continue
		}

		if returnStatement.TokenLiteral() != "ret" {
			t.Errorf("Expected token literal to be 'ret', got %s", returnStatement.TokenLiteral())
		}

		value := returnStatement.ReturnValue
		if !testLiteralExpression(t, value, testCase.expectedValue) {
			return
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
		t.Errorf("Expected the expression to have *Identifier type, got %T", statement.Expression)
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
		t.Errorf("Expected the expression to have *IntegerLiteral type, got %T", statement.Expression)
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
		Value    any
	}{
		{"!10", "!", 10},
		{"-15", "-", 15},
		{"~20", "~", 20},
		{"!true", "!", true},
		{"!false", "!", false},
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

		if !testLiteralExpression(t, prefixExpression.RightExpression, testCase.Value) {
			return
		}
	}
}

func TestParsingInfixOperators(t *testing.T) {
	tests := []struct {
		input        string
		leftOperand  any
		operator     string
		rightOperand any
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
		{"true && false", true, "&&", false},
		{"false || true", false, "||", true},
		{"true == true", true, "==", true},
		{"false == false", false, "==", false},
		{"true != false", true, "!=", false},
		{"false != true", false, "!=", true},
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

		if !testInfixExpression(t, statement.Expression, testCase.leftOperand, testCase.operator, testCase.rightOperand) {
			return
		}
	}
}

func TestBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
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

		literal := statement.Expression.(*ast.Boolean)
		if !ok {
			t.Errorf("Expected the expression to have *Boolean type, got %T", statement.Expression)
		}

		if literal.Value != testCase.expected {
			t.Errorf("Expected expression literal to be %t, got %t", testCase.expected, literal.Value)
		}

		if literal.TokenLiteral() != fmt.Sprintf("%t", testCase.expected) {
			t.Errorf("Expected token literal to be %t, got %q", testCase.expected, literal.TokenLiteral())
		}
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"true", "true"},
		{"3 <= 5 == true", "((3<=5)==true)"},
		{"-a + b", "((-a)+b)"},
		{"-a * b + c", "(((-a)*b)+c)"},
		{"a | b & c |d", "((a|(b&c))|d)"},
		{"(a | b) & (c | d)", "((a|b)&(c|d))"},
		{"(a | b) & (c | d) * add(b | c)", "((a|b)&((c|d)*add((b|c))))"},
		{"a | b && c | d ", "((a|b)&&(c|d))"},
		{"a * [1,2,5][2*1] / 2 ", "((a*[1, 2, 5][(2*1)])/2)"},
		{"call(2 * a[2], 3 + a[3])", "call((2*a[2]), (3+a[3]))"},
		{"2 * test.method()", "(2*test.method())"},
	}

	for _, testCase := range tests {
		lex := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(testCase.input)))
		p := NewParser(lex)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		expression := program.String()
		if expression != testCase.expected {
			t.Errorf("expected %q expression, got %q", testCase.expected, expression)
		}
	}
}

func TestIfExpression(t *testing.T) {
	input := `if x <= y { x }`
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

	expression := statement.Expression.(*ast.IfExpression)
	if !ok {
		t.Errorf("Expected the expression to have *IfExpression type, got %T", statement.Expression)
	}

	if !testInfixExpression(t, expression.Condition, "x", "<=", "y") {
		return
	}

	if len(expression.Consequence.Statements) != 1 {
		t.Errorf("Expected 1 consequence statement got %d", len(expression.Consequence.Statements))
	}

	consequence, ok := expression.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Errorf("Expected the consequence to have *ExpressionStatement type, got %T", expression.Consequence.Statements[0])
	}

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if expression.Alternative != nil {
		t.Errorf("Expected no alternative statement, got one")
	}
}

func TestIfElseExpression(t *testing.T) {
	input := `if (x <= y) { z } else { w }`
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

	expression := statement.Expression.(*ast.IfExpression)
	if !ok {
		t.Errorf("Expected the expression to have *IfExpression type, got %T", statement.Expression)
	}

	if !testInfixExpression(t, expression.Condition, "x", "<=", "y") {
		return
	}

	if len(expression.Consequence.Statements) != 1 {
		t.Errorf("Expected 1 consequence statement got %d", len(expression.Consequence.Statements))
	}

	consequence, ok := expression.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Errorf("Expected the consequence to have *ExpressionStatement type, got %T", expression.Consequence.Statements[0])
	}

	if !testIdentifier(t, consequence.Expression, "z") {
		return
	}

	if expression.Alternative == nil {
		t.Errorf("Expected alternative statement")
	}

	alternative, ok := expression.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Errorf("Expected the alternative consequence to have *ExpressionStatement type, got %T", expression.Consequence.Statements[0])
	}

	if !testIdentifier(t, alternative.Expression, "w") {
		return
	}
}

func TestFunctionLiteral(t *testing.T) {
	input := `fun(a, b, c) {a + b}`
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

	functionLiteral := statement.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Errorf("Expected the expression to have *FunctionLiteral type, got %T", statement.Expression)
	}

	if len(functionLiteral.Parameters) != 3 {
		t.Fatalf("expected 3 function literal parameters, got %d",
			len(functionLiteral.Parameters))
	}

	testLiteralExpression(t, functionLiteral.Parameters[0], "a")
	testLiteralExpression(t, functionLiteral.Parameters[1], "b")
	testLiteralExpression(t, functionLiteral.Parameters[2], "c")

	if len(functionLiteral.Body.Statements) != 1 {
		t.Errorf("epected 1 statement in the function body, got %d", len(functionLiteral.Body.Statements))
	}

	bodyStatement, ok := functionLiteral.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Errorf("Expected the statement to have ExpressionStatement type, got %T", functionLiteral.Body.Statements[0])
	}

	testInfixExpression(t, bodyStatement.Expression, "a", "+", "b")
}

func TestFunctionParametersParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"fun () {}", []string{}},
		{"fun (a) {}", []string{"a"}},
		{"fun (a, b, c) {}", []string{"a", "b", "c"}},
	}

	for _, testCase := range tests {
		lex := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(testCase.input)))
		p := NewParser(lex)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		statement := program.Statements[0].(*ast.ExpressionStatement)
		functionLiteral := statement.Expression.(*ast.FunctionLiteral)
		if len(functionLiteral.Parameters) != len(testCase.expected) {
			t.Fatalf("expected %d parameters, got %d",
				len(testCase.expected), len(functionLiteral.Parameters))
		}

		for idx, identifier := range testCase.expected {
			testLiteralExpression(t, functionLiteral.Parameters[idx], identifier)
		}
	}
}

func TestCallExpressionParsing(t *testing.T) {
	input := "test(a, a | e, b * c, c % f)"
	lex := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(input)))
	p := NewParser(lex)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statements, got %d", len(program.Statements))
	}
	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Errorf("Expected the statement to have ExpressionStatement type, got %T",
			program.Statements[0])
	}

	callExpression, ok := statement.Expression.(*ast.CallExpression)
	if !ok {
		t.Errorf("Expected the expression to have *CallExpression type, got %T",
			statement.Expression)
	}

	if !testIdentifier(t, callExpression.Function, "test") {
		return
	}

	if len(callExpression.Arguments) != 4 {
		t.Errorf("expected 4 arguments, got %d", len(callExpression.Arguments))
	}

	testLiteralExpression(t, callExpression.Arguments[0], "a")
	testInfixExpression(t, callExpression.Arguments[1], "a", "|", "e")
	testInfixExpression(t, callExpression.Arguments[2], "b", "*", "c")
	testInfixExpression(t, callExpression.Arguments[3], "c", "%", "f")

}

func TestStringLiteralExpression(t *testing.T) {
	input := `"test string hello world test"
`
	expected := "test string hello world test"

	lex := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(input)))
	p := NewParser(lex)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement := program.Statements[0].(*ast.ExpressionStatement)
	stringLiteral, ok := statement.Expression.(*ast.StringLiteral)
	if !ok {
		t.Errorf("Expected the statement to have StringLiteral type, got %T", statement.Expression)
	}

	if stringLiteral.Value != expected {
		t.Errorf("expected %s, got %s", expected, stringLiteral.Value)
	}
}
func TestArrayLiteralExpression(t *testing.T) {
	input := `[2, 4 % 5, 4 | 2]`

	lex := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(input)))
	p := NewParser(lex)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement := program.Statements[0].(*ast.ExpressionStatement)
	arrayLiteral, ok := statement.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Errorf("Expected the statement to have ArrayLiteral type, got %T", statement.Expression)
	}

	if len(arrayLiteral.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arrayLiteral.Elements))
	}

	testIntegerLiteral(t, arrayLiteral.Elements[0], 2)
	testInfixExpression(t, arrayLiteral.Elements[1], 4, "%", 5)
	testInfixExpression(t, arrayLiteral.Elements[2], 4, "|", 2)
}

func TestIndexExpression(t *testing.T) {
	input := `arr[4 % 5]`

	lex := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(input)))
	p := NewParser(lex)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement := program.Statements[0].(*ast.ExpressionStatement)
	indexExpression, ok := statement.Expression.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("Expected the statement to have IndexExpression type, got %T", statement.Expression)
	}

	if !testIdentifier(t, indexExpression.Left, "arr") {
		return
	}

	if !testInfixExpression(t, indexExpression.Index, 4, "%", 5) {
		return
	}
}

func TestMapLiteralParsing(t *testing.T) {
	input := `{"test": 6, "tests": 7}`
	expected := map[string]int64{
		"test":  6,
		"tests": 7,
	}

	lex := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(input)))
	p := NewParser(lex)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement := program.Statements[0].(*ast.ExpressionStatement)
	mapLiteral, ok := statement.Expression.(*ast.MapLiteral)
	if !ok {
		t.Fatalf("Expected the statement to have MapLiteral type, got %T", statement.Expression)
	}

	if len(mapLiteral.Mappings) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(mapLiteral.Mappings))
	}

	for key, val := range mapLiteral.Mappings {
		strKey, ok := key.(*ast.StringLiteral)
		if !ok {
			t.Fatalf("Expected key to have string type, got %T", key)
		}
		testIntegerLiteral(t, val, expected[strKey.Value])
	}
}

func TestEmptyMapLiteralParsing(t *testing.T) {
	input := `{}`

	lex := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(input)))
	p := NewParser(lex)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement := program.Statements[0].(*ast.ExpressionStatement)
	mapLiteral, ok := statement.Expression.(*ast.MapLiteral)
	if !ok {
		t.Fatalf("Expected the statement to have MapLiteral type, got %T", statement.Expression)
	}

	if len(mapLiteral.Mappings) != 0 {
		t.Fatalf("expected 0 elements, got %d", len(mapLiteral.Mappings))
	}
}

func TestMapLiteralParsingWithExpressions(t *testing.T) {
	input := `{"first":  1 + 2, "second": 3 & 4}`
	expectedTests := map[string]func(expression ast.Expression){
		"first": func(expression ast.Expression) {
			testInfixExpression(t, expression, 1, "+", 2)
		},
		"second": func(expression ast.Expression) {
			testInfixExpression(t, expression, 3, "&", 4)
		},
	}

	lex := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(input)))
	p := NewParser(lex)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement := program.Statements[0].(*ast.ExpressionStatement)
	mapLiteral, ok := statement.Expression.(*ast.MapLiteral)
	if !ok {
		t.Fatalf("Expected the statement to have MapLiteral type, got %T", statement.Expression)
	}

	if len(mapLiteral.Mappings) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(mapLiteral.Mappings))
	}

	for key, val := range mapLiteral.Mappings {
		strKey, ok := key.(*ast.StringLiteral)
		if !ok {
			t.Fatalf("expected string key, got %T", key)
		}
		testFunction, ok := expectedTests[strKey.String()]
		if !ok {
			t.Fatalf("expected function for key %s, not found", strKey.String())
		}
		testFunction(val)
	}
}

func TestMethodCall(t *testing.T) {
	input := "test.method(1, 2 * 2, 3 - 1)"

	lex := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(input)))
	p := NewParser(lex)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement := program.Statements[0].(*ast.ExpressionStatement)
	methodLiteral, ok := statement.Expression.(*ast.MethodCallExpression)
	if !ok {
		t.Fatalf("Expected the statement to have MethodLiteral type, got %T", statement.Expression)
	}

	if !testIdentifier(t, methodLiteral.Caller, "test") {
		return
	}

	if len(methodLiteral.Called.Arguments) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(methodLiteral.Called.Arguments))
	}

	testIntegerLiteral(t, methodLiteral.Called.Arguments[0], 1)
	testInfixExpression(t, methodLiteral.Called.Arguments[1], 2, "*", 2)
	testInfixExpression(t, methodLiteral.Called.Arguments[2], 3, "-", 1)
}

func TestTryExpression(t *testing.T) {
	input := "try test.method()"

	lex := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(input)))
	p := NewParser(lex)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement := program.Statements[0].(*ast.ExpressionStatement)
	tryExpression, ok := statement.Expression.(*ast.TryExpression)
	if !ok {
		t.Fatalf("Expected the statement to have TryExpression type, got %T", statement.Expression)
	}

	if tryExpression.Expression.String() != "test.method()" {
		t.Fatalf("expected 'test.method()', got %q", tryExpression.Expression.String())
	}
}

func testIntegerLiteral(t *testing.T, rightExpression ast.Expression, integerValue int64) bool {
	integerExprValue, ok := rightExpression.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("Expected expression to have IntegerLiteral type, got %T", rightExpression)
		return false
	}

	if integerExprValue.Value != integerValue {
		t.Errorf("Expected value %d got %d", integerValue, integerExprValue.Value)
		return false
	}

	if integerExprValue.TokenLiteral() != fmt.Sprintf("%d", integerValue) {
		t.Errorf("Expected token literal %q got %q", integerValue, integerExprValue.TokenLiteral())
		return false
	}
	return true
}

func testBooleanLiteral(t *testing.T, expression ast.Expression, value bool) bool {
	identifier, ok := expression.(*ast.Boolean)
	if !ok {
		t.Errorf("expected Boolean type, got %T", expression)
		return false
	}

	if identifier.Value != value {
		t.Errorf("expected identifier value %t, got %t", value, identifier.Value)
		return false
	}

	if identifier.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("expected identifier value %t, got %q", value, identifier.TokenLiteral())
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

func testLiteralExpression(t *testing.T, expression ast.Expression, expected any) bool {
	switch value := expected.(type) {
	case int:
		return testIntegerLiteral(t, expression, int64(value))
	case int64:
		return testIntegerLiteral(t, expression, value)
	case bool:
		return testBooleanLiteral(t, expression, value)
	case string:
		return testIdentifier(t, expression, value)
	}
	t.Errorf("unhandled type for the passed expression, got %T", expression)
	return false
}

func testInfixExpression(t *testing.T, expression ast.Expression, left any,
	operator string, right any) bool {
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

func testVarStatement(t *testing.T, statement ast.Statement, name string) bool {
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
