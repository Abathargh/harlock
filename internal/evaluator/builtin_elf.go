package evaluator

import "github.com/Abathargh/harlock/internal/object"

func elfBuiltinHasSection(this object.Object, args ...object.Object) object.Object {
	elfThis := this.(*object.ElfFile)
	section := args[0].(*object.String)
	if elfThis.File.HasSection(section.Value) {
		return TRUE
	}
	return FALSE
}

func elfBuiltinSections(this object.Object, _ ...object.Object) object.Object {
	elfThis := this.(*object.ElfFile)
	sections := elfThis.File.Sections()
	retVal := &object.Array{Elements: make([]object.Object, len(sections))}
	for idx, section := range sections {
		retVal.Elements[idx] = &object.String{Value: section}
	}
	return retVal
}

func elfBuiltinWriteSection(this object.Object, args ...object.Object) object.Object {
	elfThis := this.(*object.ElfFile)
	section := args[0].(*object.String)
	data := args[1].(*object.Array)

	offset := args[2].(*object.Integer)
	if offset.Value < 0 {
		return newError("type error: the offset must be a positive integer")
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

	err := elfThis.File.WriteSection(section.Value, byteArr, uint64(offset.Value))
	if err != nil {
		return newError("elf error: elf.write_section(%q, [%d]], %d): %s",
			section.Value, len(byteArr), uint64(offset.Value), err)
	}
	return nil
}

func elfBuiltinReadSection(this object.Object, args ...object.Object) object.Object {
	elfThis := this.(*object.ElfFile)
	section := args[0].(*object.String)

	readData, err := elfThis.File.ReadSection(section.Value)
	if err != nil {
		return newError("elf error: elf.read_section(%q): %s",
			section.Value, err)
	}

	retVal := &object.Array{Elements: make([]object.Object, len(readData))}
	for idx, readByte := range readData {
		retVal.Elements[idx] = &object.Integer{Value: int64(readByte)}
	}
	return retVal
}
