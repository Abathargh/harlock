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
		return newError("type error: too many args")
	}

	var ifcArgs []interface{}
	for _, arg := range args {
		ifcArgs = append(ifcArgs, arg.Inspect())
	}

	fmt.Println(ifcArgs...)
	return nil
}
