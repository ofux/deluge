package evaluator

import (
	"github.com/ofux/deluge/dsl/ast"
	"github.com/ofux/deluge/dsl/lexer"
	"github.com/ofux/deluge/dsl/object"
	"github.com/ofux/deluge/dsl/parser"
	"testing"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
		{"4 % 2", 0},
		{"4 % 3", 1},
		{"16 % 3", 1},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalFloatExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"5.42", 5.42},
		{"10.098", 10.098},
		{"-5.10", -5.1},
		{"-10.0", -10.0},
		{"5.0 + 5 + 5 + 5 - 10", 10.0},
		{"2.0 * 2 * 2 * 2 * 2", 32.0},
		{"-50.0 + 100.0 + -50.0", 0.0},
		{"5.0 * 2 + 10", 20.0},
		{"5.0 + 2 * 10", 25.0},
		{"20.0 + 2 * -10", 0.0},
		{"50.0 / 2.0 * 2 + 10", 60.0},
		{"2.0 * (5 + 10.0)", 30.0},
		{"3.0 * 3.0 * 3 + 10", 37.0},
		{"3.0 * (3 * 3) + 10", 37.0},
		{"(5.0 + 10.0 * 2 + 15 / 3) * 2 + -10", 50.0},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		testFloatObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	t.Run("simple boolean expressions", func(t *testing.T) {

		tests := []struct {
			input    string
			expected bool
		}{
			{"true", true},
			{"false", false},
			{"1 < 2", true},
			{"1 > 2", false},
			{"1 < 1", false},
			{"1 > 1", false},
			{"1 == 1", true},
			{"1 != 1", false},
			{"1 == 2", false},
			{"1 != 2", true},
			{"1.0 == 1", true},
			{"1.0 != 1", false},
			{"1.0 == 2", false},
			{"1.0 != 2", true},
			{"1 == 1.0", true},
			{"1 != 1.0", false},
			{"1 == 2.0", false},
			{"1 != 2.0", true},
			{"1.0 == 1.0", true},
			{"1.0 != 1.0", false},
			{"1.0 == 2.0", false},
			{"1.0 != 2.0", true},
			{"1.0001 == 1.0001", true},
			{"1.0001 != 1.0001", false},
			{"1.0001 == 1.0002", false},
			{"1.0001 != 1.0002", true},
			{"true == true", true},
			{"false == false", true},
			{"true == false", false},
			{"true != false", true},
			{"false != true", true},
			{"(1 < 2) == true", true},
			{"(1 < 2) == false", false},
			{"(1 > 2) == true", false},
			{"(1 > 2) == false", true},
			{"(1 < 1) == true", false},
			{"(1 < 1) == false", true},
			{"(1 > 1) == true", false},
			{"(1 > 1) == false", true},
			{"(1 <= 2) == true", true},
			{"(1 <= 2) == false", false},
			{"(1 >= 2) == true", false},
			{"(1 >= 2) == false", true},
			{"(1 <= 1) == true", true},
			{"(1 <= 1) == false", false},
			{"(1 >= 1) == true", true},
			{"(1 >= 1) == false", false},
			{"(1.0 < 2) == true", true},
			{"(1.0 < 2.0) == false", false},
			{"(1 > 2.0) == true", false},
			{"(1.0 > 2.0) == false", true},
			{"(1.0 < 1) == true", false},
			{"(1.0 < 1.0) == false", true},
			{"(1 > 1.0) == true", false},
			{"(1.0 > 1) == true", false},
			{"(1.0 > 1.0) == false", true},
			{"(1 <= 2.0) == true", true},
			{"(1.0 <= 2.0) == false", false},
			{"(1.0 >= 2.0) == true", false},
			{"(1.0 >= 2) == false", true},
			{"(1 <= 1.0) == true", true},
			{"(1.0 <= 1) == false", false},
			{"(1.0 >= 1.0) == true", true},
			{"(1.0 >= 1) == false", false},
			{"false || false", false},
			{"false || true", true},
			{"true || true", true},
			{"false && false", false},
			{"false && true", false},
			{"true && true", true},
			{"false && true || true", true},
			{"false && true || false", false},
			{"false && (true || true)", false},
		}

		for _, tt := range tests {
			evaluated := testEval(t, tt.input)
			testBooleanObject(t, evaluated, tt.expected)
		}
	})
	t.Run("AND, OR, code evaluation", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
		}{
			{`
			let x = 0;
			let f = function() {
				x = x + 1;
				return true;
			}
			let b = true || f();
			x;
			`, 0},
			{`
			let x = 0;
			let f = function() {
				x = x + 1;
				return true;
			}
			let b = false || f();
			x;
			`, 1},
			{`
			let x = 0;
			let f = function() {
				x = x + 1;
				return true;
			}
			let b = true && f();
			x;
			`, 1},
			{`
			let x = 0;
			let f = function() {
				x = x + 1;
				return true;
			}
			let b = false && f();
			x;
			`, 0},
			{`
			let x = 0;
			let f = function() {
				x = x + 1;
				return true;
			}
			let b = true && f() || f();
			x;
			`, 1},
			{`
			let x = 0;
			let f = function() {
				x = x + 1;
				return true;
			}
			let b = false && f() || f();
			x;
			`, 1},
			{`
			let x = 0;
			let f = function() {
				x = x + 1;
				return true;
			}
			let b = !(true && f()) || f();
			x;
			`, 2},
		}

		for _, tt := range tests {
			evaluated := testEval(t, tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		}
	})
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestNull(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"null"},
		{"let a = null; a"},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		testNullObject(t, evaluated)
	}
}
func TestIfElseStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (1 < 2) { 10 } else if (2 != 2) { 20 } else { 30 }", 10},
		{"if (2 < 2) { 10 } else if (2 == 2) { 20 } else { 30 }", 20},
		{"if (2 < 2) { 10 } else if (2 != 2) { 20 } else { 30 }", 30},
		{"if (2 < 2) { 10 } else if (2 != 2) { 20 } else if (true) { 30 }", 30},
		{"if (2 < 2) { 10 } else if (2 != 2) { 20 } else if (true) { 30 } else { 40 }", 30},
		{"if (2 < 2) { 10 } else if (2 != 2) { 20 } else if (false) { 30 } else { 40 }", 40},
		{"let a = null; if (a == null) { 10 } else { 20 }", 10},
		{"let a = null; if (a != null) { 10 } else { 20 }", 20},
		{"let a = null; if (a) { 10 } else { 20 }", 20},
		{"let a = null; if (!a) { 10 } else { 20 }", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestForStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`let sum = 0;
		for (let i = 0; i < 10; i += 2) {
			sum = sum + 1;
		}
		sum`, 5},
		{`let sum = 0;
		for (let i = 20; i > 0; i--) {
			sum = sum + i;
		}
		sum`, 210},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{"if (10 > 1) { return 10; }", 10},
		{
			`
if (10 > 1) {
  if (10 > 1) {
    return 10;
  }

  return 1;
}
`,
			10,
		},
		{
			`
let f = function(x) {
  return x;
  x + 10;
};
f(10);`,
			10,
		},
		{
			`
let f = function(x) {
   let result = x + 10;
   return result;
   return 10;
};
f(10);`,
			20,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(t, tt.input), tt.expected)
	}
}

func TestAssignStatements(t *testing.T) {
	t.Run("assign integers", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
		}{
			{"let a = 1; a = 5; a;", 5},
			{"let a = 1; a = 5 * 5; a;", 25},
			{"let a = 1; a = 5; let b = a; b;", 5},
			{"let a = 1; a = 5; let b = a; let c = a; c = c + b + 5; c;", 15},
			{"let a = 1+5; a;", 6},
			{"let a = 1; a += 5; a;", 6},
			{"let a = 1; a -= 5; a;", -4},
			{"let a = 4; a *= 5; a;", 20},
			{"let a = 4; a /= 2; a;", 2},
			{"let a = 4; a++; a;", 5},
			{"let a = 4; a--; a;", 3},
		}

		for _, tt := range tests {
			testIntegerObject(t, testEval(t, tt.input), tt.expected)
		}
	})

	t.Run("assign strings", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{`let a = "x"; a = "A"; a;`, "A"},
			{`let a = "x"; a = "A" + "B"; a;`, "AB"},
			{`let a = "x"; a += "A"; a;`, "xA"},
		}

		for _, tt := range tests {
			testStringObject(t, testEval(t, tt.input), tt.expected)
		}
	})
}

func TestFunctionObject(t *testing.T) {
	input := "function(x) { x + 2; };"

	evaluated := testEval(t, input)
	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not Function. got=%T (%+v)", evaluated, evaluated)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. Parameters=%+v",
			fn.Parameters)
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0])
	}

	expectedBody := "(x + 2)"

	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let identity = function(x) { x; }; identity(5);", 5},
		{"let identity = function(x) { return x; }; identity(5);", 5},
		{"let double = function(x) { x * 2; }; double(5);", 10},
		{"let add = function(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = function(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"function(x) { x; }(5)", 5},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(t, tt.input), tt.expected)
	}
}

func TestEnclosingEnvironments(t *testing.T) {
	t.Run("With let only", func(t *testing.T) {
		input := `
let first = 10;
let second = 10;
let third = 10;

let ourFunction = function(first) {
  let second = 20;

  first + second + third;
};

ourFunction(20) + first + second;`

		testIntegerObject(t, testEval(t, input), 70)
	})

	t.Run("With let and assign", func(t *testing.T) {
		input := `
let first = 10;
let second = 10;
let third = 10;

let ourFunction = function(first) {
  second = 20;

  first + second + third;
};

ourFunction(20) + first + second;`

		testIntegerObject(t, testEval(t, input), 80)
	})

	t.Run("If block", func(t *testing.T) {
		input := `
let first = 10;
let second = 10;
let third = 10;

if (first == 10) {
	let second = 20;
}

first + second + third;
`

		testIntegerObject(t, testEval(t, input), 30)
	})

	t.Run("For block", func(t *testing.T) {
		input := `
let first = 10;
let second = 10;
let third = 10;

for (let i=0; i < 5; i++) {
	first--;
	let second = 20;
}

first + second + third;
`

		testIntegerObject(t, testEval(t, input), 25)
	})
}

func TestClosures(t *testing.T) {
	input := `
let newAdder = function(x) {
  function(y) { x + y };
};

let addTwo = newAdder(2);
addTwo(2);`

	testIntegerObject(t, testEval(t, input), 4)
}

func TestStringLiteral(t *testing.T) {
	input := `"Hello World!"`

	evaluated := testEval(t, input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `"Hello" + " " + "World!"`

	evaluated := testEval(t, input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestCustomBuiltinFunctions(t *testing.T) {
	l := lexer.New("yo()")
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Errorf("Parsing errors: %v", p.Errors())
		t.FailNow()
	}
	env := object.NewEnvironment()
	ev := NewEvaluator()

	ev.AddBuiltin("yo", func(node ast.Node, args ...object.Object) object.Object {
		return &object.Integer{Value: 42}
	})

	evaluated := ev.Eval(program, env)
	testIntegerObject(t, evaluated, int64(42))
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	evaluated := testEval(t, input)
	result, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object is not Array. got=%T (%+v)", evaluated, evaluated)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong num of elements. got=%d",
			len(result.Elements))
	}

	testIntegerObject(t, result.Elements[0], 1)
	testIntegerObject(t, result.Elements[1], 4)
	testIntegerObject(t, result.Elements[2], 6)
}

func TestArrayIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			"[1, 2, 3][0]",
			1,
		},
		{
			"[1, 2, 3][1]",
			2,
		},
		{
			"[1, 2, 3][2]",
			3,
		},
		{
			"let i = 0; [1][i];",
			1,
		},
		{
			"[1, 2, 3][1 + 1];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[2];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];",
			6,
		},
		{
			"let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i]",
			2,
		},
		{
			"[1, 2, 3][3]",
			nil,
		},
		{
			"[1, 2, 3][-1]",
			nil,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestHashLiterals(t *testing.T) {
	input := `let two = "two";
	{
		"one": 10 - 9,
		two: 1 + 1,
		"thr" + "ee": 6 / 2,
		4: 4
	}`

	evaluated := testEval(t, input)
	result, ok := evaluated.(*object.Hash)
	if !ok {
		t.Fatalf("Eval didn't return Hash. got=%T (%+v)", evaluated, evaluated)
	}

	expected := map[object.HashKey]int64{
		(&object.String{Value: "one"}).HashKey():   1,
		(&object.String{Value: "two"}).HashKey():   2,
		(&object.String{Value: "three"}).HashKey(): 3,
		(&object.Integer{Value: 4}).HashKey():      4,
	}

	if len(result.Pairs) != len(expected) {
		t.Fatalf("Hash has wrong num of pairs. got=%d", len(result.Pairs))
	}

	for expectedKey, expectedValue := range expected {
		pair, ok := result.Pairs[expectedKey]
		if !ok {
			t.Errorf("no pair for given key in Pairs")
		}

		testIntegerObject(t, pair.Value, expectedValue)
	}
}

func TestHashIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`{"foo": 5}["foo"]`,
			5,
		},
		{
			`{"foo": 5}["bar"]`,
			nil,
		},
		{
			`let key = "foo"; {"foo": 5}[key]`,
			5,
		},
		{
			`{}["foo"]`,
			nil,
		},
		{
			`{5: 5}[5]`,
			5,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestFullPrograms(t *testing.T) {
	t.Run("Recursive Fibonacci", func(t *testing.T) {
		input := `
		let fib = function(n) {
			if (n < 2) {
				return n;
			}
			return fib(n-1) + fib(n-2);
		}

		let f5 = fib(6);
		f5;
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Integer)
		if !ok {
			t.Fatalf("Eval didn't return Integer. got=%T (%+v)", evaluated, evaluated)
		}

		if result.Value != 8 {
			t.Fatalf("Eval didn't return right value. got=%d expected=%d", result, 8)
		}
	})

	t.Run("Closure Fibonacci", func(t *testing.T) {
		input := `
		// fibonacci is a function that returns
		// a function that returns an int.
		let fibonacci = function() {
			let n = 0;
			let p = 1;
			return function() {
				let aux = n;
				n = n+p;
				p = aux;
				return n;
			}
		}

		let fib = fibonacci();
		for (let i = 0; i < 10; i=i+1) {
			fib();
		}
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Integer)
		if !ok {
			t.Fatalf("Eval didn't return Integer. got=%T (%+v)", evaluated, evaluated)
		}

		if result.Value != 55 {
			t.Fatalf("Eval didn't return right value. got=%d expected=%d", result, 55)
		}
	})
}

func testEval(t *testing.T, input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Errorf("Parsing errors: %v", p.Errors())
		t.FailNow()
	}
	env := object.NewEnvironment()
	ev := NewEvaluator()

	return ev.Eval(program, env)
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d",
			result.Value, expected)
		return false
	}

	return true
}

func testFloatObject(t *testing.T, obj object.Object, expected float64) bool {
	result, ok := obj.(*object.Float)
	if !ok {
		t.Errorf("object is not Float. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%f, want=%f",
			result.Value, expected)
		return false
	}

	return true
}

func testStringObject(t *testing.T, obj object.Object, expected string) bool {
	result, ok := obj.(*object.String)
	if !ok {
		t.Errorf("object is not String. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%s, want=%s",
			result.Value, expected)
		return false
	}

	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t",
			result.Value, expected)
		return false
	}
	return true
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}
