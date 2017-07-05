package object

import (
	"bytes"
	"fmt"
	"github.com/ofux/deluge/dsl/ast"
	"github.com/ofux/deluge/dsl/token"
	"strconv"
	"strings"
)

type BuiltinFunction func(node ast.Node, args ...Object) Object

type ObjectType string

const (
	NULL_OBJ         ObjectType = "NULL"
	ERROR_OBJ                   = "ERROR"
	INTEGER_OBJ                 = "INTEGER"
	FLOAT_OBJ                   = "FLOAT"
	BOOLEAN_OBJ                 = "BOOLEAN"
	STRING_OBJ                  = "STRING"
	RETURN_VALUE_OBJ            = "RETURN_VALUE"
	FUNCTION_OBJ                = "FUNCTION"
	BUILTIN_OBJ                 = "BUILTIN"
	ARRAY_OBJ                   = "ARRAY"
	HASH_OBJ                    = "HASH"
)

type HashKey string

type Hashable interface {
	HashKey() HashKey
}

type Object interface {
	Type() ObjectType
	Inspect() string
	Equals(other Object) bool
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Equals(other Object) bool {
	typed, ok := other.(*Integer)
	return ok && typed.Value == i.Value
}
func (i *Integer) HashKey() HashKey {
	return HashKey(strconv.FormatInt(i.Value, 10))
}

type Float struct {
	Value float64
}

func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) Inspect() string  { return fmt.Sprintf("%f", f.Value) }
func (f *Float) Equals(other Object) bool {
	typed, ok := other.(*Float)
	return ok && typed.Value == f.Value
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Equals(other Object) bool {
	typed, ok := other.(*Boolean)
	return ok && typed.Value == b.Value
}

type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }
func (n *Null) Equals(other Object) bool {
	_, ok := other.(*Null)
	return ok
}

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }
func (rv *ReturnValue) Equals(other Object) bool {
	otherRV, ok := other.(*ReturnValue)
	return ok && rv.Value.Equals(otherRV.Value)
}

type Error struct {
	Message    string
	StackToken []token.Token
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string {
	stacktrace := fmt.Sprintf("RUNTIME ERROR: %s", e.Message)
	if e.StackToken != nil {
		for _, tok := range e.StackToken {
			stacktrace += fmt.Sprintf("\n\tat %s (line %d, col %d)", tok.Literal, tok.Line, tok.Column)
		}
	}
	return stacktrace
}
func (e *Error) Equals(other Object) bool {
	typed, ok := other.(*Error)
	return ok && typed.Message == e.Message
}
func (e *Error) AddCallToStack(call *ast.CallExpression) {
	e.StackToken = append(e.StackToken, call.Function.TokenDetails())
}

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("function")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}
func (f *Function) Equals(other Object) bool {
	return f == other
}

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }
func (s *String) Equals(other Object) bool {
	typed, ok := other.(*String)
	return ok && typed.Value == s.Value
}
func (s *String) HashKey() HashKey {
	return HashKey(s.Value)
}

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }
func (b *Builtin) Equals(other Object) bool {
	return b == other
}

type Array struct {
	Elements []Object
}

func (ao *Array) Type() ObjectType { return ARRAY_OBJ }
func (ao *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}
func (ao *Array) Equals(other Object) bool {
	return ao == other
}

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType { return HASH_OBJ }
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}
func (h *Hash) Equals(other Object) bool {
	return h == other
}
func (h *Hash) Get(key string) (HashPair, bool) {
	r, ok := h.Pairs[HashKey(key)]
	return r, ok
}
