package lexer

import (
	"github.com/ofux/deluge/dsl/token"
	"strconv"
)

type Lexer struct {
	input        []rune
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	line         int  // current line in input
	column       int  // current column in input
	ch           rune // current char under examination
}

func New(input string) *Lexer {
	l := &Lexer{input: []rune(input), line: 1}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespacesAndComments()

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case '+':
		if l.peekChar() == '+' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.ASSIGN_INC1, Literal: string(ch) + string(l.ch)}
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.ASSIGN_INC, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.PLUS, l.ch)
		}
	case '-':
		if l.peekChar() == '-' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.ASSIGN_DEC1, Literal: string(ch) + string(l.ch)}
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.ASSIGN_DEC, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.MINUS, l.ch)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.BANG, l.ch)
		}
	case '/':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.ASSIGN_DIV, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.SLASH, l.ch)
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.AND, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.OR, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	case '*':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.ASSIGN_MULT, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.ASTERISK, l.ch)
		}
	case '%':
		tok = newToken(token.MODULO, l.ch)
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.LTE, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.LT, l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.GTE, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.GT, l.ch)
		}
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ':':
		tok = newToken(token.COLON, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '"':
		tok.Line = l.line
		tok.Column = l.column
		tok.Type = token.STRING
		tok.Literal = l.readDoubleQuotedString()
		l.readChar()
		return tok
	case '`':
		tok.Line = l.line
		tok.Column = l.column
		tok.Type = token.STRING
		tok.Literal = l.readBackQuotedString()
		l.readChar()
		return tok
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Line = l.line
			tok.Column = l.column
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Line = l.line
			tok.Column = l.column
			tok.Literal, tok.Type = l.readNumber()
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	tok.Line = l.line
	tok.Column = l.column

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespacesAndComments() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' || l.skipLineComment() || l.skipBlockComment() {
		l.readChar()
	}
}

func (l *Lexer) skipLineComment() bool {
	if l.ch == '/' && l.peekChar() == '/' {
		for l.ch != 0 && l.ch != '\n' {
			l.readChar()
		}
		return true
	}
	return false
}

func (l *Lexer) skipBlockComment() bool {
	if l.ch == '/' && l.peekChar() == '*' {
		for l.ch != 0 && !(l.ch == '*' && l.peekChar() == '/') {
			l.readChar()
		}
		l.readChar()
		return true
	}
	return false
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1

	if l.ch == '\n' {
		l.line += 1
		l.column = 0
	} else {
		l.column += 1
	}
}

func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return string(l.input[position:l.position])
}

func (l *Lexer) readNumber() (string, token.TokenType) {
	position := l.position
	var tokType token.TokenType = token.INT
	for isDigit(l.ch) || (tokType == token.INT && l.ch == '.') {
		if l.ch == '.' {
			tokType = token.FLOAT
		}
		l.readChar()
	}
	return string(l.input[position:l.position]), tokType
}

func (l *Lexer) readDoubleQuotedString() string {
	position := l.position + 1
	for {
		prevCh := l.ch
		l.readChar()
		if (prevCh != '\\' && l.ch == '"') || l.ch == 0 || l.ch == '\n' {
			break
		}
	}
	str := string(l.input[position:l.position])

	// handles character escaping
	str, err := strconv.Unquote(`"` + str + `"`)
	if err != nil {
		panic(err)
	}
	return str
}

func (l *Lexer) readBackQuotedString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '`' || l.ch == 0 {
			break
		}
	}
	return string(l.input[position:l.position])
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func newToken(tokenType token.TokenType, ch rune) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}
