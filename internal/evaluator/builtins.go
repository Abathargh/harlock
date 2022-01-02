package evaluator

import (
	"fmt"

	"github.com/Abathargh/harlock/internal/object"
)

func builtinLen(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("len error: too many args")
	}

	switch elem := args[0].(type) {
	case *object.String:
		return &object.Integer{Value: int64(len(elem.Value))}
	case *object.Array:
		return &object.Integer{Value: int64(len(elem.Elements))}
	default:
		return newError("len error: type not supported")
	}
}

func builtinType(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("type error: too many args")
	}

	if args[0] == nil {
		return NULL
	}

	return &object.Type{Value: args[0].Type()}
}

func builtinOpen(args ...object.Object) object.Object {
	// TODO add file APIs
	return NULL
}

func builtinPrint(args ...object.Object) object.Object {
	if len(args) == 0 {
		return newError("type error: not enough args")
	}

	var ifcArgs []interface{}
	for _, arg := range args {
		ifcArgs = append(ifcArgs, arg.Inspect())
	}

	fmt.Println(ifcArgs...)
	return nil
}

func builtinPush(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("type error: push requires an array and the element to push")
	}

	array, isArray := args[0].(*object.Array)
	if !isArray {
		return newError("type error: first argument must be an array")
	}

	newArrLen := len(array.Elements) + 1
	newArr := make([]object.Object, newArrLen, newArrLen)
	copy(newArr, array.Elements)
	newArr[newArrLen-1] = args[1]
	return &object.Array{Elements: newArr}
}

func builtinSlice(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("type error: slice requires an array and the range to slice")
	}

	if args[0].Type() != object.ArrayObj || args[1].Type() != object.IntegerObj ||
		args[2].Type() != object.IntegerObj {
		return newError("type error: slice requires an array and the range to slice")
	}

	array := args[0].(*object.Array)
	start := args[1].(*object.Integer).Value
	end := args[2].(*object.Integer).Value

	arrayLen := int64(len(array.Elements))

	if end < start || end <= 0 || start < 0 || start >= arrayLen || end > arrayLen {
		return newError("type error: required end < start, 0 <= start < len, 0 < end <= len")
	}

	length := end - start
	slice := make([]object.Object, length, length)
	copy(slice, array.Elements[start:end])
	return &object.Array{Elements: slice}
}
