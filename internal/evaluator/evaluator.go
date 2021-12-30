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
		returnValue := Eval(currentNode.ReturnValue, env)
		if isError(returnValue) {
			return returnValue
		}
		return &object.ReturnValue{Value: returnValue}
	case *ast.VarStatement:
		letValue := Eval(currentNode.Value, env)
		if isError(letValue) {
			return letValue
		}
		env.Set(currentNode.Name.Value, letValue)
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
	switch {
	case left.Type() == object.IntegerObj && right.Type() == object.IntegerObj:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.BooleanObj && right.Type() == object.BooleanObj:
		return evalBooleanInfixExpression(operator, left, right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
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
		return NULL
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
		return &object.Integer{Value: leftValue / rightValue}
	case "%":
		return &object.Integer{Value: leftValue % rightValue}
	case "|":
		return &object.Integer{Value: leftValue | rightValue}
	case "&":
		return &object.Integer{Value: leftValue & rightValue}
	case "^":
		return &object.Integer{Value: leftValue ^ rightValue}
	case "<<":
		return &object.Integer{Value: leftValue << rightValue}
	case ">>":
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

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	value, ok := env.Get(node.Value)
	if !ok {
		return newError("identifier not found: %s", node.Value)
	}
	return value
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

func callFunction(funcObj object.Object, args []object.Object) object.Object {
	function, ok := funcObj.(*object.Function)
	if !ok {
		return newError("identifier is not a function: %s", funcObj.Type())
	}
	functionEnv := extendFunctionEnvironment(function, args)
	evaluatedFunction := Eval(function.Body, functionEnv)
	return unwrapReturnValue(evaluatedFunction)
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
