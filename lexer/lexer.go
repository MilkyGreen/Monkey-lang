package lexer

import (
	"monkey/token"
)

// Lexer 对象，负责将字符串转换成token
type Lexer struct {
	input        string // 要解析的代码字符串
	position     int    // 当前字符位置
	readPosition int    // 下一个字符位置
	ch           byte   // 当前的字符串
}

/*
*

	新建一个Lexer对象引用
*/
func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

// 读取下一个字符
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

// 将Lexer当前的字符串转换为token
func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.skipWhitespace()
	switch l.ch {
	case '=':
		tok = newToken(token.ASSIGN, l.ch)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		}else if isDigit(l.ch){
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}
	l.readChar()
	return tok

}

// 创建一个token
func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

// 是否是字母，下划线'_'也算做字母了，可以作为变量
func isLetter(ch byte) bool{
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func (l *Lexer) readIdentifier() string{
	position := l.position
	for isLetter(l.ch){
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace(){
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r'{
		l.readChar()
	}
}

func (l *Lexer) readNumber() string {
	position := l.position
	for(isDigit(l.ch)){
		l.readChar()
	}
	return l.input[position:l.position]
}

func isDigit(ch byte) bool{
	return '0' <= ch && ch <= '9'
}