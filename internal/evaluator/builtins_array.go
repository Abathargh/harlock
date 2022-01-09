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
