package evaluator

import "github.com/Abathargh/harlock/internal/object"

func arrayBuiltinPop(this object.Object, args ...object.Object) object.Object {
	arrayThis := this.(*object.Array)
	if len(args) != 0 {
		return newError("type error: pop does not require arguments")
	}

	newArrLen := len(arrayThis.Elements) - 1
	if newArrLen < 0 {
		return newError("type error: cannot pop from an empty array")
	}

	newArr := make([]object.Object, newArrLen, newArrLen)
	copy(newArr, arrayThis.Elements[:len(arrayThis.Elements)-1])
	return &object.Array{Elements: newArr}
}

func arrayBuiltinPush(this object.Object, args ...object.Object) object.Object {
	arrayThis := this.(*object.Array)
	if len(args) != 1 {
		return newError("type error: push requires an argument (the element to push)")
	}

	newArrLen := len(arrayThis.Elements) + 1
	newArr := make([]object.Object, newArrLen, newArrLen)
	copy(newArr, arrayThis.Elements)
	newArr[newArrLen-1] = args[0]
	return &object.Array{Elements: newArr}
}

func arrayBuiltinSlice(this object.Object, args ...object.Object) object.Object {
	arrayThis := this.(*object.Array)
	if len(args) != 2 {
		return newError("type error: slice requires two Int arguments (start and end)")
	}

	if args[0].Type() != object.IntegerObj || args[1].Type() != object.IntegerObj {
		return newError("type error: slice requires two Int arguments (start and end)")
	}

	start := args[0].(*object.Integer).Value
	end := args[1].(*object.Integer).Value

	arrayLen := int64(len(arrayThis.Elements))

	if end < start || end <= 0 || start < 0 || start >= arrayLen || end > arrayLen {
		return newError("type error: required end < start, 0 <= start < len, 0 < end <= len")
	}

	length := end - start
	slice := make([]object.Object, length, length)
	copy(slice, arrayThis.Elements[start:end])
	return &object.Array{Elements: slice}
}

func arrayBuiltinMap(this object.Object, args ...object.Object) object.Object {
	arrayThis := this.(*object.Array)
	if len(args) != 1 {
		return newError("type error: map requires a function literal")
	}

	fun, isFun := args[0].(*object.Function)
	if !isFun {
		return newError("type error: expected a function, got %T", args[0])
	}

	if len(fun.Parameters) != 1 {
		return newError("type error: the map callback requires exactly one argument (a one-args function(x) -> x)")
	}

	retArray := make([]object.Object, len(arrayThis.Elements))

	for idx, elem := range arrayThis.Elements {
		res := callFunction("<anonymous callback>", fun, []object.Object{elem}, noLineInfo)
		if res == nil || res.Type() == object.ErrorObj {
			return newError("type error: map requires a fun taking one arg and returning one value (function(x) -> x)")
		}
		retArray[idx] = res
	}
	return &object.Array{Elements: retArray}
}

func arrayBuiltinReduce(this object.Object, args ...object.Object) object.Object {
	arrayThis := this.(*object.Array)
	argn := len(args)
	if argn < 1 || argn > 2 {
		return newError("type error: reduce requires two arguments, " +
			"a function literal and an array, and can have one optional argument, " +
			"an accumulator")
	}

	fun, isFun := args[0].(*object.Function)
	if !isFun {
		return newError("type error: expected a function, got %T", args[0])
	}

	if len(fun.Parameters) != 2 {
		return newError("type error: the reduce callbacl requires only exactly two arguments " +
			"(a two args function(x, y) -> z)")
	}

	if len(arrayThis.Elements) == 0 {
		return newError("type error: expected a non-empty array")
	}

	var start int
	var accumulator object.Object
	if argn == 2 {
		start = 0
		accumulator = args[1]
	} else {
		start = 1
		accumulator = arrayThis.Elements[0]
	}

	for _, elem := range arrayThis.Elements[start:] {
		accumulator = callFunction("<anonymous function>", fun, []object.Object{accumulator, elem}, noLineInfo)
	}

	return accumulator
}
