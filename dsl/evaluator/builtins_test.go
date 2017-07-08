package evaluator

import (
	"github.com/ofux/deluge/dsl/ast"
	"github.com/ofux/deluge/dsl/lexer"
	"github.com/ofux/deluge/dsl/object"
	"github.com/ofux/deluge/dsl/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestAddGlobalBuiltin(t *testing.T) {
	t.Run("Add a global built-in function", func(t *testing.T) {
		l := lexer.New("fooTest()")
		p := parser.New(l)
		program, ok := p.ParseProgram()
		if !ok {
			t.Errorf("Parsing errors: %v", p.Errors())
			t.FailNow()
		}
		env := object.NewEnvironment()
		ev := NewEvaluator()

		err := AddGlobalBuiltin("fooTest", func(node ast.Node, args ...object.Object) object.Object {
			return &object.Integer{Value: 42}
		})
		assert.NoError(t, err)

		evaluated := ev.Eval(program, env)
		testIntegerObject(t, evaluated, int64(42))

		// Remove the test function from global built-in to avoid any interaction with other tests
		delete(globalBuiltins, "fooTest")
	})

	t.Run("Add a global built-in function that already exists", func(t *testing.T) {

		err := AddGlobalBuiltin("fooTest", func(node ast.Node, args ...object.Object) object.Object {
			return &object.Integer{Value: 42}
		})
		assert.NoError(t, err)

		err = AddGlobalBuiltin("fooTest", func(node ast.Node, args ...object.Object) object.Object {
			return &object.Integer{Value: 42}
		})
		assert.Error(t, err)
		assert.Equal(t, "Global built-in function 'fooTest' is already defined", err.Error())

		// Remove the test function from global built-in to avoid any interaction with other tests
		delete(globalBuiltins, "fooTest")
	})
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
		{`arrayIndexOf([1, 2, 3], 2)`, 1},
		{`arrayIndexOf([], 2)`, -1},
		{`arrayIndexOf(1, 2)`, "wrong type of argument n°1. got=INTEGER, want=ARRAY"},
		{`arrayIndexOf()`, "wrong number of arguments. got=0, want=2"},
		{`stringIndexOf("abcd", "b")`, 1},
		{`stringIndexOf("abcd", "e")`, -1},
		{`stringIndexOf(1, 2)`, "wrong type of argument n°1. got=INTEGER, want=STRING"},
		{`stringIndexOf("1", 2)`, "wrong type of argument n°2. got=INTEGER, want=STRING"},
		{`stringIndexOf()`, "wrong number of arguments. got=0, want=2"},
		{`split("abcd", "b")`, []string{"a", "cd"}},
		{`split("abcd", "e")`, []string{"abcd"}},
		{`split("", "")`, []string{}},
		{`split("abcd", "")`, []string{"a", "b", "c", "d"}},
		{`split(1, 2)`, "wrong type of argument n°1. got=INTEGER, want=STRING"},
		{`split("1", 2)`, "wrong type of argument n°2. got=INTEGER, want=STRING"},
		{`split()`, "wrong number of arguments. got=0, want=2"},
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
		{`parseBool("true")`, true},
		{`parseBool("false")`, false},
		{`parseBool("a")`, `strconv.ParseBool: parsing "a": invalid syntax`},
		{`parseBool(true)`, "wrong type of argument n°1. got=BOOLEAN, want=STRING"},
		{`parseBool("true", "false")`, "wrong number of arguments. got=2, want=1"},
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
		case []string:
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
				testStringObject(t, array.Elements[i], expectedElem)
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

		require.True(t, ok)
		assert.Equal(t, "wrong type of argument n°1. got=INTEGER, want=STRING", result.Message)
	})

	t.Run("Pause with bad duration", func(t *testing.T) {
		input := `
		pause("2");
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Error)

		require.True(t, ok)
		assert.Equal(t, "time: missing unit in duration 2", result.Message)
	})

	t.Run("Pause with no argument", func(t *testing.T) {
		input := `
		pause();
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Error)

		require.True(t, ok)
		assert.Equal(t, "wrong number of arguments. got=0, want=1", result.Message)
	})
}

func TestBuiltinMerge(t *testing.T) {
	t.Run("Merge 2 simple hashes", func(t *testing.T) {
		input := `
		merge({
			"a": 1
		}, {
			"b": 2,
			"c": 3
		})
		`
		evaluated := testEval(t, input)

		result, ok := evaluated.(*object.Hash)
		require.True(t, ok)

		assert.Equal(t, map[object.HashKey]object.HashPair{
			"a": {Key: &object.String{"a"}, Value: &object.Integer{1}},
			"b": {Key: &object.String{"b"}, Value: &object.Integer{2}},
			"c": {Key: &object.String{"c"}, Value: &object.Integer{3}},
		}, result.Pairs)
	})

	t.Run("Merge 2 hashes with common keys", func(t *testing.T) {
		input := `
		merge({
			"a": 1
		}, {
			"a": 2,
			"b": 3
		})
		`
		evaluated := testEval(t, input)

		result, ok := evaluated.(*object.Hash)
		require.True(t, ok)

		assert.Equal(t, map[object.HashKey]object.HashPair{
			"a": {Key: &object.String{"a"}, Value: &object.Integer{2}},
			"b": {Key: &object.String{"b"}, Value: &object.Integer{3}},
		}, result.Pairs)
	})

	t.Run("Merge simple hash with an empty hash", func(t *testing.T) {
		input := `
		merge({
			"a": 1
		}, {
		})
		`
		evaluated := testEval(t, input)

		result, ok := evaluated.(*object.Hash)
		require.True(t, ok)

		assert.Equal(t, map[object.HashKey]object.HashPair{
			"a": {Key: &object.String{"a"}, Value: &object.Integer{1}},
		}, result.Pairs)
	})

	t.Run("Merge an empty hash with a simple hash", func(t *testing.T) {
		input := `
		merge({
		}, {
			"a": 1
		})
		`
		evaluated := testEval(t, input)

		result, ok := evaluated.(*object.Hash)
		require.True(t, ok)

		assert.Equal(t, map[object.HashKey]object.HashPair{
			"a": {Key: &object.String{"a"}, Value: &object.Integer{1}},
		}, result.Pairs)
	})

	t.Run("Merge 2 empty hashes", func(t *testing.T) {
		input := `
		merge({
		}, {
		})
		`
		evaluated := testEval(t, input)

		result, ok := evaluated.(*object.Hash)
		require.True(t, ok)

		assert.Equal(t, map[object.HashKey]object.HashPair{}, result.Pairs)
	})

	t.Run("Merge with no argument", func(t *testing.T) {
		input := `
		merge();
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Error)

		require.True(t, ok)
		assert.Equal(t, "wrong number of arguments. got=0, want=2", result.Message)
	})

	t.Run("Merge with bad 1st argument", func(t *testing.T) {
		input := `
		merge("a", {});
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Error)

		require.True(t, ok)
		assert.Equal(t, "wrong type of argument n°1. got=STRING, want=HASH", result.Message)
	})

	t.Run("Merge with bad 2nd argument", func(t *testing.T) {
		input := `
		merge({}, "a");
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Error)

		require.True(t, ok)
		assert.Equal(t, "wrong type of argument n°2. got=STRING, want=HASH", result.Message)
	})
}

func TestBuiltinKeys(t *testing.T) {
	t.Run("Keys of a simple hash", func(t *testing.T) {
		input := `
		keys({
			"a": 1,
			"b": 2
		})
		`
		evaluated := testEval(t, input)

		result, ok := evaluated.(*object.Array)
		require.True(t, ok)

		assert.Len(t, result.Elements, 2)
		assert.Contains(t, result.Elements, &object.String{"a"})
		assert.Contains(t, result.Elements, &object.String{"b"})
	})

	t.Run("Keys of a hash with integers as keys", func(t *testing.T) {
		input := `
		keys({
			"a": 1,
			"b": 2,
			3: 3
		})
		`
		evaluated := testEval(t, input)

		result, ok := evaluated.(*object.Array)
		require.True(t, ok)

		assert.Len(t, result.Elements, 3)
		assert.Contains(t, result.Elements, &object.String{"a"})
		assert.Contains(t, result.Elements, &object.String{"b"})
		assert.Contains(t, result.Elements, &object.String{"3"})
	})

	t.Run("Use the result of Keys", func(t *testing.T) {
		input := `
		let h = {
			"a": 5,
			"b": 8,
			42: 1
		};

		let k = keys(h);
		let sum = 0;
		for (let i=0; i < len(k); i++) {
			sum += h[k[i]];
		}
		sum
		`
		evaluated := testEval(t, input)

		result, ok := evaluated.(*object.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(14), result.Value)
	})

	t.Run("Keys with no argument", func(t *testing.T) {
		input := `
		keys();
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Error)

		require.True(t, ok)
		assert.Equal(t, "wrong number of arguments. got=0, want=1", result.Message)
	})

	t.Run("Keys with bad argument", func(t *testing.T) {
		input := `
		keys("a");
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Error)

		require.True(t, ok)
		assert.Equal(t, "wrong type of argument n°1. got=STRING, want=HASH", result.Message)
	})
}

func TestBuiltinParseJson(t *testing.T) {
	t.Run("Parse json", func(t *testing.T) {
		input := `
parseJson(` + "`" + `{
	"a": "foo",
	"b": 42,
	"c": {
		"ca": "cfoo",
		"cb": 43,
		"cc": [
			1,
			2
		],
		"cd": {
			"cda": "bar"
		}
	},
	"d": [
		"da",
		43,
		[],
		{},
		true,
		12.3
	],
	"e": 1.2,
	"f": false
}` + "`" + `)
`

		evaluated := testEval(t, input)
		if err, ok := evaluated.(*object.Error); ok {
			t.Fatal(err.Message, err.StackToken)
		}

		result, ok := evaluated.(*object.Hash)
		require.True(t, ok)

		deepEqual := object.DeepEquals(&object.Hash{
			Pairs: map[object.HashKey]object.HashPair{
				object.HashKey("a"): {Key: &object.String{"a"}, Value: &object.String{"foo"}},
				object.HashKey("b"): {Key: &object.String{"b"}, Value: &object.Integer{42}},
				object.HashKey("c"): {Key: &object.String{"c"}, Value: &object.Hash{
					Pairs: map[object.HashKey]object.HashPair{
						object.HashKey("ca"): {Key: &object.String{"ca"}, Value: &object.String{"cfoo"}},
						object.HashKey("cb"): {Key: &object.String{"cb"}, Value: &object.Integer{43}},
						object.HashKey("cc"): {Key: &object.String{"cc"}, Value: &object.Array{Elements: []object.Object{
							&object.Integer{1},
							&object.Integer{2},
						}}},
						object.HashKey("cd"): {Key: &object.String{"cd"}, Value: &object.Hash{
							Pairs: map[object.HashKey]object.HashPair{
								object.HashKey("cda"): {Key: &object.String{"cda"}, Value: &object.String{"bar"}},
							},
						}},
					},
				}},
				object.HashKey("d"): {Key: &object.String{"d"}, Value: &object.Array{Elements: []object.Object{
					&object.String{"da"},
					&object.Integer{43},
					&object.Array{[]object.Object{}},
					&object.Hash{map[object.HashKey]object.HashPair{}},
					&object.Boolean{true},
					&object.Float{12.3},
				}}},
				object.HashKey("e"): {Key: &object.String{"e"}, Value: &object.Float{1.2}},
				object.HashKey("f"): {Key: &object.String{"f"}, Value: &object.Boolean{false}},
			},
		}, result)

		assert.True(t, deepEqual)
	})

	t.Run("Parse json with no argument", func(t *testing.T) {
		input := `
		parseJson();
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Error)

		require.True(t, ok)
		assert.Equal(t, "wrong number of arguments. got=0, want=1", result.Message)
	})

	t.Run("Parse json with bad argument", func(t *testing.T) {
		input := `
		parseJson({
			"a":"b"
		});
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Error)

		require.True(t, ok)
		assert.Equal(t, "wrong type of argument n°1. got=HASH, want=STRING", result.Message)
	})

	t.Run("Parse json with bad json", func(t *testing.T) {
		input := `
		parseJson("(!)");
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Error)

		require.True(t, ok)
		assert.Equal(t, "invalid character '(' looking for beginning of value", result.Message)
	})
}

func TestBuiltinToJson(t *testing.T) {
	t.Run("To json", func(t *testing.T) {
		input := `
toJson({
	"a": "foo",
	"b": 42,
	"c": {
		"ca": "cfoo",
		"cb": 43,
		"cc": [
			1,
			2
		],
		"cd": {
			"cda": "bar"
		}
	},
	"d": [
		"da",
		43,
		[],
		{},
		true,
		12.3
	],
	"e": 1.2,
	"f": false
})
`

		evaluated := testEval(t, input)
		if err, ok := evaluated.(*object.Error); ok {
			t.Fatal(err.Message, err.StackToken)
		}

		result, ok := evaluated.(*object.String)
		require.True(t, ok)
		assert.Equal(t, `{"a":"foo","b":42,"c":{"ca":"cfoo","cb":43,"cc":[1,2],"cd":{"cda":"bar"}},"d":["da",43,[],{},true,12.3],"e":1.2,"f":false}`, result.Value)
	})

	t.Run("To json with no argument", func(t *testing.T) {
		input := `
		toJson();
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Error)

		require.True(t, ok)
		assert.Equal(t, "wrong number of arguments. got=0, want=1", result.Message)
	})

	t.Run("To json with bad argument", func(t *testing.T) {
		input := `
		toJson("{}");
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Error)

		require.True(t, ok)
		assert.Equal(t, "wrong type of argument n°1. got=STRING, want=HASH", result.Message)
	})

	t.Run("To json with un-serializable data", func(t *testing.T) {
		input := `
		toJson({
			"a": function(){}
		});
		`

		evaluated := testEval(t, input)
		result, ok := evaluated.(*object.Error)

		require.True(t, ok)
		assert.Equal(t, "Cannot convert Object of type *object.Function to a native type", result.Message)
	})
}

func TestBuiltinFromJsonToJson(t *testing.T) {
	t.Run("To json", func(t *testing.T) {
		input := `toJson(parseJson(toJson(parseJson(toJson({"a":"foo","b":42,"c":{"ca":"cfoo","cb":43,"cc":[1,2],"cd":{"cda":"bar"}},"d":["da",43,[],{},true,12.3],"e":1.2,"f":false})))))`

		evaluated := testEval(t, input)
		if err, ok := evaluated.(*object.Error); ok {
			t.Fatal(err.Message, err.StackToken)
		}

		result, ok := evaluated.(*object.String)
		require.True(t, ok)
		assert.Equal(t, `{"a":"foo","b":42,"c":{"ca":"cfoo","cb":43,"cc":[1,2],"cd":{"cda":"bar"}},"d":["da",43,[],{},true,12.3],"e":1.2,"f":false}`, result.Value)
	})
}
