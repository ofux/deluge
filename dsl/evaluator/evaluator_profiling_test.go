package evaluator

import (
	"github.com/ofux/deluge/dsl/lexer"
	"github.com/ofux/deluge/dsl/object"
	"github.com/ofux/deluge/dsl/parser"
	"testing"
)

func BenchmarkEvaluator_Eval_FibRecursive(b *testing.B) {

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

func BenchmarkEvaluator_Eval_FibIter(b *testing.B) {

	script := `
let fib = function(n) {
	let x = 0;
	let y = 1;
	for (let i=0; i < n; i++) {
		let aux = x;
		x = x+y;
		y = aux;
	}
	return x;
}
let r = 0;
for (let i=0; i < 1000; i++) { 
	r = fib(50);
}
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

func fib(n int) int {
	if n < 2 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func fibIter(n int) int {
	x, y := 0, 1
	for i := 0; i < n; i++ {
		m := make(map[string]*object.Object)
		m["toot"] = nil
		//a := make([]*object.Object, 1)
		//a[0] = nil
		x, y = x+y, x
	}
	return x
}

func BenchmarkEvaluator_Eval_FibRecursive_GoComparison(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := fib(25)
		if v != 75025 {
			b.Fatalf("fib(25) is not %d", v)
		}
	}
}

func BenchmarkEvaluator_Eval_FibIter_GoComparison(b *testing.B) {
	param := 50
	for i := 0; i < b.N; i++ {
		for j := 0; j < 1000; j++ {
			v := fibIter(param)
			if v != 12586269025 {
				b.Fatalf("fib(50) is not %d", v)
			}
		}
	}
}
