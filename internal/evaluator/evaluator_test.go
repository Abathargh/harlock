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
		{"a\nret (4+3)*2\n", 14},
		{"a\nret 2\na\n", 2},
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
		{"if 2 < 3 { ret 12 + true }\n", "type mismatch: INTEGER + BOOLEAN"},
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

func testEval(input string) object.Object {
	l := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(input)))
	p := parser.NewParser(l)
	program := p.ParseProgram()
	return Eval(program)
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
