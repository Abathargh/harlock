package evaluator

import (
	"bufio"
	"bytes"
	"strings"
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
		{"~3", 252},
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
		{"false + true", "unknown operator: Bool + Bool"},
		{"false + 12", "type mismatch: Bool + Int"},
		{"-true", "unknown operator: -Bool"},
		{"~false", "unknown operator: ~Bool"},
		{"if 2 < 3 { ret 12 + true }", "type mismatch: Int + Bool"},
		{`"string" + 12`, "type mismatch: String + Int"},
		{`"string" + true`, "type mismatch: String + Bool"},
		{`"string" - "string2"`, "unknown operator: String - String"},
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
		{"fun(x) { print(x)\n ret x & 1 }(15)\n", 1},
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

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`hex(255)`, "0xff"},
		{`hex("ffab21")`, object.ArrayObj},
		{`len("")`, 0},
		{`len("ciao")`, 4},
		{`type("ciao")`, object.StringObj},
		{`type(1)`, object.IntegerObj},
		{`type(1/0)`, object.ErrorObj},
		{`type("ciao")`, object.StringObj},
		{`type([])`, object.ArrayObj},
		{`type({})`, object.MapObj},
		{`type(type([]))`, object.TypeObj},
		{`print("ciao")`, nil},
	}

	for _, testCase := range tests {
		evalBuiltin := testEval(testCase.input)
		switch expected := testCase.expected.(type) {
		case int:
			testIntegerObject(t, evalBuiltin, int64(expected))
		case object.ObjectType:
			// TODO case error, string
		}
	}
}

func TestArrayLiterals(t *testing.T) {
	input := `[5, 5 % 4, 6 & 2]`

	arrayObj := testEval(input)
	arrayLiteral, ok := arrayObj.(*object.Array)
	if !ok {
		t.Fatalf("expected object of Array type, got %T", arrayObj)
	}

	if len(arrayLiteral.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arrayLiteral.Elements))
	}

	testIntegerObject(t, arrayLiteral.Elements[0], 5)
	testIntegerObject(t, arrayLiteral.Elements[1], 1)
	testIntegerObject(t, arrayLiteral.Elements[2], 2)
}

func TestArrayIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"[1][0]", 1},
		{"[1, 2, 4][1 + 1]", 4},
		{`[0xfe, "ciao", 12][2]`, 12},
		{"var arr = [2, 5, 1]\narr[1]", 5},
		{"var add = fun(x,y){ ret x+y }\n[2, add(3, 4), 3][1]", 7},
		// TODO Errors
	}

	for _, testCase := range tests {
		arrayIndexExpr := testEval(testCase.input)
		expectedIntValue, isInt := testCase.expected.(int)
		if isInt {
			testIntegerObject(t, arrayIndexExpr, int64(expectedIntValue))
		} else {
			testNullObject(t, arrayIndexExpr)
		}
	}
}

func TestMapLiterals(t *testing.T) {
	input := `var test = 22
{
	"test1": 20 * 2,
	"test2": 2 & 3,
	"tes"+"t3": 4,
	test: 22,	
	true: 1,
	false: 0,
}`
	expected := map[object.HashKey]int64{
		(&object.String{Value: "test1"}).HashKey(): 40,
		(&object.String{Value: "test2"}).HashKey(): 2,
		(&object.String{Value: "test3"}).HashKey(): 4,
		(&object.Integer{Value: 22}).HashKey():     22,
		TRUE.HashKey():                             1,
		FALSE.HashKey():                            0,
	}

	evaluated := testEval(input)
	mapObj, ok := evaluated.(*object.Map)
	if !ok {
		t.Fatalf("expected object of Map type, got %T", evaluated)
	}

	if len(mapObj.Mappings) != len(expected) {
		t.Fatalf("expected %d elements, got %d", len(expected), len(mapObj.Mappings))
	}

	for expKey, expVal := range expected {
		mapping, ok := mapObj.Mappings[expKey]
		if !ok {
			t.Errorf("expected key %+v to be present in the map", expKey)
		}
		testIntegerObject(t, mapping.Value, expVal)
	}
}

func TestMapIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`{"test": 2}["test"]`, 2},
		{`{10: 3}[10]`, 3},
		{`{true: 4}[true]`, 4},
		// TODO Errors
	}

	for _, testCase := range tests {
		arrayIndexExpr := testEval(testCase.input)
		expectedIntValue, isInt := testCase.expected.(int)
		if isInt {
			testIntegerObject(t, arrayIndexExpr, int64(expectedIntValue))
		} else {
			testNullObject(t, arrayIndexExpr)
		}
	}
}

func TestArrayBuiltinMethods(t *testing.T) {
	tests := []struct {
		input    string
		expected []int64
	}{
		{`[1, 2].push(3)`, []int64{1, 2, 3}},
		{`[1, 2].pop()`, []int64{1}},
		{`[1, 2, 3, 4].slice(1, 3)`, []int64{2, 3}},
	}

	for _, testCase := range tests {
		evalArrayBuiltin := testEval(testCase.input)
		testArrayObject(t, evalArrayBuiltin, testCase.expected)
	}
}

func TestMapBuiltinMethods(t *testing.T) {
	tests := []struct {
		input    string
		expected [][]int64
	}{
		{"var m = {1: 2}\nm.set(3, 4)\nm", [][]int64{{1, 2}, {3, 4}}},
		{"var m  = {1: 2, 3: 4}\nm.pop(3)\nm", [][]int64{{1, 2}}},
	}

	for _, testCase := range tests {
		evalArrayBuiltin := testEval(testCase.input)
		testMapObject(t, evalArrayBuiltin, testCase.expected)
	}
}

func TestTryExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"var a = try 1\na", 1},
		{"var a = fun() { ret try 12 }\na()", 12},
		{"var a = fun() { ret try 1/0 }\na()", nil},
		{`
var mul = fun(x, y) {
	ret x / y
}
var double_mul = fun(x, y) {
	var m = try mul(x, y)
	ret 2 * m
}
double_mul(1, 0)`, nil},
	}

	for _, testCase := range tests {
		evalTryExpression := testEval(testCase.input)
		switch expected := testCase.expected.(type) {
		case int:
			testIntegerObject(t, evalTryExpression, int64(expected))
		default:
			if evalTryExpression.Type() != object.ErrorObj {
				t.Errorf("Expected an Error object, got %s", object.ErrorObj)
			}
		}
	}
}

func testEval(input string) object.Object {
	l := lexer.NewLexer(bufio.NewReader(bytes.NewBufferString(input)))
	p := parser.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		return &object.Error{Message: strings.Join(p.Errors(), ", ")}
	}
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

func testArrayObject(t *testing.T, obj object.Object, expected []int64) bool {
	arrayObj, ok := obj.(*object.Array)
	if !ok {
		t.Errorf("expected object to be an Array, got %T", obj)
		return false
	}

	if len(arrayObj.Elements) != len(expected) {
		t.Errorf("expected array with %d elements, got %d", len(arrayObj.Elements), len(expected))
		return false
	}

	for idx, element := range arrayObj.Elements {
		if !testIntegerObject(t, element, expected[idx]) {
			return false
		}
	}
	return true
}

func testMapObject(t *testing.T, obj object.Object, expected [][]int64) bool {
	mapObj, ok := obj.(*object.Map)
	if !ok {
		t.Errorf("expected object to be an Map, got %T", obj)
		return false
	}

	if len(mapObj.Mappings) != len(expected) {
		t.Errorf("expected array with %d elements, got %d", len(mapObj.Mappings), len(expected))
		return false
	}

	for _, pair := range expected {
		intKey := &object.Integer{Value: pair[0]}
		keyHash := intKey.HashKey()

		keyVal, contains := mapObj.Mappings[keyHash]

		if !contains {
			t.Errorf("expected to contain element with key %d", pair[0])
			return false
		}

		if !testIntegerObject(t, keyVal.Value, pair[1]) {
			return false
		}
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
	if obj != nil {
		t.Errorf("expected null, got %T", obj)
		return false
	}
	return true
}
