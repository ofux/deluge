package evaluator

import (
	"github.com/ofux/deluge/dsl/object"
	"testing"
)

type expectedError struct {
	input           string
	expectedMessage string
}

func TestErrorHandling(t *testing.T) {
	tests := []expectedError{
		{
			"5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5.0 + true;",
			"type mismatch: FLOAT + BOOLEAN",
		},
		{
			"5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true",
			"unknown operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"true + false + true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5; true + false; 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`"Hello" - "World"`,
			"unknown operator: STRING - STRING",
		},
		{
			"if (10 > 1) { true + false; }",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if (false) { a } else if (b) { c }",
			"identifier not found: b",
		},
		{
			"if (false) { a } else if (true) { c }",
			"identifier not found: c",
		},
		{
			"if (true) { a } else if (b) { c }",
			"identifier not found: a",
		},
		{
			`
if (10 > 1) {
  if (10 > 1) {
    return true + false;
  }

  return 1;
}
`,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar",
			"identifier not found: foobar",
		},
		{
			`{"name": "Monkey"}[function(x) { x }];`,
			"unusable as hash key: FUNCTION",
		},
		{
			`999[1]`,
			"index operator not supported: INTEGER",
		},
		{
			`a = 3`,
			"identifier not found: a",
		},
		{
			`let a = 3; let a = 4;`,
			"variable a redeclared in this block",
		},
		{
			`
function() {
	let b = 1;
}();
b = 2;
`,
			"identifier not found: b",
		},
		{
			`"a" < "b"`,
			"unknown operator: STRING < STRING",
		},
		{
			`"a" > "b"`,
			"unknown operator: STRING > STRING",
		},
		{
			`"a" <= "b"`,
			"unknown operator: STRING <= STRING",
		},
		{
			`"a" >= "b"`,
			"unknown operator: STRING >= STRING",
		},
		{
			`"foo" || false`,
			"unknown operator: STRING || BOOLEAN",
		},
		{
			`false || "foo"`,
			"unknown operator: BOOLEAN || STRING",
		},
		{
			`"foo" && true`,
			"unknown operator: STRING && BOOLEAN",
		},
		{
			`true && "foo"`,
			"unknown operator: BOOLEAN && STRING",
		},
		{
			`"foo" || "bar"`,
			"unknown operator: STRING || BOOLEAN",
		},
		{
			`"foo" && "bar"`,
			"unknown operator: STRING && BOOLEAN",
		},
		{
			`let a = "x"; a--`,
			"unknown operator STRING -- INTEGER",
		},
		{
			`let a = "x"; a++`,
			"unknown operator STRING ++ INTEGER",
		},
		{
			`let a = "x"; a -= 2`,
			"unknown operator STRING -= INTEGER",
		},
		{
			`let a = "x"; a *= 2`,
			"unknown operator STRING *= INTEGER",
		},
		{
			`let a = "x"; a /= 2`,
			"unknown operator STRING /= INTEGER",
		},
		{
			`let a = 4; a += "1"`,
			"unknown operator INTEGER += STRING",
		},
		{
			`let a = 4; a -= "1"`,
			"unknown operator INTEGER -= STRING",
		},
		{
			`let a = 4; a *= "1"`,
			"unknown operator INTEGER *= STRING",
		},
		{
			`let a = 4; a /= "1"`,
			"unknown operator INTEGER /= STRING",
		},
		{
			`let a = b;`,
			"identifier not found: b",
		},
		{
			`let a = 1; a = b;`,
			"identifier not found: b",
		},
		{
			`let a = 1; a += x`,
			"identifier not found: x",
		},
		{
			`x += 1`,
			"identifier not found: x",
		},
		{
			`let a = "foo"; a += x`,
			"identifier not found: x",
		},
		{
			`-"foo";`,
			"unknown operator: -STRING",
		},
		{
			`-x;`,
			"identifier not found: x",
		},
		{
			`x();`,
			"identifier not found: x",
		},
		{
			`let f=function(){}; f(x);`,
			"identifier not found: x",
		},
		{
			`[1, 2, x]`,
			"identifier not found: x",
		},
		{
			`x[1]`,
			"identifier not found: x",
		},
		{
			`[1, 2][x]`,
			"identifier not found: x",
		},
		{
			`let a = [1, 2]; a += 1`,
			"unknown operator ARRAY +=",
		},
		{
			`let a = null; a += 1`,
			"unknown operator NULL +=",
		},
		{
			`f() && true`,
			"identifier not found: f",
		},
		{
			`true && f()`,
			"identifier not found: f",
		},
		{
			`f() || true`,
			"identifier not found: f",
		},
		{
			`false || f()`,
			"identifier not found: f",
		},
		{
			`3.0 % 2`,
			"unknown operator: FLOAT % INTEGER",
		},
	}

	for _, tt := range tests {
		testError(t, tt)
	}
}

func TestForLoopsErrorHandling(t *testing.T) {
	tests := []expectedError{
		{
			`for (f(); true; true) {}`,
			"identifier not found: f",
		},
		{
			`for (let i=0; f(); true) {}`,
			"identifier not found: f",
		},
		{
			`for (let i=0; true; f()) {}`,
			"identifier not found: f",
		},
		{
			`
			let f=function(){ return true; };
			for (let i=0; f(); i++) {
				f = null;
			}`,
			"not a function: NULL",
		},
		{
			`
			let f=function(){ return true; };
			for (let i=0; i < 10; f()) {
				f = null;
			}`,
			"not a function: NULL",
		},
		{
			`
			for (let i=0; i < 10; i++) {
				f()
			}`,
			"identifier not found: f",
		},
		{
			`{true: 5}[true]`,
			"unusable as hash key: BOOLEAN",
		},
		{
			`{false: 5}[false]`,
			"unusable as hash key: BOOLEAN",
		},
	}

	for _, tt := range tests {
		testError(t, tt)
	}
}

func TestHashErrorHandling(t *testing.T) {
	tests := []expectedError{
		{
			`{
				function(){}: "a"
			}`,
			"unusable as hash key: FUNCTION",
		},
		{
			`{
				f(): "a"
			}`,
			"identifier not found: f",
		},
		{
			`{
				"f": f()
			}`,
			"identifier not found: f",
		},
	}

	for _, tt := range tests {
		testError(t, tt)
	}
}

func TestEnclosingEnvironmentsErrors(t *testing.T) {
	t.Run("With let only", func(t *testing.T) {
		testError(t, expectedError{
			`
let first = 10;
let second = 10;

let ourFunction = function(first) {
  let second = 20;
  first + second + third;
};

ourFunction(30);
`,
			"identifier not found: third",
		})
	})

	t.Run("Function arguments", func(t *testing.T) {
		testError(t, expectedError{
			`
let first = 10;
let second = 10;

let ourFunction = function(third) {
};

first + second + third;
`,
			"identifier not found: third",
		})
	})

	t.Run("If block", func(t *testing.T) {
		testError(t, expectedError{
			`
let first = 10;
let second = 10;

if (first < 1000) {
  let second = 20;
  first + second + third;
}
`,
			"identifier not found: third",
		})
	})

	t.Run("If block with inner new variable", func(t *testing.T) {
		testError(t, expectedError{
			`
let first = 10;
let second = 10;

if (first < 1000) {
  let third = 20;
}

first + second + third;
`,
			"identifier not found: third",
		})
	})

	t.Run("For block", func(t *testing.T) {
		testError(t, expectedError{
			`
let first = 10;
let second = 10;
let third = 10;

for (let i=0; i < 20; i++) {
}

first + second + third + i;
`,
			"identifier not found: i",
		})
	})
}

func TestErrorStacktrace(t *testing.T) {
	tests := []struct {
		input              string
		expectedStacktrace string
	}{
		{
			`
if (10 > 1) {
  if (10 > 1) {
    return true + false;
  }

  return 1;
}
`,
			"RUNTIME ERROR: unknown operator: BOOLEAN + BOOLEAN\n\tat + (line 4, col 17)",
		},
		{
			`
let f = function(x) {
  return x * 1;
}

f("str");
`,
			"RUNTIME ERROR: unknown operator: STRING * INTEGER\n\tat * (line 3, col 12)\n\tat f (line 6, col 1)",
		},
		{
			`
let f = function(x, y) {
  return x + y;
}

f(42);
`,
			"RUNTIME ERROR: wrong number of arguments: expected 2, got 1\n\tat ( (line 6, col 2)\n\tat f (line 6, col 1)",
		},
		{
			`
let sum = function(a, b) {
  return a() + b();
};

let f1 = function() {
	return 42;
};

let f2 = function() {
	return true + false;
};

sum(f1, f2);
`,
			"RUNTIME ERROR: unknown operator: BOOLEAN + BOOLEAN\n\tat + (line 11, col 14)\n\tat b (line 3, col 16)\n\tat sum (line 14, col 1)",
		},
	}

	for i, tt := range tests {
		evaluated := testEval(t, tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned [%d]. got=%T(%+v)",
				i, evaluated, evaluated)
			continue
		}

		stacktrace := errObj.Inspect()
		if stacktrace != tt.expectedStacktrace {
			t.Errorf("wrong stacktrace [%d]. expected=%q, got=%q",
				i, tt.expectedStacktrace, stacktrace)
		}
	}
}

func TestEvalStringInfixExpression(t *testing.T) {
	tests := []expectedError{
		{
			`["array"] == "array"`,
			"cannot convert value of type ARRAY to STRING",
		},
		{
			`{"foo":"bar"} + "foo:bar"`,
			"cannot convert value of type HASH to STRING",
		},
		{
			`"function" == function(){}`,
			"cannot convert value of type FUNCTION to STRING",
		},
	}

	for _, tt := range tests {
		testError(t, tt)
	}
}

func testError(t *testing.T, tt expectedError) {

	evaluated := testEval(t, tt.input)

	errObj, ok := evaluated.(*object.Error)
	if !ok {
		t.Errorf("no error object returned. got=%T(%+v)",
			evaluated, evaluated)
		return
	}

	if errObj.Message != tt.expectedMessage {
		t.Errorf("wrong error message. expected=%q, got=%q",
			tt.expectedMessage, errObj.Message)
	}
}
