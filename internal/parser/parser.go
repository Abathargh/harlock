package parser

import (
	"fmt"
	"strconv"

	"github.com/Abathargh/harlock/internal/ast"
	"github.com/Abathargh/harlock/internal/lexer"
	"github.com/Abathargh/harlock/internal/token"
)

type Priority int

const (
	LOWEST Priority = iota + 1
	EQUALS
	LESSGREATER
	OR
	AND
	SHIFT
	SUM
	PRODUCT
	PREFIX
	CALL
)

var priorities = map[token.TokenType]Priority{
	token.EQUALS:    EQUALS,
	token.NOTEQUALS: EQUALS,
	token.LESS:      LESSGREATER,
	token.LESSEQ:    LESSGREATER,
	token.GREATER:   LESSGREATER,
	token.GREATEREQ: LESSGREATER,
	token.OR:        OR,
	token.XOR:       OR,
	token.AND:       AND,
	token.LSHIFT:    SHIFT,
	token.RSHIFT:    SHIFT,
	token.PLUS:      SUM,
	token.MINUS:     SUM,
	token.MUL:       PRODUCT,
	token.DIV:       PRODUCT,
	token.MOD:       PRODUCT,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(expression ast.Expression) ast.Expression
)

type Parser struct {
	lex    *lexer.Lexer
	errors []string

	current token.Token
	peeked  token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func NewParser(lex *lexer.Lexer) *Parser {
	p := &Parser{lex: lex}
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.infixParseFns = make(map[token.TokenType]infixParseFn)

	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)

	p.registerPrefix(token.NOT, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.REV, p.parsePrefixExpression)

	p.registerInfix(token.EQUALS, p.parseInfixExpression)
	p.registerInfix(token.NOTEQUALS, p.parseInfixExpression)
	p.registerInfix(token.LESS, p.parseInfixExpression)
	p.registerInfix(token.LESSEQ, p.parseInfixExpression)
	p.registerInfix(token.GREATER, p.parseInfixExpression)
	p.registerInfix(token.GREATEREQ, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.XOR, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.LSHIFT, p.parseInfixExpression)
	p.registerInfix(token.RSHIFT, p.parseInfixExpression)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.MUL, p.parseInfixExpression)
	p.registerInfix(token.DIV, p.parseInfixExpression)
	p.registerInfix(token.MOD, p.parseInfixExpression)

	p.nextToken()
	p.nextToken()
	return p
}

func (parser *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	for parser.current.Type != token.EOF {
		statement := parser.parseStatement()
		if statement != nil {
			program.Statements = append(program.Statements, statement)
		}
		parser.nextToken()
	}
	return program
}

func (parser *Parser) Errors() []string {
	return parser.errors
}

func (parser *Parser) parseStatement() ast.Statement {
	switch parser.current.Type {
	case token.VAR:
		return parser.parseVarStatement()
	case token.RET:
		return parser.parseReturnStatement()
	default:
		return parser.parseExpressionStatement()
	}
}

func (parser *Parser) parseVarStatement() *ast.VarStatement {
	statement := &ast.VarStatement{Token: parser.current}
	if !parser.expectPeek(token.IDENT) {
		return nil
	}

	statement.Name = &ast.Identifier{
		Token: parser.current,
		Value: parser.current.Literal,
	}

	if !parser.expectPeek(token.ASSIGN) {
		return nil
	}

	// TODO rest of expr
	for parser.current.Type != token.NEWLINE {
		parser.nextToken()
	}
	return statement
}

func (parser *Parser) parseReturnStatement() *ast.ReturnStatement {
	statement := &ast.ReturnStatement{Token: parser.current}
	parser.nextToken()

	// TODO rest of expr
	for parser.current.Type != token.NEWLINE {
		parser.nextToken()
	}
	return statement
}

func (parser *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	statement := &ast.ExpressionStatement{Token: parser.current}

	statement.Expression = parser.parseExpression(LOWEST)

	if parser.peeked.Type == token.NEWLINE {
		parser.nextToken()
	}
	return statement
}

func (parser *Parser) parseExpression(prio Priority) ast.Expression {
	prefix := parser.prefixParseFns[parser.current.Type]
	if prefix == nil {
		parser.noPrefixParseFunctionError(parser.current.Type)
		return nil
	}

	leftExpression := prefix()

	for parser.peeked.Type != token.NEWLINE && prio < parser.peekPrecedence() {
		infix := parser.infixParseFns[parser.peeked.Type]
		if infix == nil {
			return leftExpression
		}
		parser.nextToken()
		leftExpression = infix(leftExpression)
	}

	return leftExpression
}

func (parser *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: parser.current, Value: parser.current.Literal}
}

func (parser *Parser) parseIntegerLiteral() ast.Expression {
	literal := &ast.IntegerLiteral{Token: parser.current}
	value, err := strconv.ParseInt(parser.current.Literal, 0, 64)
	if err != nil {
		errMsg := fmt.Sprintf("%q could not be parsed as an integer", parser.current.Literal)
		parser.errors = append(parser.errors, errMsg)
		return nil
	}
	literal.Value = value
	return literal
}

func (parser *Parser) parsePrefixExpression() ast.Expression {
	prefixExpression := &ast.PrefixExpression{
		Token:    parser.current,
		Operator: parser.current.Literal,
	}

	parser.nextToken()
	prefixExpression.RightExpression = parser.parseExpression(PREFIX)
	return prefixExpression
}

func (parser *Parser) parseInfixExpression(leftExpression ast.Expression) ast.Expression {
	infixExpression := &ast.InfixExpression{
		Token:          parser.current,
		LeftExpression: leftExpression,
		Operator:       parser.current.Literal,
	}
	prio := parser.currentPrecedence()
	parser.nextToken()
	infixExpression.RightExpression = parser.parseExpression(prio)
	return infixExpression
}

func (parser *Parser) expectPeek(t token.TokenType) bool {
	if parser.peeked.Type == t {
		parser.nextToken()
		return true
	}
	parser.peekError(t)
	return false
}

func (parser *Parser) currentPrecedence() Priority {
	if prio, ok := priorities[parser.current.Type]; ok {
		return prio
	}
	return LOWEST
}

func (parser *Parser) peekPrecedence() Priority {
	if prio, ok := priorities[parser.peeked.Type]; ok {
		return prio
	}
	return LOWEST
}

func (parser *Parser) peekError(t token.TokenType) {
	errMsg := fmt.Sprintf("expected token of type %s, got %s", t, parser.peeked.Type)
	parser.errors = append(parser.errors, errMsg)
}

func (parser *Parser) noPrefixParseFunctionError(t token.TokenType) {
	errMsg := fmt.Sprintf("cannot parse prefix operator %s", t)
	parser.errors = append(parser.errors, errMsg)
}

func (parser *Parser) nextToken() {
	parser.current = parser.peeked
	parser.peeked = parser.lex.NextToken()
}

func (parser *Parser) registerPrefix(t token.TokenType, fn prefixParseFn) {
	parser.prefixParseFns[t] = fn
}

func (parser *Parser) registerInfix(t token.TokenType, fn infixParseFn) {
	parser.infixParseFns[t] = fn
}
