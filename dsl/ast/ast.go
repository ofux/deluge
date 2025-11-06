package ast

import (
	"bytes"
	"fmt"
	"github.com/ofux/deluge/dsl/token"
	"strings"
)

// The base Node interface
type Node interface {
	TokenDetails() token.Token
	TokenLiteral() string
	String() string
}

// All statement nodes implement this
type Statement interface {
	Node
	statementNode()
}

// All expression nodes implement this
type Expression interface {
	Node
	expressionNode()
}

func PrintLocation(node Node) string {
	return fmt.Sprintf("%s (line %d, col %d)", node.TokenLiteral(), node.TokenDetails().Line, node.TokenDetails().Column)
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenDetails() token.Token {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenDetails()
	} else {
		return token.Token{Type: token.EOF, Line: 1, Column: 1, Literal: token.EOF}
	}
}

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

// Statements
type LetStatement struct {
	Token token.Token // the token.LET token
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()            {}
func (ls *LetStatement) TokenDetails() token.Token { return ls.Token }
func (ls *LetStatement) TokenLiteral() string      { return ls.Token.Literal }
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

type ReturnStatement struct {
	Token       token.Token // the 'return' token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()            {}
func (rs *ReturnStatement) TokenDetails() token.Token { return rs.Token }
func (rs *ReturnStatement) TokenLiteral() string      { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}

type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()            {}
func (es *ExpressionStatement) TokenDetails() token.Token { return es.Token }
func (es *ExpressionStatement) TokenLiteral() string      { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type BlockStatement struct {
	Token      token.Token // the { token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()            {}
func (bs *BlockStatement) TokenDetails() token.Token { return bs.Token }
func (bs *BlockStatement) TokenLiteral() string      { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// Expressions
type Null struct {
	Token token.Token // the token.NULL token
}

func (i *Null) expressionNode()           {}
func (i *Null) TokenDetails() token.Token { return i.Token }
func (i *Null) TokenLiteral() string      { return i.Token.Literal }
func (i *Null) String() string            { return i.Token.Literal }

type Identifier struct {
	Token token.Token // the token.IDENT token
	Value string
}

func (i *Identifier) expressionNode()           {}
func (i *Identifier) TokenDetails() token.Token { return i.Token }
func (i *Identifier) TokenLiteral() string      { return i.Token.Literal }
func (i *Identifier) String() string            { return i.Value }

type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()           {}
func (b *Boolean) TokenDetails() token.Token { return b.Token }
func (b *Boolean) TokenLiteral() string      { return b.Token.Literal }
func (b *Boolean) String() string            { return b.Token.Literal }

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()           {}
func (il *IntegerLiteral) TokenDetails() token.Token { return il.Token }
func (il *IntegerLiteral) TokenLiteral() string      { return il.Token.Literal }
func (il *IntegerLiteral) String() string            { return il.Token.Literal }

type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()           {}
func (fl *FloatLiteral) TokenDetails() token.Token { return fl.Token }
func (fl *FloatLiteral) TokenLiteral() string      { return fl.Token.Literal }
func (fl *FloatLiteral) String() string            { return fl.Token.Literal }

type PrefixExpression struct {
	Token    token.Token // The prefix token, e.g. !
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()           {}
func (pe *PrefixExpression) TokenDetails() token.Token { return pe.Token }
func (pe *PrefixExpression) TokenLiteral() string      { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

type InfixExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (oe *InfixExpression) expressionNode()           {}
func (oe *InfixExpression) TokenDetails() token.Token { return oe.Token }
func (oe *InfixExpression) TokenLiteral() string      { return oe.Token.Literal }
func (oe *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(oe.Left.String())
	out.WriteString(" " + oe.Operator + " ")
	out.WriteString(oe.Right.String())
	out.WriteString(")")

	return out.String()
}

type AssignmentExpression struct {
	Token    token.Token // The operator token, e.g. =
	Left     Expression
	Operator string
	Right    Expression
}

func (ae *AssignmentExpression) expressionNode()           {}
func (ae *AssignmentExpression) TokenDetails() token.Token { return ae.Token }
func (ae *AssignmentExpression) TokenLiteral() string      { return ae.Token.Literal }
func (ae *AssignmentExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ae.Left.String())
	out.WriteString(" " + ae.Operator + " ")
	out.WriteString(ae.Right.String())
	out.WriteString(")")

	return out.String()
}

type PostAssignmentExpression struct {
	Token    token.Token // The operator token, e.g. ++
	Left     Expression
	Operator string
}

func (pae *PostAssignmentExpression) expressionNode()           {}
func (pae *PostAssignmentExpression) TokenDetails() token.Token { return pae.Token }
func (pae *PostAssignmentExpression) TokenLiteral() string      { return pae.Token.Literal }
func (pae *PostAssignmentExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pae.Left.String())
	out.WriteString(pae.Operator)
	out.WriteString(")")

	return out.String()
}

type IfStatement struct {
	Token       token.Token // The 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative Statement
}

func (is *IfStatement) statementNode()            {}
func (is *IfStatement) TokenDetails() token.Token { return is.Token }
func (is *IfStatement) TokenLiteral() string      { return is.Token.Literal }
func (is *IfStatement) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(is.Condition.String())
	out.WriteString(" ")
	out.WriteString(is.Consequence.String())

	if is.Alternative != nil {
		out.WriteString("else ")
		switch alternative := is.Alternative.(type) {
		case *BlockStatement:
			out.WriteString(alternative.String())
		case *IfStatement:
			out.WriteString(alternative.String())
		}
	}

	return out.String()
}

type ForStatement struct {
	Token          token.Token // The 'for' token
	Initialization Statement
	Condition      Expression
	Afterthought   Statement
	Loop           *BlockStatement
}

func (fs *ForStatement) statementNode()            {}
func (fs *ForStatement) TokenDetails() token.Token { return fs.Token }
func (fs *ForStatement) TokenLiteral() string      { return fs.Token.Literal }
func (fs *ForStatement) String() string {
	var out bytes.Buffer

	out.WriteString("for (")
	out.WriteString(fs.Initialization.String())
	out.WriteString("; ")
	out.WriteString(fs.Condition.String())
	out.WriteString("; ")
	out.WriteString(fs.Afterthought.String())
	out.WriteString(") ")
	out.WriteString(fs.Loop.String())

	return out.String()
}

type FunctionLiteral struct {
	Token      token.Token // The 'function' token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()           {}
func (fl *FunctionLiteral) TokenDetails() token.Token { return fl.Token }
func (fl *FunctionLiteral) TokenLiteral() string      { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())

	return out.String()
}

type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()           {}
func (ce *CallExpression) TokenDetails() token.Token { return ce.Token }
func (ce *CallExpression) TokenLiteral() string      { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()           {}
func (sl *StringLiteral) TokenDetails() token.Token { return sl.Token }
func (sl *StringLiteral) TokenLiteral() string      { return sl.Token.Literal }
func (sl *StringLiteral) String() string            { return sl.Token.Literal }

type ArrayLiteral struct {
	Token    token.Token // the '[' token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()           {}
func (al *ArrayLiteral) TokenDetails() token.Token { return al.Token }
func (al *ArrayLiteral) TokenLiteral() string      { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer

	elements := []string{}
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

type IndexExpression struct {
	Token token.Token // The [ token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()           {}
func (ie *IndexExpression) TokenDetails() token.Token { return ie.Token }
func (ie *IndexExpression) TokenLiteral() string      { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")

	return out.String()
}

type HashLiteral struct {
	Token token.Token // the '{' token
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) expressionNode()           {}
func (hl *HashLiteral) TokenDetails() token.Token { return hl.Token }
func (hl *HashLiteral) TokenLiteral() string      { return hl.Token.Literal }
func (hl *HashLiteral) String() string {
	var out bytes.Buffer

	pairs := []string{}
	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}
