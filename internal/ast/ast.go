package ast

import (
	"fmt"
	"strings"

	"github.com/Abathargh/harlock/internal/token"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (program *Program) TokenLiteral() string {
	if len(program.Statements) > 0 {
		return program.Statements[0].TokenLiteral()
	}
	return ""
}

func (program *Program) String() string {
	var buf strings.Builder
	for _, statement := range program.Statements {
		buf.WriteString(statement.String())
	}
	return buf.String()
}

type Identifier struct {
	Token token.Token
	Value string
}

func (id *Identifier) expressionNode() {}

func (id *Identifier) TokenLiteral() string {
	return id.Token.Literal
}

func (id *Identifier) String() string {
	return id.Value
}

type VarStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (vs *VarStatement) statementNode() {}

func (vs *VarStatement) TokenLiteral() string {
	return vs.Token.Literal
}

func (vs *VarStatement) String() string {
	var buf strings.Builder
	buf.WriteString(vs.TokenLiteral() + " ")
	buf.WriteString(vs.Name.String())
	buf.WriteString(" = ")

	if vs.Value != nil {
		buf.WriteString(vs.Value.String())
	}
	return buf.String()
}

type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode() {}

func (rs *ReturnStatement) TokenLiteral() string {
	return rs.Token.Literal
}

func (rs *ReturnStatement) String() string {
	var buf strings.Builder
	buf.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		buf.WriteString(rs.ReturnValue.String())
	}
	return buf.String()
}

type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode() {}

func (es *ExpressionStatement) TokenLiteral() string {
	return es.Token.Literal
}

func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode() {}

func (il *IntegerLiteral) TokenLiteral() string {
	return il.Token.Literal
}

func (il *IntegerLiteral) String() string {
	return il.Token.Literal
}

type PrefixExpression struct {
	Token           token.Token
	Operator        string
	RightExpression Expression
}

func (pe *PrefixExpression) expressionNode() {}

func (pe *PrefixExpression) TokenLiteral() string {
	return pe.Token.Literal
}

func (pe *PrefixExpression) String() string {
	var buf strings.Builder
	buf.WriteString("(")
	buf.WriteString(pe.Operator)
	buf.WriteString(pe.RightExpression.String())
	buf.WriteString(")")
	return buf.String()
}

type InfixExpression struct {
	Token           token.Token
	LeftExpression  Expression
	Operator        string
	RightExpression Expression
}

func (ie *InfixExpression) expressionNode() {}

func (ie *InfixExpression) TokenLiteral() string {
	return ie.Token.Literal
}

func (ie *InfixExpression) String() string {
	var buf strings.Builder
	buf.WriteString("(")
	buf.WriteString(ie.LeftExpression.String())
	buf.WriteString(ie.Operator)
	buf.WriteString(ie.RightExpression.String())
	buf.WriteString(")")
	return buf.String()
}

type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode() {}

func (b *Boolean) TokenLiteral() string {
	return b.Token.Literal
}

func (b *Boolean) String() string {
	return b.Token.Literal
}

type IfExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ife *IfExpression) expressionNode() {}

func (ife *IfExpression) TokenLiteral() string {
	return ife.Token.Literal
}

func (ife *IfExpression) String() string {
	var buf strings.Builder
	buf.WriteString("if")
	buf.WriteString(ife.Condition.String())
	buf.WriteString(" {\n")
	buf.WriteString(ife.Consequence.String())
	buf.WriteString("\n}")

	if ife.Alternative != nil {
		buf.WriteString("else ")
		buf.WriteString(ife.Alternative.String())
	}
	return buf.String()
}

type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode() {}

func (bs *BlockStatement) TokenLiteral() string {
	return bs.Token.Literal
}

func (bs *BlockStatement) String() string {
	var buf strings.Builder
	for _, statement := range bs.Statements {
		buf.WriteString(statement.String())
	}
	return buf.String()
}

type FunctionLiteral struct {
	Token      token.Token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode() {}

func (fl *FunctionLiteral) TokenLiteral() string {
	return fl.Token.Literal
}

func (fl *FunctionLiteral) String() string {
	var buf strings.Builder
	var parameters []string

	for _, param := range fl.Parameters {
		parameters = append(parameters, param.String())
	}

	buf.WriteString(fl.TokenLiteral())
	buf.WriteString("(")
	buf.WriteString(strings.Join(parameters, ", "))
	buf.WriteString(")")
	buf.WriteString(fl.Body.String())

	return buf.String()
}

type CallExpression struct {
	Token token.Token
	// this can either be an identifier e.g. fun1()
	// or a func literal e.g. fun(a){ a }(12)
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode() {}

func (ce *CallExpression) TokenLiteral() string {
	return ce.Token.Literal
}

func (ce *CallExpression) String() string {
	var buf strings.Builder
	var parameters []string
	for _, param := range ce.Arguments {
		parameters = append(parameters, param.String())
	}

	buf.WriteString(ce.Function.String())
	buf.WriteString("(")
	buf.WriteString(strings.Join(parameters, ", "))
	buf.WriteString(")")
	return buf.String()
}

type NoOp struct {
	Token token.Token
}

func (no *NoOp) statementNode() {}

func (no *NoOp) TokenLiteral() string {
	return no.Token.Literal
}

func (no *NoOp) String() string {
	return "NOOP"
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode() {}

func (sl *StringLiteral) TokenLiteral() string {
	return sl.Token.Literal
}

func (sl *StringLiteral) String() string {
	return sl.Token.Literal
}

type ArrayLiteral struct {
	Token    token.Token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode() {}

func (al *ArrayLiteral) TokenLiteral() string {
	return al.Token.Literal
}

func (al *ArrayLiteral) String() string {
	var buf strings.Builder
	var elements []string
	for _, elem := range al.Elements {
		elements = append(elements, elem.String())
	}

	buf.WriteString("[")
	buf.WriteString(strings.Join(elements, ", "))
	buf.WriteString("]")
	return buf.String()
}

type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode() {}

func (ie *IndexExpression) TokenLiteral() string {
	return ie.Token.Literal
}

func (ie *IndexExpression) String() string {
	var buf strings.Builder
	buf.WriteString(ie.Left.String())
	buf.WriteString("[")
	buf.WriteString(ie.Index.String())
	buf.WriteString("]")
	return buf.String()
}

type MapLiteral struct {
	Token    token.Token
	Mappings map[Expression]Expression
}

func (hl *MapLiteral) expressionNode() {}

func (hl *MapLiteral) TokenLiteral() string {
	return hl.Token.Literal
}

func (hl *MapLiteral) String() string {
	var buf strings.Builder
	var mappings []string
	for key, val := range hl.Mappings {
		mappings = append(mappings, fmt.Sprintf("%s: %s", key.String(), val.String()))
	}

	buf.WriteString("{")
	buf.WriteString(strings.Join(mappings, ", "))
	buf.WriteString("}")
	return buf.String()
}

type MethodCallExpression struct {
	Token  token.Token
	Caller Expression
	Called *CallExpression
}

func (ml *MethodCallExpression) expressionNode() {}

func (ml *MethodCallExpression) TokenLiteral() string {
	return ml.Token.Literal
}

func (ml *MethodCallExpression) String() string {
	var buf strings.Builder
	var methodArgs []string
	for _, arg := range ml.Called.Arguments {
		methodArgs = append(methodArgs, arg.String())
	}

	buf.WriteString(ml.Caller.String())
	buf.WriteString(".")
	buf.WriteString(ml.Called.String())
	return buf.String()
}

type TryExpression struct {
	Token      token.Token
	Expression Expression
}

func (te *TryExpression) expressionNode() {}

func (te *TryExpression) TokenLiteral() string {
	return te.Token.Literal
}

func (te *TryExpression) String() string {
	var buf strings.Builder
	buf.WriteString("try ")
	buf.WriteString(te.Expression.String())
	return buf.String()
}
