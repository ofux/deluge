package evaluator

import (
	"github.com/ofux/deluge/dsl/ast"
	"github.com/ofux/deluge/dsl/lexer"
	"github.com/ofux/deluge/dsl/object"
	"github.com/ofux/deluge/dsl/parser"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCustomGlobalBuiltinFunctions(t *testing.T) {
	l := lexer.New("yo()")
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Errorf("Parsing errors: %v", p.Errors())
		t.FailNow()
	}
	env := object.NewEnvironment()
	ev := NewEvaluator()

	AddGlobalBuiltin("yo", func(node ast.Node, args ...object.Object) object.Object {
		return &object.Integer{Value: 42}
	})

	evaluated := ev.Eval(program, env)
	testIntegerObject(t, evaluated, int64(42))
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, "wrong type of argument. got=INTEGER, want ARRAY or STRING"},
		{`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
		{`len([1, 2, 3])`, 3},
		{`len([])`, 0},
		{`first([1, 2, 3])`, 1},
		{`first([])`, nil},
		{`first(1)`, "wrong type of argument n°1. got=INTEGER, want=ARRAY"},
		{`first()`, "wrong number of arguments. got=0, want=1"},
		{`last([1, 2, 3])`, 3},
		{`last([])`, nil},
		{`last(1)`, "wrong type of argument n°1. got=INTEGER, want=ARRAY"},
		{`last()`, "wrong number of arguments. got=0, want=1"},
		{`rest([1, 2, 3])`, []int{2, 3}},
		{`rest([])`, nil},
		{`rest(1)`, "wrong type of argument n°1. got=INTEGER, want=ARRAY"},
		{`rest()`, "wrong number of arguments. got=0, want=1"},
		{`push([], 1)`, []int{1}},
		{`push(1, 1)`, "wrong type of argument n°1. got=INTEGER, want=ARRAY"},
		{`push()`, "wrong number of arguments. got=0, want=2"},
		{`parseInt("12")`, 12},
		{`parseInt("-12")`, -12},
		{`parseInt("12.3")`, `strconv.ParseInt: parsing "12.3": invalid syntax`},
		{`parseInt("a")`, `strconv.ParseInt: parsing "a": invalid syntax`},
		{`parseInt(12)`, "wrong type of argument n°1. got=INTEGER, want=STRING"},
		{`parseInt("1", "2")`, "wrong number of arguments. got=2, want=1"},
		{`parseFloat("12")`, float64(12.0)},
		{`parseFloat("-12")`, float64(-12.0)},
		{`parseFloat("12.3")`, float64(12.3)},
		{`parseFloat("a")`, `strconv.ParseFloat: parsing "a": invalid syntax`},
		{`parseFloat(12)`, "wrong type of argument n°1. got=INTEGER, want=STRING"},
		{`parseFloat("1", "2")`, "wrong number of arguments. got=2, want=1"},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case float64:
			testFloatObject(t, evaluated, expected)
		case nil:
			testNullObject(t, evaluated)
		case string:
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)",
					evaluated, evaluated)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q",
					expected, errObj.Message)
			}
		case []int:
			array, ok := evaluated.(*object.Array)
			if !ok {
				t.Errorf("obj not Array. got=%T (%+v)", evaluated, evaluated)
				continue
			}

			if len(array.Elements) != len(expected) {
				t.Errorf("wrong num of elements. want=%d, got=%d",
					len(expected), len(array.Elements))
				continue
			}

			for i, expectedElem := range expected {
				testIntegerObject(t, array.Elements[i], int64(expectedElem))
			}
		}
	}
}

func TestBuiltinExit(t *testing.T) {
	t.Run("Exit before end", func(t *testing.T) {
		input := `
		let a = 1;
		a = 2;
		exit();
		a = 3;
		undefinedFunc();
		`

		evaluated := testEval(t, input)
		_, ok := evaluated.(*object.Null)
		if !ok {
			t.Fatalf("Eval didn't return Integer. got=%T (%+v)", evaluated, evaluated)
		}
	})

	t.Run("Exit before end with argument", func(t *testing.T) {
		input := `
		let a = 1;
		a = 2;
		exit(a);
		a = 3;
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Integer)
		if !ok {
			t.Fatalf("Eval didn't return Integer. got=%T (%+v)", evaluated, evaluated)
		}

		if result.Value != 2 {
			t.Fatalf("Eval didn't return right value. got=%d expected=%d", result, 2)
		}
	})

	t.Run("Exit before end in sub-scopes", func(t *testing.T) {
		input := `
		let a = 1;
		a = 2;
		if (a == 2) {
			function() {
				exit();
			}();
		}
		a = 3;
		undefinedFunc();
		`

		evaluated := testEval(t, input)
		_, ok := evaluated.(*object.Null)
		if !ok {
			t.Fatalf("Eval didn't return Integer. got=%T (%+v)", evaluated, evaluated)
		}
	})
}

func TestBuiltinAssert(t *testing.T) {
	t.Run("Assert success", func(t *testing.T) {
		input := `
		let a = 1;
		a = 2;
		assert(a == 2);
		a = 3;
		a;
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Integer)
		if !ok {
			t.Fatalf("Eval didn't return Integer. got=%T (%+v)", evaluated, evaluated)
		}

		if result.Value != 3 {
			t.Fatalf("Eval didn't return right value. got=%d expected=%d", result, 3)
		}
	})

	t.Run("Assert failure", func(t *testing.T) {
		input := `
		let a = 1;
		a = 2;
		assert(a == 20);
		a = 3;
		undefinedFunc();
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Error)
		if !ok {
			t.Fatalf("Eval didn't return Error. got=%T (%+v)", evaluated, evaluated)
		}

		if result.Message != "Assertion failed" {
			t.Fatalf("Bad error message. Expected '%s', got '%s'", "Assertion failed", result.Message)
		}

		if result.StackToken[0].Line != 4 {
			t.Fatalf("Wrong line for error. Expected %d, got %d", 4, result.StackToken[0].Line)
		}
	})

	t.Run("Assert fail in sub-scopes", func(t *testing.T) {
		input := `
		let a = 1;
		a = 2;
		if (a == 2) {
			function() {
				assert(false);
			}();
		}
		a = 3;
		undefinedFunc();
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Error)
		if !ok {
			t.Fatalf("Eval didn't return Error. got=%T (%+v)", evaluated, evaluated)
		}

		if result.Message != "Assertion failed" {
			t.Fatalf("Bad error message. Expected '%s', got '%s'", "Assertion failed", result.Message)
		}

		if result.StackToken[0].Line != 6 {
			t.Fatalf("Wrong line for error. Expected %d, got %d", 6, result.StackToken[0].Line)
		}
	})

	t.Run("Assert without argument", func(t *testing.T) {
		input := `
		assert();
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Error)
		if !ok {
			t.Fatalf("Eval didn't return Error. got=%T (%+v)", evaluated, evaluated)
		}

		if result.Message != "wrong number of arguments. got=0, want=1" {
			t.Fatalf("Bad error message. Expected '%s', got '%s'", "wrong number of arguments. got=0, want=1", result.Message)
		}
	})

	t.Run("Assert with bad argument", func(t *testing.T) {
		input := `
		assert(3);
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Error)
		if !ok {
			t.Fatalf("Eval didn't return Error. got=%T (%+v)", evaluated, evaluated)
		}

		if result.Message != "wrong type of argument n°1. got=INTEGER, want=BOOLEAN" {
			t.Fatalf("Bad error message. Expected '%s', got '%s'", "wrong type of argument n°1. got=INTEGER, want=BOOLEAN", result.Message)
		}
	})
}

func TestBuiltinPause(t *testing.T) {
	t.Run("Pause 2ms", func(t *testing.T) {
		input := `
		let a = 1;
		pause("31ms");
		a = 2;
		a;
		`
		start := time.Now()
		evaluated := testEval(t, input)
		elaspedTime := time.Now().Sub(start)
		if elaspedTime.Nanoseconds() < 30000000 {
			t.Fatalf("Eval didn't last at least 30000000ns.")
		}
		result, ok := evaluated.(*object.Integer)
		if !ok {
			t.Fatalf("Eval didn't return Integer. got=%T (%+v)", evaluated, evaluated)
		}

		if result.Value != 2 {
			t.Fatalf("Eval didn't return right value. got=%d expected=%d", result, 2)
		}
	})

	t.Run("Pause with bad type of argument", func(t *testing.T) {
		input := `
		pause(2);
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Error)

		assert.True(t, ok)
		assert.Equal(t, "wrong type of argument n°1. got=INTEGER, want=STRING", result.Message)
	})

	t.Run("Pause with bad duration", func(t *testing.T) {
		input := `
		pause("2");
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Error)

		assert.True(t, ok)
		assert.Equal(t, "time: missing unit in duration 2", result.Message)
	})

	t.Run("Pause with no argument", func(t *testing.T) {
		input := `
		pause();
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Error)

		assert.True(t, ok)
		assert.Equal(t, "wrong number of arguments. got=0, want=1", result.Message)
	})
}
