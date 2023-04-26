package evaluator

import (
	"fmt"
	"math"
	"strings"

	"github.com/Abathargh/harlock/internal/ast"
	"github.com/Abathargh/harlock/internal/object"
)

type MethodMapping map[string]object.MethodFunction

const noLineInfo = -1

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}

	builtins       map[string]*object.Builtin
	builtinMethods map[object.ObjectType]MethodMapping
)

func init() {
	builtins = make(map[string]*object.Builtin)

	// Builtin: hex(int|string) -> string|array
	// Converts an integer to a hex-string or a hex-string
	// with no trailing '0x' to an array of bytes
	builtins["hex"] = &object.Builtin{
		Name: "hex",
		ArgTypes: []object.ObjectType{
			object.OrType(object.IntegerObj, object.StringObj),
		},
		Function: builtinHex,
	}

	// Builtin: len(string|array|map|set) -> int
	// Returns the length of the passed collection type.
	builtins["len"] = &object.Builtin{
		Name: "len",
		ArgTypes: []object.ObjectType{
			object.OrType(object.StringObj, object.ArrayObj, object.MapObj, object.SetObj),
		},
		Function: builtinLen,
	}

	// Builtin: set(...) -> set
	// Builds a set starting from the passed elements.
	// If one of the elements is iterable, its elements are
	// iterated instead of adding the iterable itself.
	builtins["set"] = &object.Builtin{
		Name:     "set",
		ArgTypes: []object.ObjectType{object.AnyVarargs},
		Function: builtinSet,
	}

	// Builtin: type(any) -> string
	// Returns the type of the object as a string.
	builtins["type"] = &object.Builtin{
		Name:     "type",
		ArgTypes: []object.ObjectType{},
		Function: builtinType,
	}

	// Builtin: open(string, string) -> file
	// Attempts to open a file with the name of the first
	// argument, with the file type specified by the second argument.
	builtins["open"] = &object.Builtin{
		Name:     "open",
		ArgTypes: []object.ObjectType{object.StringObj, object.StringObj},
		Function: builtinOpen,
	}

	// Builtin: save(hex_file|elf_file|bytes_file) -> no return
	// Saves a previously opened file's contents unto the original file.
	builtins["save"] = &object.Builtin{
		Name: "save",
		ArgTypes: []object.ObjectType{
			object.OrType(object.HexObj, object.ElfObj, object.BytesObj),
		},
		Function: builtinSave,
	}

	// Builtin: print(...) -> no return
	// Prints every passed object as a string separated by a space, with
	// a newline character at the end.
	builtins["print"] = &object.Builtin{
		Name:     "print",
		ArgTypes: []object.ObjectType{object.AnyVarargs},
		Function: builtinPrint,
	}

	// Builtin: as_bytes(hex_file|elf_file|bytes_file) -> array
	// Returns an array containing the passed file as a stream of bytes.
	builtins["as_bytes"] = &object.Builtin{
		Name: "as_bytes",
		ArgTypes: []object.ObjectType{
			object.OrType(object.HexObj, object.ElfObj, object.BytesObj),
		},
		Function: builtinAsBytes,
	}

	// Builtin: contains(any, array|map|set) -> bool
	// Returns true if the collection contains the passed object.
	builtins["contains"] = &object.Builtin{
		Name: "contains",
		ArgTypes: []object.ObjectType{
			object.OrType(object.ArrayObj, object.MapObj, object.SetObj),
			object.AnyObj,
		},
		Function: builtinContains,
	}

	builtinMethods = make(map[object.ObjectType]MethodMapping)
	builtinMethods[object.ArrayObj] = MethodMapping{
		"map":    arrayBuiltinMap,
		"pop":    arrayBuiltinPop,
		"push":   arrayBuiltinPush,
		"slice":  arrayBuiltinSlice,
		"reduce": arrayBuiltinReduce,
	}

	builtinMethods[object.MapObj] = MethodMapping{
		"set": mapBuiltinSet,
		"pop": mapBuiltinPop,
	}

	builtinMethods[object.SetObj] = MethodMapping{
		"add":    setBuiltinAdd,
		"remove": setBuiltinRemove,
	}

	builtinMethods[object.HexObj] = MethodMapping{
		"record":   hexBuiltinRecord,
		"size":     hexBuiltinSize,
		"read_at":  hexBuiltinReadAt,
		"write_at": hexBuiltinWriteAt,
	}

	builtinMethods[object.ElfObj] = MethodMapping{
		"has_section":   elfBuiltinHasSection,
		"sections":      elfBuiltinSections,
		"write_section": elfBuiltinWriteSection,
		"read_section":  elfBuiltinReadSection,
	}

	builtinMethods[object.BytesObj] = MethodMapping{
		"read_at":  bytesBuiltinReadAt,
		"write_at": bytesBuiltinWriteAt,
	}
}

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch currentNode := node.(type) {
	case *ast.Program:
		return evalProgram(currentNode, env)
	case *ast.ExpressionStatement:
		return Eval(currentNode.Expression, env)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: currentNode.Value}
	case *ast.Boolean:
		return getBoolReference(currentNode.Value)
	case *ast.StringLiteral:
		return &object.String{Value: currentNode.Value}
	case *ast.PrefixExpression:
		right := Eval(currentNode.RightExpression, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(currentNode.Operator, right, currentNode.LineNumber)
	case *ast.InfixExpression:
		left := Eval(currentNode.LeftExpression, env)
		if isError(left) {
			return left
		}
		right := Eval(currentNode.RightExpression, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(currentNode.Operator, left, right, currentNode.LineNumber)
	case *ast.BlockStatement:
		return evalBlockStatement(currentNode, env)
	case *ast.IfExpression:
		return evalIfExpression(currentNode, env)
	case *ast.ReturnStatement:
		if currentNode.ReturnValue != nil {
			returnValue := Eval(currentNode.ReturnValue, env)
			return &object.ReturnValue{Value: returnValue}
		}
		return &object.ReturnValue{Value: NULL}
	case *ast.VarStatement:
		letValue := Eval(currentNode.Value, env)
		if letValue != nil && letValue.Type() == object.ReturnValueObj {
			return unwrapReturnValue(letValue)
		}
		env.Set(currentNode.Name.Value, letValue)
	case *ast.NoOp:
		// do nothing
	case *ast.Identifier:
		return evalIdentifier(currentNode, env)
	case *ast.FunctionLiteral:
		parameters := currentNode.Parameters
		functionBody := currentNode.Body
		return &object.Function{Parameters: parameters, Body: functionBody, Env: env}
	case *ast.CallExpression:
		functionCall := Eval(currentNode.Function, env)
		args := evalExpressions(currentNode.Arguments, env, currentNode.LineNumber)
		return callFunction(currentNode.String(), functionCall, args, currentNode.LineNumber)
	case *ast.ArrayLiteral:
		elements := evalExpressions(currentNode.Elements, env, currentNode.LineNumber)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}
	case *ast.IndexExpression:
		left := Eval(currentNode.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(currentNode.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index, currentNode.LineNumber)
	case *ast.MapLiteral:
		return evalMapLiteral(currentNode, env)
	case *ast.MethodCallExpression:
		return evalMethodExpression(currentNode, env)
	case *ast.TryExpression:
		exprValue := Eval(currentNode.Expression, env)
		if isError(exprValue) {
			return &object.ReturnValue{Value: exprValue}
		}
		return exprValue
	}
	return nil
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range program.Statements {
		result = Eval(statement, env)
		switch actualResult := result.(type) {
		case *object.ReturnValue:
			return actualResult.Value
		case *object.Error:
			return actualResult
		}
	}
	return result
}

func evalPrefixExpression(operator string, right object.Object, line int) object.Object {
	switch operator {
	case "!":
		return evalUnaryNotExpression(right)
	case "-":
		return evalUnaryMinusExpression(right, line)
	case "~":
		return evalBitwiseNotExpression(right, line)
	default:
		return newError("unknown operator: %s%s on line %d", operator, right.Type(), line)
	}
}

func evalInfixExpression(operator string, left, right object.Object, line int) object.Object {
	if left.Type() != right.Type() {
		return newError("type mismatch: %s %s %s on line %d", left.Type(), operator, right.Type(), line)
	}

	switch left.Type() {
	case object.IntegerObj:
		return evalIntegerInfixExpression(operator, left, right, line)
	case object.BooleanObj:
		return evalBooleanInfixExpression(operator, left, right, line)
	case object.StringObj:
		return evalStringInfixExpression(operator, left, right, line)
	case object.TypeObj:
		return evalTypeInfixExpression(operator, left, right, line)
	case object.ArrayObj:
		return evalArrayInfixExpression(operator, left, right, line)
	case object.MapObj:
		return evalMapInfixExpression(operator, left, right, line)
	case object.SetObj:
		return evalSetInfixExpression(operator, left, right, line)
	default:
		return newError("unknown operator: %s %s %s on line %d", left.Type(), operator, right.Type(), line)
	}
}

func evalBlockStatement(blockStatement *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range blockStatement.Statements {
		result = Eval(statement, env)
		if result != nil &&
			(result.Type() == object.ReturnValueObj || result.Type() == object.ErrorObj) {
			return result
		}
	}
	return result
}

func evalIfExpression(expression *ast.IfExpression, env *object.Environment) object.Object {
	ifCondition := Eval(expression.Condition, env)
	if isError(ifCondition) {
		return ifCondition
	}

	if isTruthy(ifCondition) {
		return Eval(expression.Consequence, env)
	} else if expression.Alternative != nil {
		return Eval(expression.Alternative, env)
	} else {
		return nil
	}
}

func evalUnaryNotExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalUnaryMinusExpression(right object.Object, line int) object.Object {
	if right.Type() != object.IntegerObj {
		return newError("unsupported operand '%s' for unary minus on line %d", right.Type(), line)
	}

	intValue := right.(*object.Integer).Value
	return &object.Integer{Value: -intValue}
}

func evalBitwiseNotExpression(right object.Object, line int) object.Object {
	if right.Type() != object.IntegerObj {
		return newError("unsupported operand '%s' for bitwise not on line %d", right.Type(), line)
	}

	intValue := right.(*object.Integer).Value
	var invertedValue int64
	switch {
	case intValue < 0:
		invertedValue = ^intValue
	case intValue >= 0 && intValue <= math.MaxUint8:
		invertedValue = int64(^uint8(intValue))
	case intValue > math.MaxUint8 && intValue <= math.MaxUint16:
		invertedValue = int64(^uint16(intValue))
	case intValue > math.MaxUint16 && intValue <= math.MaxUint32:
		invertedValue = int64(^uint32(intValue))
	default:
		invertedValue = ^intValue
	}
	return &object.Integer{Value: invertedValue}
}

func evalIntegerInfixExpression(operator string, left, right object.Object, line int) object.Object {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftValue + rightValue}
	case "-":
		return &object.Integer{Value: leftValue - rightValue}
	case "*":
		return &object.Integer{Value: leftValue * rightValue}
	case "/":
		if rightValue == 0 {
			return newError("division by zero on line %d", line)
		}
		return &object.Integer{Value: leftValue / rightValue}
	case "%":
		if rightValue == 0 {
			return newError("division by zero on line %d", line)
		}
		return &object.Integer{Value: leftValue % rightValue}
	case "|":
		return &object.Integer{Value: leftValue | rightValue}
	case "&":
		return &object.Integer{Value: leftValue & rightValue}
	case "^":
		return &object.Integer{Value: leftValue ^ rightValue}
	case "<<":
		if rightValue < 0 {
			return newError("attemping a negative bit-shift on line %d", line)
		}
		return &object.Integer{Value: leftValue << rightValue}
	case ">>":
		if rightValue < 0 {
			return newError("attemping a negative bit-shift on line %d", line)
		}
		return &object.Integer{Value: leftValue >> rightValue}
	case "==":
		return getBoolReference(leftValue == rightValue)
	case "!=":
		return getBoolReference(leftValue != rightValue)
	case ">":
		return getBoolReference(leftValue > rightValue)
	case "<":
		return getBoolReference(leftValue < rightValue)
	case ">=":
		return getBoolReference(leftValue >= rightValue)
	case "<=":
		return getBoolReference(leftValue <= rightValue)
	default:
		return newError("unknown operator %s %s %s on line %d", left.Type(), operator, right.Type(), line)
	}
}

func evalBooleanInfixExpression(operator string, left, right object.Object, line int) object.Object {
	leftValue := left.(*object.Boolean).Value
	rightValue := right.(*object.Boolean).Value

	switch operator {
	case "==":
		return getBoolReference(leftValue == rightValue)
	case "!=":
		return getBoolReference(leftValue != rightValue)
	case "&&":
		return getBoolReference(leftValue && rightValue)
	case "||":
		return getBoolReference(leftValue || rightValue)
	default:
		return newError("unknown operator %s %s %s on line %d", left.Type(), operator, right.Type(), line)
	}
}

func evalStringInfixExpression(operator string, left, right object.Object, line int) object.Object {
	leftString := left.(*object.String).Value
	rightString := right.(*object.String).Value
	switch operator {
	case "+":
		return &object.String{Value: leftString + rightString}
	case "==":
		return getBoolReference(leftString == rightString)
	case "!=":
		return getBoolReference(leftString != rightString)
	default:
		return newError("unsupported operator %s %s %s on line %d", left.Type(), operator, right.Type(), line)
	}
}

func evalTypeInfixExpression(operator string, left, right object.Object, line int) object.Object {
	leftType := left.(*object.Type).Value
	rightType := right.(*object.Type).Value
	switch operator {
	case "==":
		return getBoolReference(leftType == rightType)
	case "!=":
		return getBoolReference(leftType != rightType)
	default:
		return newError("unknown operator %s %s %s on line %d", left.Type(), operator, right.Type(), line)
	}
}

func evalArrayInfixExpression(operator string, left, right object.Object, line int) object.Object {
	leftArray := left.(*object.Array)
	rightArray := right.(*object.Array)
	switch operator {
	case "+":
		return &object.Array{Elements: append(leftArray.Elements, rightArray.Elements...)}
	case "==":
		return getBoolReference(arrayEquals(leftArray, rightArray))
	case "!=":
		return getBoolReference(!arrayEquals(leftArray, rightArray))
	default:
		return newError("unknown operator %s %s %s on line %d", left.Type(), operator, right.Type(), line)
	}
}

func evalMapInfixExpression(operator string, left, right object.Object, line int) object.Object {
	leftMap := left.(*object.Map)
	rightMap := right.(*object.Map)
	switch operator {
	case "==":
		return getBoolReference(mapEquals(leftMap, rightMap))
	case "!=":
		return getBoolReference(!mapEquals(leftMap, rightMap))
	default:
		return newError("unknown operator %s %s %s on line %d", left.Type(), operator, right.Type(), line)
	}
}

func evalSetInfixExpression(operator string, left, right object.Object, line int) object.Object {
	leftSet := left.(*object.Set)
	rightSet := right.(*object.Set)
	set := &object.Set{Elements: make(map[object.HashKey]object.Object)}

	switch operator {
	case "+":
		for key, elem := range leftSet.Elements {
			set.Elements[key] = elem
		}
		for key, elem := range rightSet.Elements {
			set.Elements[key] = elem
		}
		return set
	case "-":
		for key, elem := range leftSet.Elements {
			set.Elements[key] = elem
		}
		for key := range rightSet.Elements {
			delete(set.Elements, key)
		}
		return set
	case "^":
		for key, elem := range leftSet.Elements {
			if _, contains := rightSet.Elements[key]; contains {
				set.Elements[key] = elem
			}
		}
		for key, elem := range rightSet.Elements {
			if _, contains := leftSet.Elements[key]; contains {
				set.Elements[key] = elem
			}
		}
		return set
	case "==":
		return getBoolReference(setEquals(leftSet, rightSet))
	case "!=":
		return getBoolReference(!setEquals(leftSet, rightSet))
	default:
		return newError("unknown operator %s %s %s on line %d", left.Type(), operator, right.Type(), line)
	}
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if value, ok := env.Get(node.Value); ok {
		return value
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return newError("undefined identifier '%s' on line %d", node.Value, node.LineNumber)
}

func evalExpressions(expressions []ast.Expression, env *object.Environment, line int) []object.Object {
	var evaluatedExpressions []object.Object
	for _, expression := range expressions {
		evaluatedExpr := Eval(expression, env)
		if isError(evaluatedExpr) {
			err := evaluatedExpr.(*object.Error)
			err.Message += fmt.Sprintf(" on line %d", line)
			return []object.Object{evaluatedExpr}
		}
		evaluatedExpressions = append(evaluatedExpressions, evaluatedExpr)
	}
	return evaluatedExpressions
}

func evalIndexExpression(indexed, index object.Object, line int) object.Object {
	switch {
	case indexed.Type() == object.ArrayObj && index.Type() == object.IntegerObj:
		return evalArrayIndexExpression(indexed, index, line)
	case indexed.Type() == object.MapObj:
		return evalMapIndexExpression(indexed, index, line)
	default:
		return newError("attempting to use a non-integer as an array index on line %d", line)
	}
}

func evalArrayIndexExpression(array, index object.Object, line int) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	maxIdx := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > maxIdx {
		return newError("attempted an out of bounds access to an array with index %d on line %d ", idx, line)
	}
	return arrayObject.Elements[idx]
}

func evalMapIndexExpression(hashmap, index object.Object, line int) object.Object {
	mapObject := hashmap.(*object.Map)
	key, isHashable := index.(object.Hashable)
	if !isHashable {
		return newError("attempted to access a map with a non-hashable key on line %d", line)
	}

	pair, ok := mapObject.Mappings[key.HashKey()]
	if !ok {
		// element is not present, default is null for now
		return NULL
	}
	return pair.Value
}

func evalMapLiteral(mapLiteral *ast.MapLiteral, env *object.Environment) object.Object {
	mappings := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range mapLiteral.Mappings {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("attempted to access a map with a non-hashable key on line %d", mapLiteral.LineNumber)
		}

		value := Eval(valueNode, env)
		if isError(key) {
			return key
		}

		hashedValue := hashKey.HashKey()
		mappings[hashedValue] = object.HashPair{Key: key, Value: value}
	}
	return &object.Map{Mappings: mappings}
}

func evalMethodExpression(methodExpression *ast.MethodCallExpression, env *object.Environment) object.Object {
	evaluatedCaller := Eval(methodExpression.Caller, env)

	methodName := methodExpression.Called.Function.String()
	method, exists := builtinMethods[evaluatedCaller.Type()][methodName]
	if !exists {
		return newError("%s has no method called %s on line %d", evaluatedCaller.Type(), methodName, methodExpression.LineNumber)
	}

	args := evalExpressions(methodExpression.Called.Arguments, env, methodExpression.LineNumber)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}
	expArgs := make([]object.Object, len(args)+1, cap(args)+1)
	expArgs[0] = evaluatedCaller
	copy(expArgs[1:], args)

	methodObj := &object.Method{MethodFunc: method}
	return callFunction(methodName, methodObj, expArgs, methodExpression.LineNumber)
}

func callFunction(funcName string, funcObj object.Object, args []object.Object, line int) object.Object {
	switch function := funcObj.(type) {
	case *object.Function:
		if validateFunctionCall(function, args) {
			functionEnv := extendFunctionEnvironment(function, args)
			evaluatedFunction := Eval(function.Body, functionEnv)
			return unwrapReturnValue(evaluatedFunction)
		}
		nameOnly := funcName[:strings.Index(funcName, "(")]
		return newError("function %q was called with a wrong number of args on line %d", nameOnly, line)
	case *object.Builtin:
		return execBuiltin(function, args...) // this is an actual go function call
	case *object.Method:
		if len(args) == 1 {
			return function.MethodFunc(args[0])
		}
		return function.MethodFunc(args[0], args[1:]...)
	default:
		return newError("'%s' identifier is not a function on line %d", funcObj.Type(), line)
	}
}

func validateFunctionCall(function *object.Function, args []object.Object) bool {
	return len(function.Parameters) == len(args)
}

func extendFunctionEnvironment(function *object.Function, args []object.Object) *object.Environment {
	newEnv := object.WrappedEnvironment(function.Env)
	for idx, parameter := range function.Parameters {
		newEnv.Set(parameter.Value, args[idx])
	}
	return newEnv
}

func unwrapReturnValue(returnObj object.Object) object.Object {
	if returned, ok := returnObj.(*object.ReturnValue); ok {
		return returned.Value
	}
	return returnObj
}

func getBoolReference(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func isTruthy(obj object.Object) bool {
	// TODO add support for collections
	// TODO empty => false, non-empty => true
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func arrayEquals(obj1, obj2 *object.Array) bool {
	if obj1 == obj2 {
		return true
	}

	if len(obj1.Elements) != len(obj2.Elements) {
		return false
	}

	for idx, elem := range obj1.Elements {
		res := evalInfixExpression("==", elem, obj2.Elements[idx], noLineInfo)
		if res != TRUE {
			return false
		}
	}
	return true
}

func mapEquals(obj1, obj2 *object.Map) bool {
	if obj1 == obj2 {
		return true
	}

	if len(obj1.Mappings) != len(obj2.Mappings) {
		return false
	}

	for _, pair := range obj1.Mappings {
		hashable := pair.Key.(object.Hashable)
		hashedKey := hashable.HashKey()
		elemObj2, exists := obj2.Mappings[hashedKey]

		if !exists || evalInfixExpression("==", pair.Value, elemObj2.Value, noLineInfo) != TRUE {
			return false
		}
	}
	return true
}

func setEquals(obj1, obj2 *object.Set) bool {
	if obj1 == obj2 {
		return true
	}

	if len(obj1.Elements) != len(obj2.Elements) {
		return false
	}

	for key, val := range obj1.Elements {
		elemObj2, exists := obj2.Elements[key]
		if !exists || evalInfixExpression("==", val, elemObj2, noLineInfo) != TRUE {
			return false
		}
	}
	return true
}

func newError(format string, args ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, args...)}
}

func isError(obj object.Object) bool {
	if obj == nil {
		return false
	}
	return obj.Type() == object.ErrorObj
}
