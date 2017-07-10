package object

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStringHashKey(t *testing.T) {
	hello1 := &String{Value: "Hello World"}
	hello2 := &String{Value: "Hello World"}
	diff1 := &String{Value: "My name is johnny"}
	diff2 := &String{Value: "My name is johnny"}

	if hello1.HashKey() != hello2.HashKey() {
		t.Errorf("strings with same content have different hash keys")
	}

	if diff1.HashKey() != diff2.HashKey() {
		t.Errorf("strings with same content have different hash keys")
	}

	if hello1.HashKey() == diff1.HashKey() {
		t.Errorf("strings with different content have same hash keys")
	}
}

func TestIntegerHashKey(t *testing.T) {
	one1 := &Integer{Value: 1}
	one2 := &Integer{Value: 1}
	two1 := &Integer{Value: 2}
	two2 := &Integer{Value: 2}

	if one1.HashKey() != one2.HashKey() {
		t.Errorf("integers with same content have twoerent hash keys")
	}

	if two1.HashKey() != two2.HashKey() {
		t.Errorf("integers with same content have twoerent hash keys")
	}

	if one1.HashKey() == two1.HashKey() {
		t.Errorf("integers with twoerent content have same hash keys")
	}
}

func TestEquals(t *testing.T) {
	fn := &Function{}
	bl := &Builtin{}
	h := &Hash{}
	arr := &Array{}

	tests := []struct {
		input1   Object
		input2   Object
		expected bool
	}{
		{&Integer{Value: 1}, &Integer{Value: 1}, true},
		{&Integer{Value: 1}, &Integer{Value: 2}, false},
		{&Integer{Value: 1}, &String{Value: "1"}, false},
		{&Integer{Value: 1}, &Float{Value: 1}, false},

		{&Float{Value: 1}, &Float{Value: 1}, true},
		{&Float{Value: 1}, &Float{Value: 2}, false},
		{&Float{Value: 1}, &String{Value: "1"}, false},
		{&Float{Value: 1}, &Integer{Value: 1}, false},

		{&Boolean{Value: true}, &Boolean{Value: true}, true},
		{&Boolean{Value: true}, &Boolean{Value: false}, false},
		{&Boolean{Value: true}, &String{Value: "true"}, false},

		{&String{Value: "1"}, &String{Value: "1"}, true},
		{&String{Value: "1"}, &String{Value: "2"}, false},

		{&Error{Message: "1"}, &Error{Message: "1"}, true},
		{&Error{Message: "1"}, &Error{Message: "2"}, false},
		{&Error{Message: "1"}, &String{Value: "1"}, false},

		{&Null{}, &Null{}, true},
		{&Null{}, &String{Value: "null"}, false},
		{&Null{}, &Integer{Value: 0}, false},
		{&Null{}, &Function{}, false},
		{&Null{}, &Boolean{Value: false}, false},
		{&Null{}, &Float{Value: 0}, false},
		{&Null{}, &Hash{}, false},
		{&Null{}, &Array{}, false},

		{&ReturnValue{Value: &String{Value: "1"}}, &ReturnValue{Value: &String{Value: "1"}}, true},
		{&ReturnValue{Value: &String{Value: "1"}}, &ReturnValue{Value: &String{Value: "2"}}, false},
		{&ReturnValue{Value: fn}, &ReturnValue{Value: fn}, true},
		{&ReturnValue{Value: &Function{}}, &ReturnValue{Value: &Function{}}, false},

		{&Function{}, &Function{}, false},
		{fn, fn, true},

		{&Builtin{}, &Builtin{}, false},
		{bl, bl, true},

		{&Hash{}, &Hash{}, false},
		{h, h, true},

		{&Array{}, &Array{}, false},
		{arr, arr, true},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.input1.Equals(tt.input2))
		assert.Equal(t, tt.expected, tt.input2.Equals(tt.input1))
	}
}

func TestHash_GetAs(t *testing.T) {
	hash := &Hash{
		Pairs: map[HashKey]HashPair{
			HashKey("a"): {Key: &String{"a"}, Value: &String{"foo"}},
			HashKey("b"): {Key: &String{"b"}, Value: &Integer{42}},
			HashKey("c"): {Key: &String{"c"}, Value: &Hash{Pairs: map[HashKey]HashPair{}}},
			HashKey("d"): {Key: &String{"d"}, Value: &Array{[]Object{}}},
			HashKey("e"): {Key: &String{"e"}, Value: &Float{1.2}},
			HashKey("f"): {Key: &String{"f"}, Value: &Boolean{true}},
		},
	}

	va, err := hash.GetAsString("a")
	assert.NoError(t, err)
	assert.Equal(t, "foo", va.Value)

	vb, err := hash.GetAsInt("b")
	assert.NoError(t, err)
	assert.Equal(t, int64(42), vb.Value)

	vc, err := hash.GetAsHash("c")
	assert.NoError(t, err)
	assert.NotNil(t, vc.Pairs)

	vd, err := hash.GetAsArray("d")
	assert.NoError(t, err)
	assert.NotNil(t, vd.Elements)

	ve, err := hash.GetAsFloat("e")
	assert.NoError(t, err)
	assert.Equal(t, float64(1.2), ve.Value)

	vf, err := hash.GetAsBool("f")
	assert.NoError(t, err)
	assert.Equal(t, true, vf.Value)

	// Wrong type
	_, err = hash.GetAsString("b")
	assert.Error(t, err)
	// Key does not exist
	_, err = hash.GetAsString("bar")
	assert.Error(t, err)

	// Wrong type
	_, err = hash.GetAsInt("a")
	assert.Error(t, err)
	// Key does not exist
	_, err = hash.GetAsInt("bar")
	assert.Error(t, err)

	// Wrong type
	_, err = hash.GetAsFloat("a")
	assert.Error(t, err)
	// Key does not exist
	_, err = hash.GetAsFloat("bar")
	assert.Error(t, err)

	// Wrong type
	_, err = hash.GetAsBool("a")
	assert.Error(t, err)
	// Key does not exist
	_, err = hash.GetAsBool("bar")
	assert.Error(t, err)

	// Wrong type
	_, err = hash.GetAsArray("a")
	assert.Error(t, err)
	// Key does not exist
	_, err = hash.GetAsArray("bar")
	assert.Error(t, err)

	// Wrong type
	_, err = hash.GetAsHash("a")
	assert.Error(t, err)
	// Key does not exist
	_, err = hash.GetAsHash("bar")
	assert.Error(t, err)
}
