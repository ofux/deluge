package evaluator

import (
	"github.com/ofux/deluge/dsl/lexer"
	"github.com/ofux/deluge/dsl/object"
	"github.com/ofux/deluge/dsl/parser"
	"testing"
)

func BenchmarkNewDeluge(b *testing.B) {

	script := `
let fib = function(n) {
	if (n < 2) {
		return n;
	}
	return fib(n-1) + fib(n-2);
}
let r = fib(25);
r
`

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		b.Errorf("Parsing errors: %v", p.Errors())
		b.FailNow()
	}
	ev := NewEvaluator()

	for i := 0; i < b.N; i++ {
		v := ev.Eval(program, object.NewEnvironment())
		if v.Type() == object.ERROR_OBJ {
			b.Fatalf("error: %s", v.(*object.Error).Message)
		}
	}
}
