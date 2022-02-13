package evaluator

import "github.com/Abathargh/harlock/internal/object"

const (
	maxByte = (1 << 8) - 1
)

func hexBuiltinRecord(this object.Object, args ...object.Object) object.Object {
	hexThis := this.(*object.HexFile)
	if len(args) != 1 {
		return newError("type error: read_at requires one argument (the index of the record to read)")
	}

	idx, isInt := args[0].(*object.Integer)
	if !isInt || idx.Value < 0 {
		return newError("type error: index must be a positive integer")
	}

	readData := hexThis.File.Record(int(idx.Value))
	if readData == nil {
		return newError("hex error: invalid record index")
	}

	return &object.String{Value: readData.AsString()}
}

func hexBuiltinSize(this object.Object, args ...object.Object) object.Object {
	hexThis := this.(*object.HexFile)
	if len(args) != 0 {
		return newError("type error: size does not require any input arguments")
	}

	size := hexThis.File.Size()
	return &object.Integer{Value: int64(size)}
}

func hexBuiltinReadAt(this object.Object, args ...object.Object) object.Object {
	hexThis := this.(*object.HexFile)
	if len(args) != 2 {
		return newError("type error: read_at requires two arguments (the address and the size)")
	}

	pos, isInt := args[0].(*object.Integer)
	if !isInt || pos.Value < 0 {
		return newError("type error: position must be a positive integer")
	}

	size, isInt := args[1].(*object.Integer)
	if !isInt {
		return newError("type error: size must be an integer")
	}

	readData, err := hexThis.File.ReadAt(uint32(pos.Value), int(size.Value))
	if err != nil {
		return newError("hex error: %s", err)
	}

	retVal := &object.Array{Elements: make([]object.Object, len(readData))}
	for idx, readByte := range readData {
		retVal.Elements[idx] = &object.Integer{Value: int64(readByte)}
	}
	return retVal
}

func hexBuiltinWriteAt(this object.Object, args ...object.Object) object.Object {
	hexThis := this.(*object.HexFile)
	if len(args) != 2 {
		return newError("type error: write_at requires two arguments (the address and the data)")
	}

	pos, isInt := args[0].(*object.Integer)
	if !isInt || pos.Value < 0 {
		return newError("type error: address must be a positive integer")
	}

	data, isArr := args[1].(*object.Array)
	if !isArr {
		return newError("type error: data must be an array")
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
