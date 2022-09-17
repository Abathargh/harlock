package evaluator

import "github.com/Abathargh/harlock/internal/object"

func bytesBuiltinWriteAt(this object.Object, args ...object.Object) object.Object {
	bytesThis := this.(*object.BytesFile)
	if len(args) != 2 {
		return newError("type error: write_at requires two arguments " +
			"(the data array and the position)")
	}

	data, isArray := args[0].(*object.Array)
	if !isArray {
		return newError("type error: data must be an array")
	}

	position, isInt := args[1].(*object.Integer)
	if !isInt || position.Value < 0 {
		return newError("type error: position must be a positive integer")
	}

	byteArr := make([]byte, len(data.Elements))
	for idx, elem := range data.Elements {
		intElem, isInt := elem.(*object.Integer)
		if !isInt || intElem.Value > maxByte || intElem.Value < 0 {
			return newError("type error: data must be an array of 1 byte positive integers "+
				"(data[%d] = %d does not follow this constraint)", idx, intElem.Value)
		}
		byteArr[idx] = byte(intElem.Value)
	}

	err := bytesThis.Bytes.WriteAt(int(position.Value), byteArr)
	if err != nil {
		return newError("bytes error: bytes.write_at([%d], %d): %s",
			len(byteArr), uint64(position.Value), err)
	}
	return nil
}

func bytesBuiltinReadAt(this object.Object, args ...object.Object) object.Object {
	bytesThis := this.(*object.BytesFile)
	if len(args) != 2 {
		return newError("type error: read_at requires two arguments " +
			"(the position and the size of the data to read)")
	}

	position, isInt := args[1].(*object.Integer)
	if !isInt || position.Value < 0 {
		return newError("type error: position must be a positive integer")
	}

	size, isInt := args[1].(*object.Integer)
	if !isInt || size.Value < 0 {
		return newError("type error: size must be a positive integer")
	}

	readData, err := bytesThis.Bytes.ReadAt(int(position.Value), int(size.Value))
	if err != nil {
		return newError("bytes error: bytes.read_at(%d, %d): %s",
			position.Value, size.Value, err)
	}
	retVal := &object.Array{Elements: make([]object.Object, len(readData))}
	for idx, readByte := range readData {
		retVal.Elements[idx] = &object.Integer{Value: int64(readByte)}
	}
	return retVal
}
