package object

import (
	"fmt"
	"strings"

	"github.com/Abathargh/harlock/internal/ast"
)

type ObjectType string

const (
	NullObj        = "NULL"
	TypeObj        = "Type"
	ErrorObj       = "ERROR"
	StringObj      = "STRING"
	IntegerObj     = "INTEGER"
	BooleanObj     = "BOOLEAN"
	BuiltinObj     = "BUILTIN"
	FunctionObj    = "FUNCTION"
	ReturnValueObj = "RETURN_VALUE"
)

type BuiltinFunction func(args ...Object) Object

type Object interface {
	Type() ObjectType
	Inspect() string
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

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType {
	return BooleanObj
}

func (b *Boolean) Inspect() string {
	return fmt.Sprintf("%t", b.Value)
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
	// TODO add support for error line/column tracing (needs changes in lexer)
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
	buf.WriteString("\n}")
	return buf.String()
}

type String struct {
	Value string
}

func (str *String) Type() ObjectType {
	return StringObj
}

func (str *String) Inspect() string {
	return fmt.Sprintf("\"%s\"", str.Value)
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
