package evaluator

import (
	"fmt"
	"math"
	"strings"

	"github.com/Abathargh/harlock/internal/ast"
	"github.com/Abathargh/harlock/internal/object"
)

type MethodMapping map[string]object.MethodFunction

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}

	builtins       map[string]*object.Builtin
	builtinMethods map[object.ObjectType]MethodMapping
)

func init() {
	builtins = make(map[string]*object.Builtin)
	builtins["hex"] = &object.Builtin{Function: builtinHex}
	builtins["len"] = &object.Builtin{Function: builtinLen}
	builtins["set"] = &object.Builtin{Function: builtinSet}
	builtins["type"] = &object.Builtin{Function: builtinType}
	builtins["open"] = &object.Builtin{Function: builtinOpen}
	builtins["save"] = &object.Builtin{Function: builtinSave}
	builtins["print"] = &object.Builtin{Function: builtinPrint}
	builtins["as_bytes"] = &object.Builtin{Function: builtinAsBytes}
	builtins["contains"] = &object.Builtin{Function: builtinContains}

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
		return evalPrefixExpression(currentNode.Operator, right)
	case *ast.InfixExpression:
		left := Eval(currentNode.LeftExpression, env)
		if isError(left) {
			return left
		}
		right := Eval(currentNode.RightExpression, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(currentNode.Operator, left, right)
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
		args := evalExpressions(currentNode.Arguments, env)
		return callFunction(currentNode.String(), functionCall, args)
	case *ast.ArrayLiteral:
		elements := evalExpressions(currentNode.Elements, env)
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
		return evalIndexExpression(left, index)
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

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalUnaryNotExpression(right)
	case "-":
		return evalUnaryMinusExpression(right)
	case "~":
		return evalBitwiseNotExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	if left.Type() != right.Type() {
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	}

	switch left.Type() {
	case object.IntegerObj:
		return evalIntegerInfixExpression(operator, left, right)
	case object.BooleanObj:
		return evalBooleanInfixExpression(operator, left, right)
	case object.StringObj:
		return evalStringInfixExpression(operator, left, right)
	case object.TypeObj:
		return evalTypeInfixExpression(operator, left, right)
	case object.ArrayObj:
		return evalArrayInfixExpression(operator, left, right)
	case object.MapObj:
		return evalMapInfixExpression(operator, left, right)
	case object.SetObj:
		return evalSetInfixExpression(operator, left, right)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
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

func evalUnaryMinusExpression(right object.Object) object.Object {
	if right.Type() != object.IntegerObj {
		return newError("unknown operator: -%s", right.Type())
	}

	intValue := right.(*object.Integer).Value
	return &object.Integer{Value: -intValue}
}

func evalBitwiseNotExpression(right object.Object) object.Object {
	if right.Type() != object.IntegerObj {
		return newError("unknown operator: ~%s", right.Type())
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

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
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
			return newError("division by zero")
		}
		return &object.Integer{Value: leftValue / rightValue}
	case "%":
		if rightValue == 0 {
			return newError("division by zero")
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
			return newError("negative bit-shift")
		}
		return &object.Integer{Value: leftValue << rightValue}
	case ">>":
		if rightValue < 0 {
			return newError("negative bit-shift")
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
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalBooleanInfixExpression(operator string, left, right object.Object) object.Object {
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
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
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
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalTypeInfixExpression(operator string, left, right object.Object) object.Object {
	leftType := left.(*object.Type).Value
	rightType := right.(*object.Type).Value
	switch operator {
	case "==":
		return getBoolReference(leftType == rightType)
	case "!=":
		return getBoolReference(leftType != rightType)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalArrayInfixExpression(operator string, left, right object.Object) object.Object {
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
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalMapInfixExpression(operator string, left, right object.Object) object.Object {
	leftMap := left.(*object.Map)
	rightMap := right.(*object.Map)
	switch operator {
	case "==":
		return getBoolReference(mapEquals(leftMap, rightMap))
	case "!=":
		return getBoolReference(!mapEquals(leftMap, rightMap))
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalSetInfixExpression(operator string, left, right object.Object) object.Object {
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
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if value, ok := env.Get(node.Value); ok {
		return value
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return newError("identifier not found: %s", node.Value)
}

func evalExpressions(expressions []ast.Expression, env *object.Environment) []object.Object {
	var evaluatedExpressions []object.Object
	for _, expression := range expressions {
		evaluatedExpr := Eval(expression, env)
		if isError(evaluatedExpr) {
			return []object.Object{evaluatedExpr}
		}
		evaluatedExpressions = append(evaluatedExpressions, evaluatedExpr)
	}
	return evaluatedExpressions
}

func evalIndexExpression(indexed, index object.Object) object.Object {
	switch {
	case indexed.Type() == object.ArrayObj && index.Type() == object.IntegerObj:
		return evalArrayIndexExpression(indexed, index)
	case indexed.Type() == object.MapObj:
		return evalMapIndexExpression(indexed, index)
	default:
		return newError("the index used to access an array must be an integer")
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	maxIdx := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > maxIdx {
		return newError("index error: %d is out of bounds", idx)
	}
	return arrayObject.Elements[idx]
}

func evalMapIndexExpression(hashmap, index object.Object) object.Object {
	mapObject := hashmap.(*object.Map)
	key, isHashable := index.(object.Hashable)
	if !isHashable {
		return newError("type error: index is not hashable")
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
			return newError("the passed key is not an hashable object")
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
		return newError("object of type %s has no method called %s", evaluatedCaller.Type(), methodName)
	}

	args := evalExpressions(methodExpression.Called.Arguments, env)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}
	expArgs := make([]object.Object, len(args)+1, cap(args)+1)
	expArgs[0] = evaluatedCaller
	copy(expArgs[1:], args)

	methodObj := &object.Method{MethodFunc: method}
	return callFunction(methodName, methodObj, expArgs)
}

func callFunction(funcName string, funcObj object.Object, args []object.Object) object.Object {
	switch function := funcObj.(type) {
	case *object.Function:
		if validateFunctionCall(function, args) {
			functionEnv := extendFunctionEnvironment(function, args)
			evaluatedFunction := Eval(function.Body, functionEnv)
			return unwrapReturnValue(evaluatedFunction)
		}
		nameOnly := funcName[:strings.Index(funcName, "(")]
		return newError("type error: function %q was called with a wrong number of args", nameOnly)
	case *object.Builtin:
		return function.Function(args...)
	case *object.Method:
		if len(args) == 1 {
			return function.MethodFunc(args[0])
		}
		return function.MethodFunc(args[0], args[1:]...)
	default:
		return newError("identifier is not a function: %s", funcObj.Type())
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
		res := evalInfixExpression("==", elem, obj2.Elements[idx])
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

		if !exists || evalInfixExpression("==", pair.Value, elemObj2.Value) != TRUE {
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
		if !exists || evalInfixExpression("==", val, elemObj2) != TRUE {
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
