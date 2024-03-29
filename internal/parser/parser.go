package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Abathargh/harlock/internal/ast"
	"github.com/Abathargh/harlock/internal/lexer"
	"github.com/Abathargh/harlock/internal/token"
)

type Priority int

const (
	LOWEST Priority = iota + 1
	LOGICAL
	EQUALS
	LESSGREATER
	OR
	AND
	SHIFT
	SUM
	PRODUCT
	PREFIX
	METHOD
	CALL
	INDEX
)

var priorities = map[token.TokenType]Priority{
	token.LOGICOR:   LOGICAL,
	token.LOGICAND:  LOGICAL,
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
	token.PERIOD:    METHOD,
	token.LPAREN:    CALL,
	token.LBRACK:    INDEX,
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
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)

	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.TRY, p.parseTryExpression)

	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)

	p.registerPrefix(token.STR, p.parseStringLiteral)

	p.registerPrefix(token.LBRACK, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseMapLiteral)

	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.RPAREN, p.parseGroupedExpression)

	p.registerPrefix(token.NOT, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.REV, p.parsePrefixExpression)

	p.registerInfix(token.PERIOD, p.parseMethodExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACK, p.parseIndexExpression)

	p.registerInfix(token.LOGICOR, p.parseInfixExpression)
	p.registerInfix(token.LOGICAND, p.parseInfixExpression)
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
	case token.NEWLINE:
		return parser.parseNewlineRow()
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
		LineMetadata: ast.LineMetadata{LineNumber: parser.lex.GetLineNumber()},
		Token:        parser.current,
		Value:        parser.current.Literal,
	}

	if !parser.expectPeek(token.ASSIGN) {
		return nil
	}

	parser.nextToken()
	statement.Value = parser.parseExpression(LOWEST)
	for parser.current.Type != token.NEWLINE && parser.current.Type != token.EOF {
		parser.nextToken()
	}
	return statement
}

func (parser *Parser) parseReturnStatement() *ast.ReturnStatement {
	statement := &ast.ReturnStatement{
		LineMetadata: ast.LineMetadata{LineNumber: parser.lex.GetLineNumber()},
		Token:        parser.current,
	}

	if parser.peeked.Type == token.NEWLINE || parser.peeked.Type == token.RBRACE {
		statement.ReturnValue = nil
		return statement
	}

	parser.nextToken()
	statement.ReturnValue = parser.parseExpression(LOWEST)
	for parser.current.Type != token.NEWLINE &&
		(parser.peeked.Type != token.RBRACE && parser.peeked.Type != token.NEWLINE) {
		if parser.current.Type == token.EOF {
			errMsg := fmt.Sprintf("unexpected %s on line %d", token.EOF, parser.lex.GetLineNumber())
			parser.errors = append(parser.errors, errMsg)
			return nil
		}
		parser.nextToken()
	}
	return statement
}

func (parser *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	statement := &ast.ExpressionStatement{
		LineMetadata: ast.LineMetadata{LineNumber: parser.lex.GetLineNumber()},
		Token:        parser.current,
	}

	statement.Expression = parser.parseExpression(LOWEST)

	if parser.peeked.Type == token.IDENT {
		parser.invalidExpressionError(parser.current, parser.peeked)
		return nil
	}

	if parser.peeked.Type == token.NEWLINE {
		parser.nextToken()
	}
	return statement
}

func (parser *Parser) parseExpression(prio Priority) ast.Expression {
	prefix := parser.prefixParseFns[parser.current.Type]
	if prefix == nil {
		parser.noPrefixParseFunctionError(parser.current)
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
	return &ast.Identifier{
		LineMetadata: ast.LineMetadata{LineNumber: parser.lex.GetLineNumber()},
		Token:        parser.current, Value: parser.current.Literal,
	}
}

func (parser *Parser) parseIntegerLiteral() ast.Expression {
	var value int64
	var err error
	literal := &ast.IntegerLiteral{
		LineMetadata: ast.LineMetadata{LineNumber: parser.lex.GetLineNumber()},
		Token:        parser.current,
	}
	if strings.HasPrefix(parser.current.Literal, "0x") ||
		strings.HasPrefix(parser.current.Literal, "0X") {
		value, err = strconv.ParseInt(parser.current.Literal[2:], 16, 64)
	} else {
		value, err = strconv.ParseInt(parser.current.Literal, 0, 64)
	}
	if err != nil {
		errMsg := fmt.Sprintf("%q could not be parsed as an integer, on line %d", parser.current.Literal,
			parser.lex.GetLineNumber())
		parser.errors = append(parser.errors, errMsg)
		return nil
	}
	literal.Value = value
	return literal
}

func (parser *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{
		LineMetadata: ast.LineMetadata{LineNumber: parser.lex.GetLineNumber()},
		Token:        parser.current,
		Value:        parser.current.Type == token.TRUE,
	}
}

func (parser *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{
		LineMetadata: ast.LineMetadata{LineNumber: parser.lex.GetLineNumber()},
		Token:        parser.current,
		Value:        parser.current.Literal,
	}
}

func (parser *Parser) parseArrayLiteral() ast.Expression {
	return &ast.ArrayLiteral{
		LineMetadata: ast.LineMetadata{LineNumber: parser.lex.GetLineNumber()},
		Token:        parser.current,
		Elements:     parser.parseExpressionList(token.RBRACK),
	}
}

func (parser *Parser) parseMapLiteral() ast.Expression {
	mapLiteral := &ast.MapLiteral{
		LineMetadata: ast.LineMetadata{LineNumber: parser.lex.GetLineNumber()},
		Token:        parser.current,
		Mappings:     make(map[ast.Expression]ast.Expression),
	}

	for parser.peeked.Type != token.RBRACE {
		if !parser.skipNewline() {
			errMsg := fmt.Sprintf("unexpected %s on line %d", token.EOF, parser.lex.GetLineNumber())
			parser.errors = append(parser.errors, errMsg)
			return nil
		}

		parser.nextToken()
		currentKey := parser.parseExpression(LOWEST)
		if !parser.expectPeek(token.COLON) {
			return nil
		}

		parser.nextToken()
		currentVal := parser.parseExpression(LOWEST)
		mapLiteral.Mappings[currentKey] = currentVal
		if (parser.peeked.Type != token.RBRACE && !parser.expectPeek(token.COMMA)) || !parser.skipNewline() {
			return nil
		}
	}
	if !parser.expectPeek(token.RBRACE) {
		return nil
	}
	return mapLiteral
}

func (parser *Parser) parseNewlineRow() ast.Statement {
	return nil
}

func (parser *Parser) parseGroupedExpression() ast.Expression {
	parser.nextToken()
	expression := parser.parseExpression(LOWEST)
	if !parser.expectPeek(token.RPAREN) {
		return nil
	}
	return expression
}

func (parser *Parser) parseIfExpression() ast.Expression {
	// TODO modify AST for if and this to allow else if
	expression := &ast.IfExpression{
		LineMetadata: ast.LineMetadata{LineNumber: parser.lex.GetLineNumber()},
		Token:        parser.current,
	}

	parser.nextToken()
	expression.Condition = parser.parseExpression(LOWEST)
	if !parser.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = parser.parseBlockStatement()
	if parser.peeked.Type == token.ELSE {
		parser.nextToken()
		if !parser.expectPeek(token.LBRACE) {
			return nil
		}
		expression.Alternative = parser.parseBlockStatement()
	}
	return expression
}

func (parser *Parser) parseTryExpression() ast.Expression {
	tryExpression := &ast.TryExpression{
		LineMetadata: ast.LineMetadata{LineNumber: parser.lex.GetLineNumber()},
		Token:        parser.current,
	}
	parser.nextToken()
	tryExpression.Expression = parser.parseExpression(LOWEST)
	return tryExpression
}

func (parser *Parser) parseFunctionLiteral() ast.Expression {
	functionLiteral := &ast.FunctionLiteral{
		LineMetadata: ast.LineMetadata{LineNumber: parser.lex.GetLineNumber()},
		Token:        parser.current,
	}
	if !parser.expectPeek(token.LPAREN) {
		return nil
	}

	functionLiteral.Parameters = parser.parseFunctionParameters()
	if !parser.expectPeek(token.LBRACE) {
		return nil
	}
	functionLiteral.Body = parser.parseBlockStatement()
	return functionLiteral
}

func (parser *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	callExpression := &ast.CallExpression{
		LineMetadata: ast.LineMetadata{LineNumber: parser.lex.GetLineNumber()},
		Token:        parser.current,
		Function:     function,
	}
	callExpression.Arguments = parser.parseExpressionList(token.RPAREN)

	return callExpression
}

func (parser *Parser) parseMethodExpression(caller ast.Expression) ast.Expression {
	methodExpression := &ast.MethodCallExpression{
		LineMetadata: ast.LineMetadata{LineNumber: parser.lex.GetLineNumber()},
		Token:        parser.current,
		Caller:       caller,
	}
	if !parser.expectPeek(token.IDENT) {
		return nil
	}
	methodName := parser.parseIdentifier()
	if !parser.expectPeek(token.LPAREN) {
		return nil
	}

	methodExpression.Called = parser.parseCallExpression(methodName).(*ast.CallExpression)
	return methodExpression
}

func (parser *Parser) parseIndexExpression(array ast.Expression) ast.Expression {
	indexExpression := &ast.IndexExpression{
		LineMetadata: ast.LineMetadata{LineNumber: parser.lex.GetLineNumber()},
		Token:        parser.current,
		Left:         array,
	}
	parser.nextToken()
	indexExpression.Index = parser.parseExpression(LOWEST)

	if !parser.expectPeek(token.RBRACK) {
		return nil
	}
	return indexExpression
}

func (parser *Parser) parsePrefixExpression() ast.Expression {
	prefixExpression := &ast.PrefixExpression{
		LineMetadata: ast.LineMetadata{LineNumber: parser.lex.GetLineNumber()},
		Token:        parser.current,
		Operator:     parser.current.Literal,
	}

	parser.nextToken()
	prefixExpression.RightExpression = parser.parseExpression(PREFIX)
	return prefixExpression
}

func (parser *Parser) parseInfixExpression(leftExpression ast.Expression) ast.Expression {
	infixExpression := &ast.InfixExpression{
		LineMetadata:   ast.LineMetadata{LineNumber: parser.lex.GetLineNumber()},
		Token:          parser.current,
		LeftExpression: leftExpression,
		Operator:       parser.current.Literal,
	}
	prio := parser.currentPrecedence()
	parser.nextToken()
	infixExpression.RightExpression = parser.parseExpression(prio)
	return infixExpression
}

func (parser *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: parser.current}
	parser.nextToken()

	for parser.current.Type != token.RBRACE {
		if parser.current.Type == token.EOF {
			errMsg := fmt.Sprintf("expected %s, got %s on line %d", token.RBRACE, token.EOF,
				parser.lex.GetLineNumber())
			parser.errors = append(parser.errors, errMsg)
			return nil
		}
		statement := parser.parseStatement()
		if statement != nil {
			block.Statements = append(block.Statements, statement)
		}
		parser.nextToken()
	}
	return block
}

func (parser *Parser) parseFunctionParameters() []*ast.Identifier {
	var parameters []*ast.Identifier

	if parser.peeked.Type == token.RPAREN {
		parser.nextToken()
		return parameters
	}

	parser.nextToken()
	parameter := &ast.Identifier{
		LineMetadata: ast.LineMetadata{LineNumber: parser.lex.GetLineNumber()},
		Token:        parser.current,
		Value:        parser.current.Literal,
	}
	parameters = append(parameters, parameter)

	for parser.peeked.Type == token.COMMA {
		parser.nextToken()
		parser.nextToken()
		parameter = &ast.Identifier{Token: parser.current, Value: parser.current.Literal}
		parameters = append(parameters, parameter)
	}

	if !parser.expectPeek(token.RPAREN) {
		return nil
	}
	return parameters
}

func (parser *Parser) parseExpressionList(terminator token.TokenType) []ast.Expression {
	var parameters []ast.Expression
	if parser.peeked.Type == terminator {
		parser.nextToken()
		return parameters
	}

	parser.nextToken()
	parameters = append(parameters, parser.parseExpression(LOWEST))
	for parser.peeked.Type == token.COMMA {
		parser.nextToken()
		parser.nextToken()
		parameters = append(parameters, parser.parseExpression(LOWEST))
	}

	if !parser.expectPeek(terminator) {
		return nil
	}
	return parameters
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
	errMsg := fmt.Sprintf("expected token of type %q, got %q on line %d", t, parser.peeked.Type,
		parser.lex.GetLineNumber())
	parser.errors = append(parser.errors, errMsg)
}

func (parser *Parser) noPrefixParseFunctionError(t token.Token) {
	errMsg := fmt.Sprintf("cannot parse: prefix operator %q on line %d", t.Literal, parser.lex.GetLineNumber())
	parser.errors = append(parser.errors, errMsg)
}

func (parser *Parser) invalidExpressionError(t token.Token, p token.Token) {
	errMsg := fmt.Sprintf("cannot parse: invalid expression \"%s%s\" on line %d", t.Literal, p.Literal,
		parser.lex.GetLineNumber())
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

func (parser *Parser) skipNewline() bool {
	for parser.peeked.Type == token.NEWLINE {
		if parser.peeked.Type == token.EOF {
			return false
		}
		parser.nextToken()
	}
	return true
}
