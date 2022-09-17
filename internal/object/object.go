package object

import (
	"fmt"
	"github.com/Abathargh/harlock/internal/evaluator/bytes"
	"github.com/Abathargh/harlock/internal/evaluator/elf"
	"github.com/Abathargh/harlock/internal/evaluator/hex"
	"hash/fnv"
	"strconv"
	"strings"

	"github.com/Abathargh/harlock/internal/ast"
)

type ObjectType string

const (
	NullObj        = "Null"
	TypeObj        = "Type"
	SetObj         = "Set"
	MapObj         = "Map"
	HexObj         = "Hex File"
	ElfObj         = "Elf File"
	BytesObj       = "Bytes File"
	ErrorObj       = "Error"
	ArrayObj       = "Array"
	StringObj      = "String"
	MethodObj      = "Method"
	IntegerObj     = "Int"
	BooleanObj     = "Bool"
	BuiltinObj     = "Builtin Function"
	FunctionObj    = "Function"
	ReturnValueObj = "Return value"
)

type BuiltinFunction func(args ...Object) Object
type MethodFunction func(this Object, args ...Object) Object

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Hashable interface {
	HashKey() HashKey
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType {
	return IntegerObj
}

func (i *Integer) Inspect() string {
	return fmt.Sprintf("%d", i.Value)
}

func (i *Integer) HashKey() HashKey {
	return HashKey{Type: IntegerObj, Value: uint64(i.Value)}
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType {
	return BooleanObj
}

func (b *Boolean) Inspect() string {
	return fmt.Sprintf("%t", b.Value)
}

func (b *Boolean) HashKey() HashKey {
	if b.Value {
		return HashKey{Type: BooleanObj, Value: 1}
	}
	return HashKey{Type: BooleanObj, Value: 0}
}

type Null struct{}

func (n *Null) Type() ObjectType {
	return NullObj
}

func (n *Null) Inspect() string {
	return "null"
}

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType {
	return ReturnValueObj
}

func (rv *ReturnValue) Inspect() string {
	return rv.Value.Inspect()
}

type Error struct {
	// TODO add subtype info referring to a typed enum for different error values
	Message string
}

func (e *Error) Type() ObjectType {
	return ErrorObj
}

func (e *Error) Inspect() string {
	return fmt.Sprintf("Error: %s", e.Message)
}

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType {
	return FunctionObj
}

func (f *Function) Inspect() string {
	var buf strings.Builder
	var parameters []string

	for _, parameter := range f.Parameters {
		parameters = append(parameters, parameter.String())
	}

	buf.WriteString("fun(")
	buf.WriteString(strings.Join(parameters, ", "))
	buf.WriteString(") {\n")
	buf.WriteString(f.Body.String())
	buf.WriteString("}")
	return buf.String()
}

type String struct {
	Value string
}

func (str *String) Type() ObjectType {
	return StringObj
}

func (str *String) Inspect() string {
	return str.Value
}

func (str *String) HashKey() HashKey {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(str.Value))
	return HashKey{Type: StringObj, Value: hash.Sum64()}
}

type Builtin struct {
	Function BuiltinFunction
}

func (b *Builtin) Type() ObjectType {
	return BuiltinObj
}

func (b *Builtin) Inspect() string {
	return "builtin"
}

type Type struct {
	Value ObjectType
}

func (t *Type) Type() ObjectType {
	return TypeObj
}

func (t *Type) Inspect() string {
	return fmt.Sprintf("%s", t.Value)
}

type Array struct {
	Elements []Object
}

func (arr *Array) Type() ObjectType {
	return ArrayObj
}

func (arr *Array) Inspect() string {
	var buf strings.Builder
	var elements []string
	for _, element := range arr.Elements {
		elements = append(elements, element.Inspect())
	}

	buf.WriteString("[")
	buf.WriteString(strings.Join(elements, ", "))
	buf.WriteString("]")
	return buf.String()
}

type HashPair struct {
	Key   Object
	Value Object
}

type Map struct {
	Mappings map[HashKey]HashPair
}

func (h *Map) Type() ObjectType {
	return MapObj
}

func (h *Map) Inspect() string {
	var buf strings.Builder
	var mappings []string
	for _, mapping := range h.Mappings {

		mappings = append(mappings,
			fmt.Sprintf("%s: %s", mapping.Key.Inspect(), mapping.Value.Inspect()))
	}

	buf.WriteString("{")
	buf.WriteString(strings.Join(mappings, ", "))
	buf.WriteString("}")
	return buf.String()
}

type Method struct {
	MethodFunc MethodFunction
}

func (m *Method) Type() ObjectType {
	return MethodObj
}

func (m *Method) Inspect() string {
	return "builtin method"
}

type Set struct {
	Elements map[HashKey]Object
}

func (s *Set) Type() ObjectType {
	return SetObj
}

func (s *Set) Inspect() string {
	var buf strings.Builder
	var elements []string
	for _, mapping := range s.Elements {
		elements = append(elements, mapping.Inspect())
	}

	buf.WriteString("set(")
	buf.WriteString(strings.Join(elements, ", "))
	buf.WriteString(")")
	return buf.String()
}

type File interface {
	Name() string
	Perms() uint32
	AsBytes() []byte
}

type HexFile struct {
	name  string
	perms uint32
	File  *hex.File
}

func NewHexFile(name string, perms uint32, hexfile *hex.File) *HexFile {
	return &HexFile{
		name:  name,
		perms: perms,
		File:  hexfile,
	}
}

func (hf *HexFile) Name() string {
	return hf.name
}

func (hf *HexFile) Perms() uint32 {
	return hf.perms
}

func (hf *HexFile) AsBytes() []byte {
	var buf []byte
	ch := hf.File.Iterator()
	for rec := range ch {
		buf = append(buf, rec.AsBytes()...)
	}
	return buf
}

func (hf *HexFile) Type() ObjectType {
	return HexObj
}

func (hf *HexFile) Inspect() string {
	var buf strings.Builder
	var records []string

	ch := hf.File.Iterator()
	for rec := range ch {
		records = append(records, rec.AsString())
	}

	buf.WriteString(strings.Join(records, "\n"))
	return buf.String()
}

type ElfFile struct {
	name  string
	perms uint32
	File  *elf.File
}

func NewElfFile(name string, perms uint32, elffile *elf.File) *ElfFile {
	return &ElfFile{
		name:  name,
		perms: perms,
		File:  elffile,
	}
}

func (ef *ElfFile) Name() string {
	return ef.name
}

func (ef *ElfFile) Perms() uint32 {
	return ef.perms
}

func (ef *ElfFile) AsBytes() []byte {
	return ef.File.AsBytes()
}

func (ef *ElfFile) Type() ObjectType {
	return ElfObj
}

func (ef *ElfFile) Inspect() string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("ElfFile(@%s) {\n", ef.name))
	buf.WriteString("  Sections: [")
	for _, section := range ef.File.Sections() {
		buf.WriteString(fmt.Sprintf("%s ", section))
	}
	buf.WriteString("]\n")
	buf.WriteString("}")

	return buf.String()
}

type BytesFile struct {
	name  string
	perms uint32
	size  int64
	Bytes *bytes.File
}

func NewBytesFile(name string, perms uint32, size int64, bytesFile *bytes.File) *BytesFile {
	return &BytesFile{
		name:  name,
		perms: perms,
		size:  size,
		Bytes: bytesFile,
	}
}

func (bf *BytesFile) Name() string {
	return bf.name
}

func (bf *BytesFile) Perms() uint32 {
	return bf.perms
}

func (bf *BytesFile) AsBytes() []byte {
	data, _ := bf.Bytes.ReadAt(0, int(bf.size))
	return data
}

func (bf *BytesFile) Type() ObjectType {
	return BytesObj
}

func (bf *BytesFile) Inspect() string {
	var buf strings.Builder
	bs := bf.AsBytes()
	for idx, b := range bs {
		buf.WriteString(strconv.Itoa(int(b)))
		if idx != len(bs)-1 {
			buf.WriteString(", ")
		}
	}
	return buf.String()
}
