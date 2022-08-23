package parser

import (
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"fmt"
)

type Parser struct {
	l *lexer.Lexer

	curToken  token.Token // 当前的token
	peekToken token.Token // 下一个token

	errors []string
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()
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
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return nil
	}
}

// parse let 语句
func (p *Parser) parseLetStatement() *ast.LetStatement {
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
func (p *Parser) parseReturnStatement() *ast.ReturnStatement{
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
