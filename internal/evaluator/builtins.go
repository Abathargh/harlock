package evaluator

import (
	"fmt"

	"github.com/Abathargh/harlock/internal/object"
)

func builtinHex(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("type error: hex requires one integer as argument")
	}

	intArg, isInt := args[0].(*object.Integer)
	if !isInt {
		return newError("type error: hex requires one integer as argument")
	}

	value := intArg.Value
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}

	return &object.String{Value: fmt.Sprintf("%s0x%02x", sign, value)}
}

func builtinLen(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("len error: too many args")
	}

	switch elem := args[0].(type) {
	case *object.String:
		return &object.Integer{Value: int64(len(elem.Value))}
	case *object.Array:
		return &object.Integer{Value: int64(len(elem.Elements))}
	case *object.Map:
		return &object.Integer{Value: int64(len(elem.Mappings))}
	default:
		return newError("type error: type not supported")
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

func builtinMapSet(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("type error: map_set requires a map, the key and element to set")
	}

	hashmap, isMap := args[0].(*object.Map)
	if !isMap {
		return newError("type error: map_set requires a map as first argument")
	}

	hashableKey, isHashable := args[1].(object.Hashable)
	if !isHashable {
		return newError("type error: map_set requires an hashable key")
	}

	hashedKey := hashableKey.HashKey()
	hashmap.Mappings[hashedKey] = object.HashPair{Key: args[1], Value: args[2]}
	return nil
}

func builtinPop(args ...object.Object) object.Object {
	if len(args) == 0 {
		return newError("type error: not enough arguments")
	}

	switch dataStructure := args[0].(type) {
	case *object.Array:
		return builtinPopArray(dataStructure)
	case *object.Map:
		if len(args) != 2 {
			return newError("type error: not enough arguments to call pop for a map")
		}
		return builtinPopMap(dataStructure, args[1])
	default:
		return newError("type error: invalid data structure")
	}
}

func builtinPopArray(array *object.Array) object.Object {
	newArrLen := len(array.Elements) - 1
	if newArrLen < 0 {
		return newError("type error: cannot pop from an empty array")
	}
	newArr := make([]object.Object, newArrLen, newArrLen)
	copy(newArr, array.Elements[:len(array.Elements)-1])
	return &object.Array{Elements: newArr}
}

func builtinPopMap(hashmap *object.Map, key object.Object) object.Object {
	hashableKey, isHashable := key.(object.Hashable)
	if !isHashable {
		return newError("type error: the passed key is not an hashable type")
	}
	delete(hashmap.Mappings, hashableKey.HashKey())
	return nil
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
