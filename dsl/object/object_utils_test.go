package object

import (
	"github.com/dustin/gojson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDeepEquals(t *testing.T) {
	t.Run("Deep equality", func(t *testing.T) {

		o1 := &Hash{
			Pairs: map[HashKey]HashPair{
				HashKey("a"): {Key: &String{"a"}, Value: &String{"foo"}},
				HashKey("b"): {Key: &String{"b"}, Value: &Integer{42}},
				HashKey("c"): {Key: &String{"c"}, Value: &Hash{
					Pairs: map[HashKey]HashPair{
						HashKey("ca"): {Key: &String{"ca"}, Value: &String{"cfoo"}},
						HashKey("cb"): {Key: &String{"cb"}, Value: &Integer{43}},
						HashKey("cc"): {Key: &String{"cc"}, Value: &Array{[]Object{
							&Integer{1},
							&Integer{2},
						}}},
						HashKey("cd"): {Key: &String{"cd"}, Value: &Hash{
							Pairs: map[HashKey]HashPair{
								HashKey("cda"): {Key: &String{"cda"}, Value: &String{"bar"}},
							},
						}},
					},
				}},
				HashKey("d"): {Key: &String{"d"}, Value: &Array{[]Object{
					&String{"da"},
					&Integer{43},
					&Array{[]Object{}},
					&Hash{map[HashKey]HashPair{}},
					&Boolean{true},
					&Float{12.3},
				}}},
				HashKey("e"): {Key: &String{"e"}, Value: &Float{1.2}},
				HashKey("f"): {Key: &String{"f"}, Value: &Boolean{false}},
			},
		}

		o2 := &Hash{
			Pairs: map[HashKey]HashPair{
				HashKey("a"): {Key: &String{"a"}, Value: &String{"foo"}},
				HashKey("b"): {Key: &String{"b"}, Value: &Integer{42}},
				HashKey("c"): {Key: &String{"c"}, Value: &Hash{
					Pairs: map[HashKey]HashPair{
						HashKey("ca"): {Key: &String{"ca"}, Value: &String{"cfoo"}},
						HashKey("cb"): {Key: &String{"cb"}, Value: &Integer{43}},
						HashKey("cc"): {Key: &String{"cc"}, Value: &Array{[]Object{
							&Integer{1},
							&Integer{2},
						}}},
						HashKey("cd"): {Key: &String{"cd"}, Value: &Hash{
							Pairs: map[HashKey]HashPair{
								HashKey("cda"): {Key: &String{"cda"}, Value: &String{"bar"}},
							},
						}},
					},
				}},
				HashKey("d"): {Key: &String{"d"}, Value: &Array{[]Object{
					&String{"da"},
					&Integer{43},
					&Array{[]Object{}},
					&Hash{map[HashKey]HashPair{}},
					&Boolean{true},
					&Float{12.3},
				}}},
				HashKey("e"): {Key: &String{"e"}, Value: &Float{1.2}},
				HashKey("f"): {Key: &String{"f"}, Value: &Boolean{false}},
			},
		}

		assert.True(t, DeepEquals(o1, o2))
	})

	t.Run("Not equal because of array length", func(t *testing.T) {
		o1 := &Array{
			[]Object{
				&String{"da"},
				&Integer{43},
			},
		}

		o2 := &Array{
			[]Object{
				&String{"da"},
				&Integer{43},
				&Integer{44},
			},
		}

		assert.False(t, DeepEquals(o1, o2))
	})

	t.Run("Not equal because of array's value", func(t *testing.T) {
		o1 := &Array{
			[]Object{
				&String{"da"},
				&Integer{43},
			},
		}

		o2 := &Array{
			[]Object{
				&String{"da"},
				&Integer{44},
			},
		}

		assert.False(t, DeepEquals(o1, o2))
	})

	t.Run("Not equal because of array's value's type", func(t *testing.T) {
		o1 := &Array{
			[]Object{
				&String{"1"},
			},
		}

		o2 := &Array{
			[]Object{
				&Integer{1},
			},
		}

		assert.False(t, DeepEquals(o1, o2))
	})

	t.Run("Not equal because of hash length", func(t *testing.T) {
		o1 := &Hash{
			Pairs: map[HashKey]HashPair{
				HashKey("a"): {Key: &String{"a"}, Value: &String{"foo"}},
				HashKey("b"): {Key: &String{"b"}, Value: &Integer{42}},
			},
		}

		o2 := &Hash{
			Pairs: map[HashKey]HashPair{
				HashKey("a"): {Key: &String{"a"}, Value: &String{"foo"}},
				HashKey("b"): {Key: &String{"b"}, Value: &Integer{42}},
				HashKey("c"): {Key: &String{"c"}, Value: &Integer{43}},
			},
		}

		assert.False(t, DeepEquals(o1, o2))
	})

	t.Run("Not equal because of hash's value", func(t *testing.T) {
		o1 := &Hash{
			Pairs: map[HashKey]HashPair{
				HashKey("a"): {Key: &String{"a"}, Value: &String{"foo"}},
				HashKey("b"): {Key: &String{"b"}, Value: &Integer{42}},
			},
		}

		o2 := &Hash{
			Pairs: map[HashKey]HashPair{
				HashKey("a"): {Key: &String{"a"}, Value: &String{"foo"}},
				HashKey("b"): {Key: &String{"b"}, Value: &Integer{43}},
			},
		}

		assert.False(t, DeepEquals(o1, o2))
	})

	t.Run("Not equal because of hash's key", func(t *testing.T) {
		o1 := &Hash{
			Pairs: map[HashKey]HashPair{
				HashKey("a"): {Key: &String{"a"}, Value: &String{"foo"}},
				HashKey("b"): {Key: &String{"b"}, Value: &Integer{42}},
			},
		}

		o2 := &Hash{
			Pairs: map[HashKey]HashPair{
				HashKey("a"): {Key: &String{"a"}, Value: &String{"foo"}},
				HashKey("c"): {Key: &String{"c"}, Value: &Integer{42}},
			},
		}

		assert.False(t, DeepEquals(o1, o2))
	})
}

func TestToObject(t *testing.T) {
	t.Run("Test ToObject", func(t *testing.T) {
		tests := []struct {
			input    interface{}
			expected Object
		}{
			{"string", &String{"string"}},
			{int(3), &Integer{3}},
			{int8(3), &Integer{3}},
			{int16(3), &Integer{3}},
			{int32(3), &Integer{3}},
			{int64(3), &Integer{3}},
			{uint(3), &Integer{3}},
			{uint8(3), &Integer{3}},
			{uint16(3), &Integer{3}},
			{uint32(3), &Integer{3}},
			{uint64(3), &Integer{3}},
			{float32(3), &Integer{3}},
			{float64(3), &Integer{3}},
			{float32(3.2), &Float{3.2}},
			{float64(3.2), &Float{3.2}},
			{float32(3E38), &Float{3E38}},
			{float64(3E300), &Float{3E300}},
		}

		for _, tt := range tests {
			obj, err := ToObject(tt.input)
			require.NoError(t, err)
			assert.True(t, obj.Equals(tt.expected))
		}
	})

	t.Run("Test ToObject", func(t *testing.T) {
		tests := []struct {
			input    interface{}
			expected string
		}{
			{func() {}, "Cannot convert value of type func() to Object"},
			{map[string]interface{}{"a": func() {}}, "Cannot convert value of type func() to Object"},
			{[]interface{}{func() {}}, "Cannot convert value of type func() to Object"},
		}

		for _, tt := range tests {
			_, err := ToObject(tt.input)
			require.Error(t, err)
			assert.Equal(t, tt.expected, err.Error())
		}
	})

	t.Run("Test ToObject from JSON input", func(t *testing.T) {

		input := `
{
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
}`
		in := make(map[string]interface{})
		err := json.Unmarshal([]byte(input), &in)
		if err != nil {
			t.Fatal(err)
		}

		obj, err := ToObject(in)
		if err != nil {
			t.Fatal(err)
		}
		deepEqual := DeepEquals(&Hash{
			Pairs: map[HashKey]HashPair{
				HashKey("a"): {Key: &String{"a"}, Value: &String{"foo"}},
				HashKey("b"): {Key: &String{"b"}, Value: &Integer{42}},
				HashKey("c"): {Key: &String{"c"}, Value: &Hash{
					Pairs: map[HashKey]HashPair{
						HashKey("ca"): {Key: &String{"ca"}, Value: &String{"cfoo"}},
						HashKey("cb"): {Key: &String{"cb"}, Value: &Integer{43}},
						HashKey("cc"): {Key: &String{"cc"}, Value: &Array{[]Object{
							&Integer{1},
							&Integer{2},
						}}},
						HashKey("cd"): {Key: &String{"cd"}, Value: &Hash{
							Pairs: map[HashKey]HashPair{
								HashKey("cda"): {Key: &String{"cda"}, Value: &String{"bar"}},
							},
						}},
					},
				}},
				HashKey("d"): {Key: &String{"d"}, Value: &Array{[]Object{
					&String{"da"},
					&Integer{43},
					&Array{[]Object{}},
					&Hash{map[HashKey]HashPair{}},
					&Boolean{true},
					&Float{12.3},
				}}},
				HashKey("e"): {Key: &String{"e"}, Value: &Float{1.2}},
				HashKey("f"): {Key: &String{"f"}, Value: &Boolean{false}},
			},
		}, obj)

		assert.True(t, deepEqual)
	})
}
