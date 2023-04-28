package evaluator

import (
	"bufio"
	"fmt"
	"github.com/Abathargh/harlock/internal/evaluator/bytes"
	harlockElf "github.com/Abathargh/harlock/internal/evaluator/elf"
	"github.com/Abathargh/harlock/internal/evaluator/hex"
	"github.com/Abathargh/harlock/internal/object"
	"os"
	"strconv"
	"strings"
)

const (
	typeErrTemplate = "type error: function '%s' requires %d parameter(s) (%s), got %s(%s) (%s)"
	typeErrNoArgs   = "type error: function '%s' - %s"
)

func checkType(expected, actual object.ObjectType) bool {

	okTypes := strings.Split(string(expected), "/")
	for _, okType := range okTypes {
		if object.ObjectType(okType) == actual {
			return true
		}
	}
	return false
}

func typeArgsError(builtin object.CallableBuiltin, args []object.Object) *object.Error {
	name := builtin.GetBuiltinName()
	reqTypes := builtin.GetBuiltinArgTypes()

	argValues := make([]string, len(args))
	for idx, obj := range args {
		argValues[idx] = strings.ReplaceAll(obj.Inspect(), "\n", " ")
	}

	argTypes := make([]string, len(args))
	for idx, obj := range args {
		argTypes[idx] = fmt.Sprintf("1 %s", obj.Type())
	}

	argsValueStr := strings.Join(argValues, ", ")
	argsTypeStr := strings.Join(argTypes, ", ")
	reqStrList := make([]string, len(reqTypes))
	for idx, reqArg := range reqTypes {
		reqStrList[idx] = string(reqArg)
	}

	reqStr := strings.Join(reqStrList, ", ")

	errorStr := fmt.Sprintf(typeErrTemplate, name, len(reqTypes), reqStr, name, argsValueStr, argsTypeStr)
	return &object.Error{Message: errorStr}
}

func execBuiltin(builtin object.CallableBuiltin, args ...object.Object) object.Object {
	name := builtin.GetBuiltinName()
	argTypes := builtin.GetBuiltinArgTypes()

	argcExpected := len(argTypes)
	for _, expected := range argTypes {
		if expected == object.AnyOptional {
			argcExpected--
		}
	}

	argc := len(args)
	var argsToValidate []object.Object

	_, isMethod := builtin.(*object.Method)
	if isMethod {
		argc -= 1
		argsToValidate = args[1:] // do not validate 'self'/'this'
	} else {
		argsToValidate = args
	}

	if argcExpected == 1 && argTypes[0] == object.AnyVarargs {
		goto exec
	}

	// TODO this varies for AnyOptional
	if argcExpected != argc {
		return typeArgsError(builtin, argsToValidate)
	}

	for idx, argExpected := range argTypes {
		if argExpected == object.AnyObj || argExpected == object.AnyOptional {
			continue
		}

		if !checkType(argExpected, argsToValidate[idx].Type()) {
			return typeArgsError(builtin, argsToValidate)
		}
	}

exec:
	outcome := builtin.Call(args...)
	switch typedOutcome := outcome.(type) {
	case *object.Error:
		return newError(typeErrNoArgs, name, typedOutcome.Message)
	default:
		return outcome
	}
}

func builtinHex(args ...object.Object) object.Object {
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
				return newError("invalid hex digit %s", strVal[idx:idx+2])
			}
			arr[idx/2] = &object.Integer{Value: digit}
		}
		return &object.Array{Elements: arr}
	default:
		return newError("hex requires one integer/string as argument")
	}
}

func builtinLen(args ...object.Object) object.Object {
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
		return newError("type not supported")
	}
}

func builtinType(args ...object.Object) object.Object {
	if args[0] == nil {
		return NULL
	}
	return &object.Type{Value: args[0].Type()}
}

func builtinPrint(args ...object.Object) object.Object {
	var ifcArgs []any
	for _, arg := range args {
		if arg != nil {
			ifcArgs = append(ifcArgs, arg.Inspect())
		}
	}
	fmt.Println(ifcArgs...)
	return nil
}

func builtinSet(args ...object.Object) object.Object {
	set := &object.Set{Elements: make(map[object.HashKey]object.Object)}
	for _, arg := range args {
		switch seq := arg.(type) {
		case *object.Array:
			for _, elem := range seq.Elements {
				hashableElem, isHashable := elem.(object.Hashable)
				if !isHashable {
					return newError("the passed key is not an hashable object")
				}

				hash := hashableElem.HashKey()
				set.Elements[hash] = elem
			}
		case *object.Map:
			for key, pair := range seq.Mappings {
				set.Elements[key] = pair.Key
			}
		default:
			hashableElem, isHashable := seq.(object.Hashable)
			if !isHashable {
				return newError("the passed key is not an hashable object")
			}

			hash := hashableElem.HashKey()
			set.Elements[hash] = seq
		}
	}
	return set
}

func builtinContains(args ...object.Object) object.Object {
	switch cont := args[0].(type) {
	case *object.Array:
		for _, elem := range cont.Elements {
			res := evalInfixExpression("==", args[1], elem, noLineInfo)
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
		return newError("the passed object is not a valid container")
	}
}

func builtinOpen(args ...object.Object) object.Object {
	filename := args[0].(*object.String)
	fileType := args[1].(*object.String)

	file, err := os.Open(filename.Value)
	if err != nil {
		return newError("could not open file %q", filename.Value)
	}
	defer func() { _ = file.Close() }()

	switch fileType.Value {
	case "bytes":
		bytesFile, err := bytes.ReadAll(file)
		if err != nil {
			return newError("cannot read the contents of the passed file")
		}
		info, _ := file.Stat()
		return object.NewBytesFile(file.Name(), uint32(info.Mode().Perm()), info.Size(), bytesFile)

	case "hex":
		hexFile, err := hex.ReadAll(bufio.NewReader(file))
		if err != nil {
			return newError("file error - %s", err)
		}
		info, _ := file.Stat()
		return object.NewHexFile(file.Name(), uint32(info.Mode().Perm()), hexFile)

	case "elf":
		elfFile, err := harlockElf.ReadAll(file)
		if err != nil {
			return newError("file error - %s", err)
		}
		info, _ := file.Stat()
		return object.NewElfFile(file.Name(), uint32(info.Mode().Perm()), elfFile)

	default:
		return newError("unsupported file type")
	}
}

func builtinSave(args ...object.Object) object.Object {
	switch file := args[0].(type) {
	case object.File:
		err := os.WriteFile(file.Name(), file.AsBytes(), os.FileMode(file.Perms()))
		if err != nil {
			return newError("could not save file")
		}
		return nil
	default:
		return newError("must pass a file (hex, elf, bytes)")
	}
}

func builtinAsBytes(args ...object.Object) object.Object {
	switch file := args[0].(type) {
	case object.File:
		bs := file.AsBytes()
		buf := make([]object.Object, len(bs))
		for idx, b := range bs {
			buf[idx] = &object.Integer{Value: int64(b)}
		}
		return &object.Array{Elements: buf}
	default:
		return newError("must pass a file (hex, elf, bytes)")
	}
}
