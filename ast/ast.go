package ast

import (
	"bytes"
	"monkey/token"
)

// ast 节点
type Node interface {
	TokenLiteral() string
	String() string
}

// 声明接口
type Statement interface {
	Node // 一个结构里直接放入另一个结构，是继承
	statementNode()
}

// 表达式接口
type Expression interface {
	Node
	expressionNode()
}

// 抽象语法树的跟节点
type Program struct {
	// 程序由很多声明组成
	Statements []Statement
}

// 实现Node接口
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}
func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// let 声明
type LetStatement struct {
	Token token.Token // the token.LET token
	Name  *Identifier
	Value Expression
}
func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

// Identifier 表达式
type Identifier struct {
	Token token.Token // the token.IDENT token
	Value string
}
func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// return声明
type ReturnStatement struct {
	Token       token.Token // return token
	ReturnValue Expression  // 后面跟一个表达式
}
func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}

// 表达式类型的声明（表达式也是一种特殊的声明）
type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}
func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}