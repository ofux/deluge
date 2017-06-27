package token

type TokenType string

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	IDENT  = "IDENT"  // add, foobar, x, y, ...
	INT    = "INT"    // 1343456
	FLOAT  = "FLOAT"  // 1238.873
	STRING = "STRING" // "foobar"

	ASSIGN      = "="
	ASSIGN_DEC1 = "--"
	ASSIGN_INC1 = "++"
	ASSIGN_DEC  = "-="
	ASSIGN_INC  = "+="
	ASSIGN_MULT = "*="
	ASSIGN_DIV  = "/="

	// Operators
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	MODULO   = "%"

	AND = "&&"
	OR  = "||"

	LT  = "<"
	GT  = ">"
	LTE = "<="
	GTE = ">="

	EQ     = "=="
	NOT_EQ = "!="

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	// Keywords
	FUNCTION = "FUNCTION"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	FOR      = "FOR"
	NULL     = "NULL"
)

type Token struct {
	Type    TokenType
	Line    int
	Column  int
	Literal string
}

var keywords = map[string]TokenType{
	"function": FUNCTION,
	"let":      LET,
	"true":     TRUE,
	"false":    FALSE,
	"if":       IF,
	"else":     ELSE,
	"return":   RETURN,
	"for":      FOR,
	"null":     NULL,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

func IsAssign(t TokenType) bool {
	return t == ASSIGN || t == ASSIGN_INC1 || t == ASSIGN_DEC1 || t == ASSIGN_INC || t == ASSIGN_DEC || t == ASSIGN_MULT || t == ASSIGN_DIV
}
