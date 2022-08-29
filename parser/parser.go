package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

// 操作符优先级，优先parse高级别的
const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
}

type (
	// 处理前缀表达式
	prefixParseFn func() ast.Expression
	// 处理中缀表达式
	infixParseFn func(ast.Expression) ast.Expression
)

type Parser struct {
	l *lexer.Lexer

	curToken  token.Token // 当前的token
	peekToken token.Token // 下一个token

	errors []string

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)

	p.registerPrefix(token.IDENT, p.parseIdentifier)

	p.registerPrefix(token.INT, p.parseIntegerLiteral)

	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)

	return p
}

// 读取下一个token
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// 将token parse成AST，返回根节点 Program
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{} // 创建了一个长度为0的数组，注意并不是nil
	// 不断的parse声明语句，直到结束
	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			// 加入到声明列表中
			program.Statements = append(program.Statements, stmt)
		}
		// 读取下一个token
		p.nextToken()
	}
	return program
}

// 从当前的token位置 parse一条声明
func (p *Parser) parseStatement() ast.Statement {
	defer untrace(trace("parseStatement"))
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parse let 语句
func (p *Parser) parseLetStatement() *ast.LetStatement {
	defer untrace(trace("parseLetStatement"))
	// 第一个token是 let
	stmt := &ast.LetStatement{Token: p.curToken}
	// 第二个必须是个变量名
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	// 变量名
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	// 后面一定要是等号
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	// TODO: We're skipping the expressions until we
	// encounter a semicolon
	// 这里暂时先把后面表达式的部分跳过
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// parse return 声明
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	defer untrace(trace("parseReturnStatement"))
	// 第一个token是 return
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()
	// TODO: We're skipping the expressions until we
	// encounter a semicolon
	// 这里暂时先把后面表达式的部分跳过
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// 当前token类似是否符合预期
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// 下一个token是否符合预期
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// 下一个token类型是否符合预期，符合的话读取下一个
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// parse 表达式类型的声明
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	defer untrace(trace("parseExpressionStatement"))
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// parse 表达式
// 先按前缀解析左边，再判断后面有没符合条件的中缀
func (p *Parser) parseExpression(precedence int) ast.Expression {
	defer untrace(trace("parseExpression"))
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	// 先用前缀方法解析左边的表达式
	leftExp := prefix()
	// 如果后面有级别更高的操作符，则按中缀解析，前面的表达式当做中缀表达式的一分部
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

// 解析Identifier表达式
func (p *Parser) parseIdentifier() ast.Expression {
	defer untrace(trace("parseIdentifier"))
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	defer untrace(trace("parseIntegerLiteral"))
	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

// parse 前缀表达式
func (p *Parser) parsePrefixExpression() ast.Expression {
	defer untrace(trace("parsePrefixExpression"))
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// parse 中缀表达式
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	defer untrace(trace("parseInfixExpression"))
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}
