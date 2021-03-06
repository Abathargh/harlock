package evaluator

import "github.com/Abathargh/harlock/internal/object"

func setBuiltinAdd(this object.Object, args ...object.Object) object.Object {
	setThis := this.(*object.Set)
	if len(args) != 1 {
		return newError("type error: add requires one argument")
	}

	hashable, isHashable := args[0].(object.Hashable)
	if !isHashable {
		return newError("type error: the passed type is not hashable")
	}

	key := hashable.HashKey()
	setThis.Elements[key] = args[0]
	return nil
}

func setBuiltinRemove(this object.Object, args ...object.Object) object.Object {
	setThis := this.(*object.Set)
	if len(args) != 1 {
		return newError("type error: remove requires one argument")
	}

	hashable, isHashable := args[0].(object.Hashable)
	if !isHashable {
		return newError("type error: the passed type is not hashable")
	}

	key := hashable.HashKey()
	delete(setThis.Elements, key)
	return nil
}
