package evaluator

import (
	"fmt"
	"strconv"

	"github.com/Abathargh/harlock/internal/object"
)

func builtinHex(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("type error: hex requires one integer/string as argument")
	}

	switch argObj := args[0].(type) {
	case *object.Integer:
		value := argObj.Value
		sign := ""
		if value < 0 {
			sign = "-"
			value = -value
		}
		return &object.String{Value: fmt.Sprintf("%s0x%02x", sign, value)}
	case *object.String:
		strVal := argObj.Value
		strLen := len(strVal)
		if strLen%2 != 0 || strLen == 0 {
			return newError("type error: wrong size for hex string literal")
		}
		arr := make([]object.Object, strLen/2, strLen/2)
		for idx := 0; idx < strLen; idx += 2 {
			digit, err := strconv.ParseInt(strVal[idx:idx+2], 16, 64)
			if err != nil {
				return newError("type error: invalid hex digit %s", strVal[idx:idx+2])
			}
			arr[idx/2] = &object.Integer{Value: digit}
		}
		return &object.Array{Elements: arr}
	default:
		return newError("type error: hex requires one integer/string as argument")
	}
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

func builtinSet(args ...object.Object) object.Object {
	if len(args) == 0 {
		return newError("type error: not enough args, " +
			"either pass a sequence or a variable number of hashable objects")
	}

	set := &object.Set{Elements: make(map[object.HashKey]object.Object)}

	if len(args) == 1 {
		switch seq := args[0].(type) {
		case *object.Array:
			for _, elem := range seq.Elements {
				hashableElem, isHashable := elem.(object.Hashable)
				if !isHashable {
					return newError("the passed key is not an hashable object")
				}

				hash := hashableElem.HashKey()
				set.Elements[hash] = elem
			}
			return set
		case *object.Map:
			for key, pair := range seq.Mappings {
				set.Elements[key] = pair.Key
			}
			return set
		default:
			hashableElem, isHashable := seq.(object.Hashable)
			if !isHashable {
				return newError("the passed key is not an hashable object")
			}

			hash := hashableElem.HashKey()
			set.Elements[hash] = seq
			return set
		}
	}

	for _, elem := range args {
		hashableElem, isHashable := elem.(object.Hashable)
		if !isHashable {
			return newError("the passed key is not an hashable object")
		}

		hash := hashableElem.HashKey()
		set.Elements[hash] = elem
	}
	return set
}
