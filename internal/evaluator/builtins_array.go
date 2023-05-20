package evaluator

import (
	"github.com/Abathargh/harlock/internal/object"
)

func arrayBuiltinPop(this object.Object, _ ...object.Object) object.Object {
	arrayThis := this.(*object.Array)

	newArrLen := len(arrayThis.Elements) - 1
	if newArrLen < 0 {
		return newTypeError("cannot pop from an empty array")
	}

	newArr := make([]object.Object, newArrLen, newArrLen)
	copy(newArr, arrayThis.Elements[:len(arrayThis.Elements)-1])
	return &object.Array{Elements: newArr}
}

func arrayBuiltinPush(this object.Object, args ...object.Object) object.Object {
	arrayThis := this.(*object.Array)

	newArrLen := len(arrayThis.Elements) + 1
	newArr := make([]object.Object, newArrLen, newArrLen)
	copy(newArr, arrayThis.Elements)
	newArr[newArrLen-1] = args[0]
	return &object.Array{Elements: newArr}
}

func arrayBuiltinSlice(this object.Object, args ...object.Object) object.Object {
	arrayThis := this.(*object.Array)

	start := args[0].(*object.Integer).Value
	end := args[1].(*object.Integer).Value

	arrayLen := int64(len(arrayThis.Elements))

	if end < start || end <= 0 || start < 0 || start >= arrayLen || end > arrayLen {
		return newTypeError("required end < start, 0 <= start < len, 0 < end <= len")
	}

	length := end - start
	slice := make([]object.Object, length, length)
	copy(slice, arrayThis.Elements[start:end])
	return &object.Array{Elements: slice}
}

func arrayBuiltinMap(this object.Object, args ...object.Object) object.Object {
	arrayThis := this.(*object.Array)
	fun := args[0]

	switch callable := fun.(type) {
	case *object.Function:
		if len(callable.Parameters) != 1 {
			return newTypeError("the map callback requires exactly one argument (a one-args function(x) -> x)")
		}
	case *object.Builtin:
		if len(callable.GetBuiltinArgTypes()) != 1 {
			return newTypeError("the map callback requires exactly one argument (a one-args function(x) -> x)")
		}
	}

	retArray := make([]object.Object, len(arrayThis.Elements))

	for idx, elem := range arrayThis.Elements {
		res := callFunction("<anonymous callback>", fun, []object.Object{elem}, noLineInfo)
		if res == nil || res.Type() == object.ErrorObj {
			return newTypeError("map requires a fun taking one arg and returning one value (function(x) -> x)")
		}
		retArray[idx] = res
	}
	return &object.Array{Elements: retArray}
}

func arrayBuiltinReduce(this object.Object, args ...object.Object) object.Object {
	arrayThis := this.(*object.Array)
	argn := len(args)

	fun := args[0].(*object.Function)
	if len(fun.Parameters) != 2 {
		return newTypeError("the reduce callbacl requires only exactly two arguments " +
			"(a two args function(x, y) -> z)")
	}

	if len(arrayThis.Elements) == 0 {
		return newTypeError("expected a non-empty array")
	}

	start := 1
	accumulator := arrayThis.Elements[0]
	if argn == 2 {
		start = 0
		accumulator = args[1]
	}

	for _, elem := range arrayThis.Elements[start:] {
		accumulator = callFunction("<anonymous function>", fun, []object.Object{accumulator, elem}, noLineInfo)
	}

	return accumulator
}
