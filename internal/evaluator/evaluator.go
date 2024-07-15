package evaluator

import (
	"fmt"
	"math"
	"strings"

	"github.com/Abathargh/harlock/internal/ast"
	"github.com/Abathargh/harlock/internal/object"
)

type MethodMapping map[string]*object.Method

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

	// Builtin: hex(int|array) -> string
	// Converts an integer or a byte array to a hex-string.
	builtins["hex"] = &object.Builtin{
		Name: "hex",
		ArgTypes: []object.ObjectType{
			object.OrType(object.IntegerObj, object.ArrayObj),
		},
		Function: builtinHex,
	}

	// Builtin: from_hex(string) -> array
	// Converts a hex-string with to an array of bytes
	builtins["from_hex"] = &object.Builtin{
		Name:     "from_hex",
		ArgTypes: []object.ObjectType{object.StringObj},
		Function: builtinFromhex,
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
		ArgTypes: []object.ObjectType{object.AnyObj},
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

	// Builtin: print(...any) -> no return
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

	// Builtin: hash(array, string) -> array
	// Returns an array containing the computed hash of the passed
	// array, using the specified algorithm.
	builtins["hash"] = &object.Builtin{
		Name:     "hash",
		ArgTypes: []object.ObjectType{object.ArrayObj, object.StringObj},
		Function: builtinHash,
	}

	// Builtin: int(string) -> int
	// Converts a string representing an integer to an actual integer.
	builtins["int"] = &object.Builtin{
		Name:     "int",
		ArgTypes: []object.ObjectType{object.StringObj},
		Function: builtinInt,
	}

	// Builtin: error(...any) -> error
	// Creates a custom error that can be used in code.
	builtins["error"] = &object.Builtin{
		Name:     "error",
		ArgTypes: []object.ObjectType{object.AnyVarargs},
		Function: builtinError,
	}

	// Builtin: as_array(int, int, string) -> array
	// Converts an integer to its representation as an array of bytes of specific
	// size and endianness.
	builtins["as_array"] = &object.Builtin{
		Name:     "as_array",
		ArgTypes: []object.ObjectType{object.IntegerObj, object.IntegerObj, object.StringObj},
		Function: builtinAsArray,
	}

	// Builtin: builtinHelp(int, int, string) -> array
	// Converts an integer to its representation as an array of bytes of specific
	// size and endianness.
	builtins["help"] = &object.Builtin{
		Name:     "help",
		ArgTypes: []object.ObjectType{object.StringObj},
		Function: builtinHelp,
	}

	builtinMethods = make(map[object.ObjectType]MethodMapping)
	builtinMethods[object.ArrayObj] = MethodMapping{
		// Builtin: array.map(function) -> array
		// Applies the passed function to each element of the array and returns a new
		// array with the modified values.
		"map": &object.Method{
			Name: "array.map",
			ArgTypes: []object.ObjectType{
				object.OrType(object.FunctionObj, object.BuiltinObj),
			},
			MethodFunc: arrayBuiltinMap,
		},

		// Builtin: array.pop() -> array
		// Removes the last element from the array and returns a copy of the new array.
		"pop": &object.Method{
			Name:       "array.pop",
			ArgTypes:   []object.ObjectType{},
			MethodFunc: arrayBuiltinPop,
		},

		// Builtin: array.push(any) -> array
		// Adds an element to the tail of the array and returns the new array.
		// The original array remains unchanged.
		"push": &object.Method{
			Name:       "array.push",
			ArgTypes:   []object.ObjectType{object.AnyObj},
			MethodFunc: arrayBuiltinPush,
		},

		// Builtin: array.slice(int, int) -> array
		// Returns a sub-array slicing the original array in the [args[0]:args[1])
		// interval. This returns a new array and copies each element in the new
		// array. Lists/Maps/Sets/Files are copied as references.
		"slice": &object.Method{
			Name:       "array.slice",
			ArgTypes:   []object.ObjectType{object.IntegerObj, object.IntegerObj},
			MethodFunc: arrayBuiltinSlice,
		},

		// Builtin: array.reduce(function [, any]) -> any
		// Applies the passed function to each element of the array; the first argument
		// gets used as the result of the previous iteration. An accumulator init value
		// can be passed as an optional final argument.
		"reduce": &object.Method{
			Name:       "array.reduce",
			ArgTypes:   []object.ObjectType{object.FunctionObj, object.AnyOptional},
			MethodFunc: arrayBuiltinReduce,
		},
	}

	builtinMethods[object.MapObj] = MethodMapping{
		// Builtin: map.set(any, any) -> no return
		// Adds the (arg[0], arg[1]) key value couple to the map. This mutates the map.
		"set": &object.Method{
			Name:       "map.set",
			ArgTypes:   []object.ObjectType{object.AnyObj, object.AnyObj},
			MethodFunc: mapBuiltinSet,
		},

		// Builtin: map.pop(any) -> no return
		// Removes the passed key from the map if it exists. This mutates the map.
		"pop": &object.Method{
			Name:       "map.pop",
			ArgTypes:   []object.ObjectType{object.AnyObj},
			MethodFunc: mapBuiltinPop,
		},
	}

	builtinMethods[object.SetObj] = MethodMapping{
		// Builtin: set.add(any) -> no return
		// Adds the element to the set. This mutates the set.
		"add": &object.Method{
			Name:       "set.add",
			ArgTypes:   []object.ObjectType{object.AnyObj},
			MethodFunc: setBuiltinAdd,
		},

		// Builtin: set.remove(any) -> no return
		// Removes the passed element from the set if it exists. This mutates the set.
		"remove": &object.Method{
			Name:       "set.remove",
			ArgTypes:   []object.ObjectType{object.AnyObj},
			MethodFunc: setBuiltinRemove,
		},
	}

	builtinMethods[object.HexObj] = MethodMapping{
		// Builtin: hex.record(int) -> string
		// Returns the nth record as a string, if it exists and is a valid index, or an error.
		"record": &object.Method{
			Name:       "hex.record",
			ArgTypes:   []object.ObjectType{object.IntegerObj},
			MethodFunc: hexBuiltinRecord,
		},

		// Builtin: hex.size(int) -> int
		// Returns the size of the file as a number of records it contains.
		"size": &object.Method{
			Name:       "hex.size",
			ArgTypes:   []object.ObjectType{},
			MethodFunc: hexBuiltinSize,
		},

		// Builtin: hex.read_at(int, int) -> array
		// Attempts to read arg[1] number of bytes starting from arg[0] position.
		// This returns an array containing the data that would be found in the
		// corresponding .bin file obtained from the hex file as a byte stream.
		"read_at": &object.Method{
			Name:       "hex.read_at",
			ArgTypes:   []object.ObjectType{object.IntegerObj, object.IntegerObj},
			MethodFunc: hexBuiltinReadAt,
		},

		// Builtin: hex.write_at(int, array) -> no return
		// Attempts to write the contents of the arg[1] byte array to the  arg[0]
		// position. This mutates the hex file object but not the copy on disk.
		// Call the save() function to make the changes persistent.
		"write_at": &object.Method{
			Name:       "hex.write_at",
			ArgTypes:   []object.ObjectType{object.IntegerObj, object.ArrayObj},
			MethodFunc: hexBuiltinWriteAt,
		},

		// Builtin: hex.binary_size(int) -> int
		// Returns the size of the file as the actual number of bytes contained in the data
		// section of the data records found within the hex file.
		"binary_size": &object.Method{
			Name:       "hex.binary_size",
			ArgTypes:   []object.ObjectType{},
			MethodFunc: hexBuiltinBinarySize,
		},
	}

	builtinMethods[object.ElfObj] = MethodMapping{
		// Builtin: elf.has_section(string) -> bool
		// Returns whether the elf file contains a section with the passed name or not.
		"has_section": &object.Method{
			Name:       "elf.has_section",
			ArgTypes:   []object.ObjectType{object.StringObj},
			MethodFunc: elfBuiltinHasSection,
		},

		// Builtin: elf.sections() -> array
		// Returns an array containing the section header names as strings.
		"sections": &object.Method{
			Name:       "elf.sections",
			ArgTypes:   []object.ObjectType{},
			MethodFunc: elfBuiltinSections,
		},

		// Builtin: elf.section_address(string) -> int
		// Returns the address of the specified section, if it exists.
		"section_address": &object.Method{
			Name:       "elf.section_address",
			ArgTypes:   []object.ObjectType{object.StringObj},
			MethodFunc: elfBuiltinSectionAddress,
		},

		// Builtin: elf.section_size(string) -> int
		// Returns the size of the specified section, if it exists.
		"section_size": &object.Method{
			Name:       "elf.section_address",
			ArgTypes:   []object.ObjectType{object.StringObj},
			MethodFunc: elfBuiltinSectionSize,
		},

		// Builtin: elf.read_section(string) -> array
		// Attempts to read the contents of the specified section, if it exists, and
		// returns it as a byte array.
		"read_section": &object.Method{
			Name:       "elf.read_section",
			ArgTypes:   []object.ObjectType{object.StringObj},
			MethodFunc: elfBuiltinReadSection,
		},

		// Builtin: elf.write_section(string, array, int) -> no return
		// Attempts to write the contents of the arg[1] byte array to the arg[0]
		// section with arg[2] offset. This mutates the elf file object but not the copy on disk.
		// Call the save() function to make the changes persistent.
		"write_section": &object.Method{
			Name:       "elf.write_section",
			ArgTypes:   []object.ObjectType{object.StringObj, object.ArrayObj, object.IntegerObj},
			MethodFunc: elfBuiltinWriteSection,
		},
	}

	builtinMethods[object.BytesObj] = MethodMapping{
		// Builtin: bytes.read_at(int, int) -> array
		// Attempts to read arg[1] number of bytes starting from arg[0] position.
		// This returns an array containing the data that would be found in the
		// corresponding .bin file obtained from the bytes file as a byte stream.
		"read_at": &object.Method{
			Name:       "bytes.read_at",
			ArgTypes:   []object.ObjectType{object.IntegerObj, object.IntegerObj},
			MethodFunc: bytesBuiltinReadAt,
		},

		// Builtin: bytes.write_at(int, array) -> no return
		// Attempts to write the contents of the arg[1] byte array to the  arg[0]
		// position. This mutates the bytes file object but not the copy on disk.
		// Call the save() function to make the changes persistent.
		"write_at": &object.Method{
			Name:       "bytes.write_at",
			ArgTypes:   []object.ObjectType{object.IntegerObj, object.ArrayObj},
			MethodFunc: bytesBuiltinWriteAt,
		},
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
			if isError(returnValue) {
				return returnValue
			}
			return &object.ReturnValue{Value: returnValue}
		}
		return &object.ReturnValue{Value: NULL}
	case *ast.VarStatement:
		varValue := Eval(currentNode.Value, env)
		if isError(varValue) {
			return varValue
		}
		if varValue == nil || varValue == NULL {
			return NULL
		}
		if varValue.Type() == object.ReturnValueObj {
			unwrapped := unwrapReturnValue(varValue)
			if unwrapped.Type() == object.RuntimeErrorObj {
				return varValue
			}
		}
		env.Set(currentNode.Name.Value, varValue)
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
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
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
		if isRuntimeError(exprValue) {
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
		if isReturnValOrError(result) {
			return result
		}
	}
	return result
}

func isReturnValOrError(obj object.Object) bool {
	switch {
	case obj == nil:
		return false
	case obj.Type() == object.ReturnValueObj:
		fallthrough
	case obj.Type() == object.ErrorObj:
		fallthrough
	case obj.Type() == object.RuntimeErrorObj:
		return true
	default:
		return false
	}
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
	case indexed.Type() == object.ArrayObj && index.Type() != object.IntegerObj:
		return newError("attempting to use a non-integer as an array index on line %d", line)
	default:
		return newError("attempting to index a non-subscriptable object (%s) on line %d", indexed.Type(), line)
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
		return newKeyError("%s", index.Inspect())
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
	if isError(evaluatedCaller) {
		return evaluatedCaller
	}

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

	return callFunction(methodName, method, expArgs, methodExpression.LineNumber)
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
		return execBuiltin(function, line, args...)
	case *object.Method:
		return execBuiltin(function, line, args...)
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

func newError(format string, args ...any) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, args...)}
}

func isError(obj object.Object) bool {
	if obj == nil {
		return false
	}
	t := obj.Type()
	return t == object.ErrorObj
}

func newTypeError(msg string, args ...any) *object.RuntimeError {
	return &object.RuntimeError{
		Kind:    object.TypeError,
		Message: fmt.Sprintf(msg, args...),
	}
}

func newKeyError(msg string, args ...any) *object.RuntimeError {
	return &object.RuntimeError{
		Kind:    object.KeyError,
		Message: fmt.Sprintf(msg, args...),
	}
}

func newFileError(msg string, args ...any) *object.RuntimeError {
	return &object.RuntimeError{
		Kind:    object.FileError,
		Message: fmt.Sprintf(msg, args...),
	}
}

func newHexError(msg string, args ...any) *object.RuntimeError {
	return &object.RuntimeError{
		Kind:    object.HexError,
		Message: fmt.Sprintf(msg, args...),
	}
}

func newElfError(msg string, args ...any) *object.RuntimeError {
	return &object.RuntimeError{
		Kind:    object.ElfError,
		Message: fmt.Sprintf(msg, args...),
	}
}

func newBytesError(msg string, args ...any) *object.RuntimeError {
	return &object.RuntimeError{
		Kind:    object.BytesError,
		Message: fmt.Sprintf(msg, args...),
	}
}

func newCustomError(msg string, args ...any) *object.RuntimeError {
	return &object.RuntimeError{
		Kind:    object.CustomError,
		Message: fmt.Sprintf(msg, args...),
	}
}

func isRuntimeError(obj object.Object) bool {
	if obj == nil {
		return false
	}
	return obj.Type() == object.RuntimeErrorObj
}
