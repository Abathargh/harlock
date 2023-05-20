package evaluator

import "github.com/Abathargh/harlock/internal/object"

func bytesBuiltinWriteAt(this object.Object, args ...object.Object) object.Object {
	bytesThis := this.(*object.BytesFile)

	position := args[0].(*object.Integer)
	data := args[1].(*object.Array)
	if position.Value < 0 {
		return newBytesError("position must be a positive integer")
	}

	byteArr := make([]byte, len(data.Elements))
	for idx, elem := range data.Elements {
		intElem, isInt := elem.(*object.Integer)
		if !isInt || intElem.Value > maxByte || intElem.Value < 0 {
			return newTypeError("data must be an array of 1 byte positive integers "+
				"(data[%d] = %s does not follow this constraint)", idx, elem.Inspect())
		}
		byteArr[idx] = byte(intElem.Value)
	}

	err := bytesThis.Bytes.WriteAt(int(position.Value), byteArr)
	if err != nil {
		return newBytesError("bytes.write_at([%d], %d): %s",
			len(byteArr), uint64(position.Value), err)
	}
	return nil
}

func bytesBuiltinReadAt(this object.Object, args ...object.Object) object.Object {
	bytesThis := this.(*object.BytesFile)

	position := args[0].(*object.Integer)
	size := args[1].(*object.Integer)
	if position.Value < 0 || size.Value < 0 {
		return newBytesError("position and size must be positive integers")
	}

	readData, err := bytesThis.Bytes.ReadAt(int(position.Value), int(size.Value))
	if err != nil {
		return newBytesError("bytes.read_at(%d, %d): %s",
			position.Value, size.Value, err)
	}
	retVal := &object.Array{Elements: make([]object.Object, len(readData))}
	for idx, readByte := range readData {
		retVal.Elements[idx] = &object.Integer{Value: int64(readByte)}
	}
	return retVal
}
