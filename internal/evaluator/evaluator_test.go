package evaluator

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/Abathargh/harlock/internal/lexer"
	"github.com/Abathargh/harlock/internal/object"
	"github.com/Abathargh/harlock/internal/parser"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue int64
	}{
		{"15", 15},
		{"23", 23},
		{"-5", -5},
		{"--5", 5},
		{"-10", -10},
		{"~3", -4},
		{"1 | 2", 3},
		{"1 & 2", 0},
		{"(2 + 2) * 3", 12},
		{"2 + 2 * 3", 8},
		{"4 / 2", 2},
		{"(1 << 2) / 2 ", 2},
		{"(2 >> 1) * 2 / 2", 1},
		{"1 ^ 1", 0},
		{"2 * (8 % 3)", 4},
	}

	for _, testCase := range tests {
		evaluatedObj := testEval(testCase.input)
		testIntegerObject(t, evaluatedObj, testCase.expectedValue)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"3 < 2", false},
		{"2 > 1", true},
		{"0 > 1", false},
		{"1 <= 2", true},
		{"2 <= 2", true},
		{"5 <= 4", false},
		{"3 >= 2", true},
		{"2 >= 2", true},
		{"3 >= 4", false},
		{"2 > 1", true},
		{"0 > 1", false},
		{"5 == 5", true},
		{"4 == 5", false},
		{"4 != 5", true},
		{"5 != 5", false},
		{"true == true", true},
		{"true == false", false},
		{"false == true", false},
		{"false == false", true},
		{"true != true", false},
		{"true != false", true},
		{"false != true", true},
		{"false != false", false},
		{"(5 > 4) != false", true},
		{"(5 > 4) == false", false},
		{"true && true", true},
		{"true && false", false},
		{"false && true", false},
		{"false && false", false},
		{"true || true", true},
		{"true || false", true},
		{"false || true", true},
		{"false || false", false},
	}

	for _, testCase := range tests {
		evaluatedObj := testEval(testCase.input)
		testBooleanObject(t, evaluatedObj, testCase.expectedValue)
	}
}

func TestUnaryNotOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!!true", true},
		{"!!false", false},
		{"!5", false},
	}

	for _, testCase := range tests {
		evaluatedNotExpression := testEval(testCase.input)
		testBooleanObject(t, evaluatedNotExpression, testCase.expected)
	}
}

func TestIfElseExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if true { 1 }", 1},
		{"if false { 1 }", nil},
		{"if 12 { 2 }", 2},
		{"if 2 < 3 { 3 }", 3},
		{"if 2 > 3 { b }", nil},
		{"if true && true { 4 }", 4},
		{"if false || false { b }", nil},
		{"if false || false { b } else { 12 }", 12},
		{"if true && false { b } else { 24 }", 24},
	}

	for _, testCase := range tests {
		evaluatedIfExpression := testEval(testCase.input)
		expectedInt, ok := testCase.expected.(int)
		if ok {
			testIntegerObject(t, evaluatedIfExpression, int64(expectedInt))
		} else {
			testNullObject(t, evaluatedIfExpression)
		}
	}
}

func TestReturnStatement(t *testing.T) {
	tests := []struct {
		input               string
		expectedReturnValue int64
	}{
		{"ret 10\n", 10},
		{"ret 5 * 5\n", 25},
		{"ret (4+3)*2\n", 14},
		{"ret 2\na\n", 2},
		{`
if 10 > 1 {
	if 10 > 1 {
		ret 10
	}
	ret 1
}
`, 10},
	}

	for _, testCase := range tests {
		evaluatedReturn := testEval(testCase.input)
		testIntegerObject(t, evaluatedReturn, testCase.expectedReturnValue)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input            string
		expectedErrorMsg string
	}{
		{"false + true", "unknown operator: BOOLEAN + BOOLEAN"},
		{"false + 12", "type mismatch: BOOLEAN + INTEGER"},
		{"-true", "unknown operator: -BOOLEAN"},
		{"~false", "unknown operator: ~BOOLEAN"},
		{"if 2 < 3 { ret 12 + true }", "type mismatch: INTEGER + BOOLEAN"},
		{`"string" + 12`, "type mismatch: STRING + INTEGER"},
		{`"string" + true`, "type mismatch: STRING + BOOLEAN"},
		{`"string" - "string2"`, "unknown operator: STRING - STRING"},
	}

	for _, testCase := range tests {
		evaluatedError := testEval(testCase.input)
		errorObj, ok := evaluatedError.(*object.Error)
		if !ok {
			t.Errorf("expected Error type, got %T (%+v)", evaluatedError, testCase.input)
			continue
		}

		if errorObj.Message != testCase.expectedErrorMsg {
			t.Errorf("expected %s error, got %s", testCase.expectedErrorMsg, errorObj.Message)
		}
	}
}

func TestVarStatement(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue int64
	}{
		{"var a = 5\na", 5},
		{"var a = (5 * 2) % 4\na", 2},
		{"var a = 4 % 4\nvar b = a\n b", 0},
		{"var a = 4\nvar b = a + 2\nvar c = b\nc", 6},
	}

	for _, testCase := range tests {
		testIntegerObject(t, testEval(testCase.input), testCase.expectedValue)
	}
}

func TestFunctionLiterals(t *testing.T) {
	input := "fun(a) { a * a }\n"
	expectedFunBody := "(a*a)"

	obj := testEval(input)
	functionObject, ok := obj.(*object.Function)
	if !ok {
		t.Fatalf("expected object of Function type, got %T", obj)
	}

	if len(functionObject.Parameters) != 1 {
		t.Fatalf("expected 1 parameters, got %d", len(functionObject.Parameters))
	}

	if functionObject.Parameters[0].String() != "a" {
		t.Fatalf("expected a parameter with name \"a\", got %s", functionObject.Parameters[0].String())
	}

	if functionObject.Body.String() != expectedFunBody {
		t.Errorf("expected function body = %q, got %q", expectedFunBody, functionObject.Body.String())
	}
}

func TestFunction(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput int64
	}{
		{"var a = fun(x) { x }\na(20)\n", 20},
		{"var mul = fun(x, y) { ret x * y }\nmul(2, 3)\n", 6},
		{"var double = fun(x) { ret x << 2 }\ndouble(5)\n", 20},
		{"fun(x) { ret x & 1 }(15)\n", 1},
		{"var mod = fun(x, y) { ret x % y }\n mod(mod(6, 5), 3)", 1},
	}

	for _, testCase := range tests {
		testIntegerObject(t, testEval(testCase.input), testCase.expectedOutput)
	}
}

func TestStringOperators(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput interface{}
	}{
		{`'single ' + 'single'`, "single single"},
		{`'single ' + "double"`, "single double"},
		{`"double " + 'single'`, "double single"},
		{`"double " + "double"`, "double double"},
		{`"single" == "single"`, true},
		{`"single" == "double"`, false},
		{`"single" != "single"`, false},
		{`"single" != "double"`, true},
		{`'single' == 'single'`, true},
		{`'single' == 'double'`, false},
		{`'single' != 'single'`, false},
		{`'single' != 'double'`, true},
	}

	for _, testCase := range tests {
		evalString := testEval(testCase.input)
		switch result := evalString.(type) {
		case *object.String:
			if result.Value != testCase.expectedOutput {
				t.Errorf("expected %s, got %s", testCase.expectedOutput, result.Value)
			}
		case *object.Boolean:
			if result.Value != testCase.expectedOutput {
				t.Errorf("expected %t, got %t", testCase.expectedOutput, result.Value)
			}
		default:
			t.Errorf("expected expression of type String or Boolean, got %T", result)
		}

	}
}

func TestStringLiteral(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
	}{
		{`'test single quotes'`, "test single quotes"},
		{`"test double quotes"`, "test double quotes"},
	}

	for _, testCase := range tests {
		evalString := testEval(testCase.input)
		stringObj, ok := evalString.(*object.String)
		if !ok {
			t.Fatalf("expected String type, got %T", evalString)
		}

		if stringObj.Value != testCase.expectedOutput {
			t.Errorf("expected %s, got %s", testCase.expectedOutput, stringObj.Value)
		}
	}
}

func testEval(input string) object.Object {
	l := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(input)))
	p := parser.NewParser(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()
	return Eval(program, env)
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	integerObj, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("expected object to be an Integer, got %T", obj)
		return false
	}

	if integerObj.Value != expected {
		t.Errorf("expected %d, got %d", expected, integerObj.Value)
		return false
	}
	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	booleanObj, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("expected object to be an Boolean, got %T", obj)
		return false
	}

	if booleanObj.Value != expected {
		t.Errorf("expected %t, got %t", expected, booleanObj.Value)
		return false
	}
	return true
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("expected null, got %T", obj)
		return false
	}
	return true
}
