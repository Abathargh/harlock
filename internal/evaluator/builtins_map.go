package evaluator

import "github.com/Abathargh/harlock/internal/object"

func mapBuiltinSet(this object.Object, args ...object.Object) object.Object {
	mapThis := this.(*object.Map)

	hashableKey, isHashable := args[0].(object.Hashable)
	if !isHashable {
		return newError("type error: map_set requires an hashable key")
	}

	hashedKey := hashableKey.HashKey()
	mapThis.Mappings[hashedKey] = object.HashPair{Key: args[0], Value: args[1]}
	return nil
}

func mapBuiltinPop(this object.Object, args ...object.Object) object.Object {
	mapThis := this.(*object.Map)

	hashableKey, isHashable := args[0].(object.Hashable)
	if !isHashable {
		return newError("type error: the passed key is not an hashable type")
	}
	delete(mapThis.Mappings, hashableKey.HashKey())
	return nil
}
