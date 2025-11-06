package lexer

import (
	"testing"

	"github.com/ofux/deluge/dsl/token"
)

type tokenExpectation struct {
	expectedType    token.TokenType
	expectedLiteral string
	expectedLine    int
	expectedColumn  int
}

func TestNextToken(t *testing.T) {
	input := `let five = 5;
let ten = 10;

let add = function(x, y) {
  x + y;
};

let result = add(five, ten);
! - / * 5;
5 < 10 > 5;

if (5 < 10) {
	return true;
} else {
	return false;
}

10 == 10;
10 != 9;
"foobar"
"foo bar"
"some utf8 : 🌧"
[1, 2];
{"foo": "bar"}
{
	"foo": "bar",
	"🌨🌨🌨":"⛄"
}

// line comment
1
/* block comment */
2
/* multi
line
comment
*/
3
4 // 2
5/* inline comment */6
7
// line comment
/* block comment */// yataa/* yotoo
/*
 // comment
*/
8

1 <= 2 >= 3
true && false || true
i--
i++
i += 1
i -= 1
i *= 1
i /= 1
`

	tests := []tokenExpectation{
		{token.LET, "let", 1, 1},
		{token.IDENT, "five", 1, 5},
		{token.ASSIGN, "=", 1, 10},
		{token.INT, "5", 1, 12},
		{token.SEMICOLON, ";", 1, 13},
		{token.LET, "let", 2, 1},
		{token.IDENT, "ten", 2, 5},
		{token.ASSIGN, "=", 2, 9},
		{token.INT, "10", 2, 11},
		{token.SEMICOLON, ";", 2, 13},
		{token.LET, "let", 4, 1},
		{token.IDENT, "add", 4, 5},
		{token.ASSIGN, "=", 4, 9},
		{token.FUNCTION, "function", 4, 11},
		{token.LPAREN, "(", 4, 19},
		{token.IDENT, "x", 4, 20},
		{token.COMMA, ",", 4, 21},
		{token.IDENT, "y", 4, 23},
		{token.RPAREN, ")", 4, 24},
		{token.LBRACE, "{", 4, 26},
		{token.IDENT, "x", 5, 3},
		{token.PLUS, "+", 5, 5},
		{token.IDENT, "y", 5, 7},
		{token.SEMICOLON, ";", 5, 8},
		{token.RBRACE, "}", 6, 1},
		{token.SEMICOLON, ";", 6, 2},
		{token.LET, "let", 8, 1},
		{token.IDENT, "result", 8, 5},
		{token.ASSIGN, "=", 8, 12},
		{token.IDENT, "add", 8, 14},
		{token.LPAREN, "(", 8, 17},
		{token.IDENT, "five", 8, 18},
		{token.COMMA, ",", 8, 22},
		{token.IDENT, "ten", 8, 24},
		{token.RPAREN, ")", 8, 27},
		{token.SEMICOLON, ";", 8, 28},
		{token.BANG, "!", 9, 1},
		{token.MINUS, "-", 9, 3},
		{token.SLASH, "/", 9, 5},
		{token.ASTERISK, "*", 9, 7},
		{token.INT, "5", 9, 9},
		{token.SEMICOLON, ";", 9, 10},
		{token.INT, "5", 10, 1},
		{token.LT, "<", 10, 3},
		{token.INT, "10", 10, 5},
		{token.GT, ">", 10, 8},
		{token.INT, "5", 10, 10},
		{token.SEMICOLON, ";", 10, 11},
		{token.IF, "if", 12, 1},
		{token.LPAREN, "(", 12, 4},
		{token.INT, "5", 12, 5},
		{token.LT, "<", 12, 7},
		{token.INT, "10", 12, 9},
		{token.RPAREN, ")", 12, 11},
		{token.LBRACE, "{", 12, 13},
		{token.RETURN, "return", 13, 2},
		{token.TRUE, "true", 13, 9},
		{token.SEMICOLON, ";", 13, 13},
		{token.RBRACE, "}", 14, 1},
		{token.ELSE, "else", 14, 3},
		{token.LBRACE, "{", 14, 8},
		{token.RETURN, "return", 15, 2},
		{token.FALSE, "false", 15, 9},
		{token.SEMICOLON, ";", 15, 14},
		{token.RBRACE, "}", 16, 1},
		{token.INT, "10", 18, 1},
		{token.EQ, "==", 18, 5},
		{token.INT, "10", 18, 7},
		{token.SEMICOLON, ";", 18, 9},
		{token.INT, "10", 19, 1},
		{token.NOT_EQ, "!=", 19, 5},
		{token.INT, "9", 19, 7},
		{token.SEMICOLON, ";", 19, 8},
		{token.STRING, "foobar", 20, 1},
		{token.STRING, "foo bar", 21, 1},
		{token.STRING, "some utf8 : 🌧", 22, 1},
		{token.LBRACKET, "[", 23, 1},
		{token.INT, "1", 23, 2},
		{token.COMMA, ",", 23, 3},
		{token.INT, "2", 23, 5},
		{token.RBRACKET, "]", 23, 6},
		{token.SEMICOLON, ";", 23, 7},
		{token.LBRACE, "{", 24, 1},
		{token.STRING, "foo", 24, 2},
		{token.COLON, ":", 24, 7},
		{token.STRING, "bar", 24, 9},
		{token.RBRACE, "}", 24, 14},
		{token.LBRACE, "{", 25, 1},
		{token.STRING, "foo", 26, 2},
		{token.COLON, ":", 26, 7},
		{token.STRING, "bar", 26, 9},
		{token.COMMA, ",", 26, 14},
		{token.STRING, "🌨🌨🌨", 27, 2},
		{token.COLON, ":", 27, 7},
		{token.STRING, "⛄", 27, 8},
		{token.RBRACE, "}", 28, 1},
		{token.INT, "1", 31, 1},
		{token.INT, "2", 33, 1},
		{token.INT, "3", 38, 1},
		{token.INT, "4", 39, 1},
		{token.INT, "5", 40, 1},
		{token.INT, "6", 40, 22},
		{token.INT, "7", 41, 1},
		{token.INT, "8", 47, 1},
		{token.INT, "1", 49, 1},
		{token.LTE, "<=", 49, 4},
		{token.INT, "2", 49, 6},
		{token.GTE, ">=", 49, 9},
		{token.INT, "3", 49, 11},
		{token.TRUE, "true", 50, 1},
		{token.AND, "&&", 50, 7},
		{token.FALSE, "false", 50, 9},
		{token.OR, "||", 50, 16},
		{token.TRUE, "true", 50, 18},
		{token.IDENT, "i", 51, 1},
		{token.ASSIGN_DEC1, "--", 51, 3},
		{token.IDENT, "i", 52, 1},
		{token.ASSIGN_INC1, "++", 52, 3},
		{token.IDENT, "i", 53, 1},
		{token.ASSIGN_INC, "+=", 53, 4},
		{token.INT, "1", 53, 6},
		{token.IDENT, "i", 54, 1},
		{token.ASSIGN_DEC, "-=", 54, 4},
		{token.INT, "1", 54, 6},
		{token.IDENT, "i", 55, 1},
		{token.ASSIGN_MULT, "*=", 55, 4},
		{token.INT, "1", 55, 6},
		{token.IDENT, "i", 56, 1},
		{token.ASSIGN_DIV, "/=", 56, 4},
		{token.INT, "1", 56, 6},
		{token.EOF, "", 57, 1},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()
		testToken(t, tok, tt, i)
	}
}

func TestReadDoubleQuotedString(t *testing.T) {
	input := `"aaa\nbbb"
"aaa\"bbb"
"aaa
bbb"
`

	tests := []tokenExpectation{
		{token.STRING, "aaa\nbbb", 1, 1},
		{token.STRING, "aaa\"bbb", 2, 1},
		{token.STRING, "aaa", 3, 1},
		{token.IDENT, "bbb", 4, 1},
		{token.STRING, "", 4, 4},
		{token.EOF, "", 5, 1},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()
		testToken(t, tok, tt, i)
	}
}

func TestReadBackQuotedString(t *testing.T) {
	input := "`aaa\\nbbb"
	input += "\n"
	input += "ccc`"

	tests := []tokenExpectation{
		{token.STRING, "aaa\\nbbb\nccc", 1, 1},
		{token.EOF, "", 2, 5},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()
		testToken(t, tok, tt, i)
	}
}

func TestForLoop(t *testing.T) {
	input := `
for (let i=0; i < 10; i=i+1) {
}
`

	tests := []tokenExpectation{
		{token.FOR, "for", 2, 1},
		{token.LPAREN, "(", 2, 5},
		{token.LET, "let", 2, 6},
		{token.IDENT, "i", 2, 10},
		{token.ASSIGN, "=", 2, 11},
		{token.INT, "0", 2, 12},
		{token.SEMICOLON, ";", 2, 13},
		{token.IDENT, "i", 2, 15},
		{token.LT, "<", 2, 17},
		{token.INT, "10", 2, 19},
		{token.SEMICOLON, ";", 2, 21},
		{token.IDENT, "i", 2, 23},
		{token.ASSIGN, "=", 2, 24},
		{token.IDENT, "i", 2, 25},
		{token.PLUS, "+", 2, 26},
		{token.INT, "1", 2, 27},
		{token.RPAREN, ")", 2, 28},
		{token.LBRACE, "{", 2, 30},
		{token.RBRACE, "}", 3, 1},
		{token.EOF, "", 4, 1},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()
		testToken(t, tok, tt, i)
	}
}

func TestFloats(t *testing.T) {
	input := `
33.0
42.42
67.6898 29938928.7
`

	tests := []tokenExpectation{
		{token.FLOAT, "33.0", 2, 1},
		{token.FLOAT, "42.42", 3, 1},
		{token.FLOAT, "67.6898", 4, 1},
		{token.FLOAT, "29938928.7", 4, 9},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()
		testToken(t, tok, tt, i)
	}
}

func TestOperators(t *testing.T) {
	input := `
5 % 2
`

	tests := []tokenExpectation{
		{token.INT, "5", 2, 1},
		{token.MODULO, "%", 2, 3},
		{token.INT, "2", 2, 5},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()
		testToken(t, tok, tt, i)
	}
}

func testToken(t *testing.T, tok token.Token, tt tokenExpectation, i int) {
	if tok.Type != tt.expectedType {
		t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
			i, tt.expectedType, tok.Type)
	}

	if tok.Literal != tt.expectedLiteral {
		t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
			i, tt.expectedLiteral, tok.Literal)
	}

	if tok.Line != tt.expectedLine {
		t.Fatalf("tests[%d] - line wrong. expected=%d, got=%d",
			i, tt.expectedLine, tok.Line)
	}

	if tok.Column != tt.expectedColumn {
		t.Fatalf("tests[%d] - column wrong. expected=%d, got=%d",
			i, tt.expectedColumn, tok.Column)
	}
}
