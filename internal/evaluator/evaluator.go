package evaluator

import (
	"fmt"

	"github.com/Abathargh/harlock/internal/ast"
	"github.com/Abathargh/harlock/internal/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}

	builtins = map[string]*object.Builtin{
		"len":   {Function: builtinLen},
		"type":  {Function: builtinType},
		"push":  {Function: builtinPush},
		"slice": {Function: builtinSlice},
		"print": {Function: builtinPrint},
	}
)

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
			if isError(returnValue) {
				return returnValue
			}
			return &object.ReturnValue{Value: returnValue}
		}
		return &object.ReturnValue{Value: NULL}
	case *ast.VarStatement:
		letValue := Eval(currentNode.Value, env)
		if isError(letValue) {
			return letValue
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
		if isError(functionCall) {
			return functionCall
		}
		args := evalExpressions(currentNode.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return callFunction(functionCall, args)
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
	return &object.Integer{Value: ^intValue}
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
	if indexed.Type() == object.ArrayObj && index.Type() == object.IntegerObj {
		return evalArrayIndexExpression(indexed, index)
	}
	return newError("the index used to access an array must be an integer")
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	maxIdx := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > maxIdx {
		// TODO return error, not NULL
		return NULL
	}
	return arrayObject.Elements[idx]
}

func callFunction(funcObj object.Object, args []object.Object) object.Object {
	switch function := funcObj.(type) {
	case *object.Function:
		functionEnv := extendFunctionEnvironment(function, args)
		evaluatedFunction := Eval(function.Body, functionEnv)
		return unwrapReturnValue(evaluatedFunction)
	case *object.Builtin:
		return function.Function(args...)
	default:
		return newError("identifier is not a function: %s", funcObj.Type())
	}
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

func newError(format string, args ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, args...)}
}

func isError(obj object.Object) bool {
	if obj == nil {
		return false
	}
	return obj.Type() == object.ErrorObj
}
