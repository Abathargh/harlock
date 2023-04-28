package evaluator

import "github.com/Abathargh/harlock/internal/object"

const (
	maxByte = (1 << 8) - 1
)

func hexBuiltinRecord(this object.Object, args ...object.Object) object.Object {
	hexThis := this.(*object.HexFile)

	idx := args[0].(*object.Integer)
	readData, err := hexThis.File.Record(int(idx.Value))
	if err != nil {
		return newError("hex error: %s", err.Error())
	}
	return &object.String{Value: readData.AsString()}
}

func hexBuiltinSize(this object.Object, args ...object.Object) object.Object {
	hexThis := this.(*object.HexFile)
	size := hexThis.File.Size()
	return &object.Integer{Value: int64(size)}
}

func hexBuiltinBinarySize(this object.Object, args ...object.Object) object.Object {
	hexThis := this.(*object.HexFile)
	size := hexThis.File.BinarySize()
	return &object.Integer{Value: int64(size)}
}

func hexBuiltinReadAt(this object.Object, args ...object.Object) object.Object {
	hexThis := this.(*object.HexFile)

	pos := args[0].(*object.Integer)
	size := args[1].(*object.Integer)
	if pos.Value < 0 || size.Value < 0 {
		return newError("type error: position and size must be positive integers")
	}

	readData, err := hexThis.File.ReadAt(uint32(pos.Value), int(size.Value))
	if err != nil {
		return newError("hex error: hex.ReadAt(%d, %d): %s",
			uint32(pos.Value), int(size.Value), err)
	}

	retVal := &object.Array{Elements: make([]object.Object, len(readData))}
	for idx, readByte := range readData {
		retVal.Elements[idx] = &object.Integer{Value: int64(readByte)}
	}
	return retVal
}

func hexBuiltinWriteAt(this object.Object, args ...object.Object) object.Object {
	hexThis := this.(*object.HexFile)

	pos := args[0].(*object.Integer)
	data := args[1].(*object.Array)
	if pos.Value < 0 {
		return newError("type error: address must be a positive integer")
	}

	byteArr := make([]byte, len(data.Elements))
	for idx, elem := range data.Elements {
		intElem, isInt := elem.(*object.Integer)
		if !isInt || intElem.Value > maxByte || intElem.Value < 0 {
			return newError("type error: data must be an array of 1 byte positive integers")
		}
		byteArr[idx] = byte(intElem.Value)
	}

	err := hexThis.File.WriteAt(uint32(pos.Value), byteArr)
	if err != nil {
		return newError("hex error: %s", err)
	}
	return nil
}
