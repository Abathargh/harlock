package object

import (
	"fmt"
	"strings"

	"github.com/Abathargh/harlock/internal/ast"
)

type ObjectType string

const (
	NullObj        = "NULL"
	ErrorObj       = "ERROR"
	IntegerObj     = "INTEGER"
	BooleanObj     = "BOOLEAN"
	FunctionObj    = "FUNCTION"
	ReturnValueObj = "RETURN_VALUE"
)

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
