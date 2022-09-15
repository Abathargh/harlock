package evaluator

import (
	"bufio"
	"fmt"
	"github.com/Abathargh/harlock/internal/evaluator/hex"
	"io"
	"os"
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
	case *object.Set:
		return &object.Integer{Value: int64(len(elem.Elements))}
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
	set := &object.Set{Elements: make(map[object.HashKey]object.Object)}
	if len(args) == 0 {
		return set
	}

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

func builtinContains(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("type error: contains requires two arguments, " +
			"the container and the element to test")
	}

	switch cont := args[0].(type) {
	case *object.Array:
		for _, elem := range cont.Elements {
			res := evalInfixExpression("==", args[1], elem)
			boolRes := res.(*object.Boolean)
			if boolRes.Value {
				return TRUE
			}
		}
		return FALSE
	case *object.Map:
		hashable, isHashable := args[1].(object.Hashable)
		if !isHashable {
			return newError("the passed key is not an hashable object")
		}
		_, contains := cont.Mappings[hashable.HashKey()]
		if contains {
			return TRUE
		}
		return FALSE
	case *object.Set:
		hashable, isHashable := args[1].(object.Hashable)
		if !isHashable {
			return newError("the passed key is not an hashable object")
		}
		_, contains := cont.Elements[hashable.HashKey()]
		if contains {
			return TRUE
		}
		return FALSE
	default:
		return newError("type error: the passed object is not a valid container")
	}
}

func builtinOpen(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("type error: open requires two arguments, " +
			"the input file and a string with the type of file")
	}

	filename, isString := args[0].(*object.String)
	if !isString {
		return newError("type error: expected a string with the file name "+
			"got %T", args[0])
	}

	fileType, isString := args[1].(*object.String)
	if !isString {
		return newError("type error: expected a string with the file type "+
			"got %T", args[0])
	}

	file, err := os.Open(filename.Value)
	if err != nil {
		return newError("file error: could not open file %q", filename.Value)
	}
	defer func() { _ = file.Close() }()

	switch fileType.Value {
	case "bytes":
		bytes, err := io.ReadAll(file)
		if err != nil {
			return newError("file error: cannot read the contents of the passed file")
		}
		info, _ := file.Stat()
		return object.NewBytesFile(file.Name(), uint32(info.Mode().Perm()), bytes)

	case "hex":
		hexFile, err := hex.ReadAll(bufio.NewReader(file))
		if err != nil {
			return newError("file error: %s", err)
		}
		info, _ := file.Stat()
		return object.NewHexFile(file.Name(), uint32(info.Mode().Perm()), hexFile)

	case "elf":
		// TODO
		fallthrough
	default:
		return newError("type error: unsupported file type")
	}
}

func builtinSave(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("type error: save requires only one argument " +
			"(a file object)")
	}
	switch file := args[0].(type) {
	case object.File:
		err := os.WriteFile(file.Name(), file.AsBytes(), os.FileMode(file.Perms()))
		if err != nil {
			return newError("file error: could not save file")
		}
		return nil
	default:
		return newError("type error: must pass a file (hex, elf, bytes)")
	}
}

func builtinAsBytes(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("type error: as_bytes requires only one argument (a file object)")
	}
	switch file := args[0].(type) {
	case object.File:
		bytes := file.AsBytes()
		buf := make([]object.Object, len(bytes))
		for idx, b := range bytes {
			buf[idx] = &object.Integer{Value: int64(b)}
		}
		return &object.Array{Elements: buf}
	default:
		return newError("type error: must pass a file (hex, elf, bytes)")
	}
}
