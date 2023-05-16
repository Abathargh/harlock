package evaluator

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	hex2 "encoding/hex"
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

	if len(argsTypeStr) == 0 {
		argsTypeStr = "no args"
	}
	errorStr := fmt.Sprintf(typeErrTemplate, name, len(reqTypes), reqStr, name, argsValueStr, argsTypeStr)
	return &object.Error{Message: errorStr}
}

func execBuiltin(builtin object.CallableBuiltin, args ...object.Object) object.Object {
	name := builtin.GetBuiltinName()
	argTypes := builtin.GetBuiltinArgTypes()

	argcExpectedCount := 0
	argcExpected := len(argTypes)
	for _, expected := range argTypes {
		if expected == object.AnyOptional {
			argcExpectedCount++
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

	switch argcExpectedCount {
	case 0:
		if argcExpected != argc {
			return typeArgsError(builtin, argsToValidate)
		}
	default:
		if argc < argcExpected-argcExpectedCount || argc > argcExpected {
			return typeArgsError(builtin, argsToValidate)
		}
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
	case *object.Array:
		byteData := make([]byte, len(argObj.Elements))
		if err := intArrayToBytes(argObj, byteData); err != nil {
			return err
		}
		return &object.String{Value: hex2.EncodeToString(byteData)}
	default:
		return newError("hex requires one integer/string as argument")
	}
}

func builtinFromhex(args ...object.Object) object.Object {
	hexString := args[0].(*object.String)
	strVal := hexString.Value
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
	return &object.String{Value: string(args[0].Type())}
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

func builtinHash(args ...object.Object) object.Object {
	data := args[0].(*object.Array)
	hashFunc := args[1].(*object.String)

	// TODO: right now this iterates everything twice
	byteData := make([]byte, len(data.Elements))
	if err := intArrayToBytes(data, byteData); err != nil {
		return err
	}

	switch hashFunc.Value {
	case "sha1":
		sha1Sum := sha1.Sum(byteData)
		return bytestoIntarray(sha1Sum[:])
	case "sha256":
		sha256um := sha256.Sum256(byteData)
		return bytestoIntarray(sha256um[:])
	case "md5":
		md5Sum := md5.Sum(byteData)
		return bytestoIntarray(md5Sum[:])
	default:
		return newError("unsupported hash function %s", hashFunc.Value)
	}
}

func intArrayToBytes(src *object.Array, dst []byte) *object.Error {
	for idx, obj := range src.Elements {
		intByte, isInt := obj.(*object.Integer)
		if !isInt || (intByte.Value < 0 || intByte.Value > 255) {
			return newError("expecting an array of bytes (0 <= n <= 255)")
		}
		dst[idx] = byte(intByte.Value)
	}
	return nil
}

func bytestoIntarray(data []byte) *object.Array {
	arr := &object.Array{
		Elements: make([]object.Object, len(data)),
	}

	for idx, elem := range data {
		arr.Elements[idx] = &object.Integer{Value: int64(elem)}
	}
	return arr
}
