package object

import (
	"errors"
	"fmt"
	"math"
	"reflect"
)

func IsNumeric(object Object) bool {
	return object.Type() == INTEGER_OBJ || object.Type() == FLOAT_OBJ
}

func IsInteger(object Object) bool {
	return object.Type() == INTEGER_OBJ
}

func DeepEquals(o1, o2 Object) bool {
	if o1.Type() != o2.Type() {
		return false
	}
	switch o1.Type() {
	case HASH_OBJ:
		o1 := o1.(*Hash)
		o2 := o2.(*Hash)
		if len(o1.Pairs) != len(o2.Pairs) {
			return false
		}
		for k, v1 := range o1.Pairs {
			v2, ok := o2.Pairs[k]
			if !ok {
				return false
			}
			if !DeepEquals(v1.Key, v2.Key) || !DeepEquals(v1.Value, v2.Value) {
				return false
			}
		}
	case ARRAY_OBJ:
		o1 := o1.(*Array)
		o2 := o2.(*Array)
		if len(o1.Elements) != len(o2.Elements) {
			return false
		}
		for i, v := range o1.Elements {
			if !DeepEquals(v, o2.Elements[i]) {
				return false
			}
		}
	default:
		return o1.Equals(o2)
	}
	return true
}

func ToObject(in interface{}) (Object, error) {
	switch in := in.(type) {
	case string:
		return &String{Value: in}, nil
	case int:
		return &Integer{Value: int64(in)}, nil
	case int8:
		return &Integer{Value: int64(in)}, nil
	case int16:
		return &Integer{Value: int64(in)}, nil
	case int32:
		return &Integer{Value: int64(in)}, nil
	case int64:
		return &Integer{Value: in}, nil
	case uint:
		return &Integer{Value: int64(in)}, nil
	case uint8:
		return &Integer{Value: int64(in)}, nil
	case uint16:
		return &Integer{Value: int64(in)}, nil
	case uint32:
		return &Integer{Value: int64(in)}, nil
	case uint64:
		return &Integer{Value: int64(in)}, nil
	case float64:
		if in == math.Trunc(in) && in < math.MaxInt64 {
			return &Integer{Value: int64(in)}, nil
		}
		return &Float{Value: in}, nil
	case float32:
		if float64(in) == math.Trunc(float64(in)) && in < math.MaxInt64 {
			return &Integer{Value: int64(in)}, nil
		}
		return &Float{Value: float64(in)}, nil
	case bool:
		return &Boolean{Value: in}, nil
	case []interface{}:
		elements := make([]Object, 0, len(in))
		for _, v := range in {
			obj, err := ToObject(v)
			if err != nil {
				return nil, err
			}
			elements = append(elements, obj)
		}
		return &Array{Elements: elements}, nil
	case map[string]interface{}:
		pairs := make(map[HashKey]HashPair)
		for k, v := range in {
			obj, err := ToObject(v)
			if err != nil {
				return nil, err
			}
			key := &String{Value: k}
			pairs[key.HashKey()] = HashPair{Key: key, Value: obj}
		}
		return &Hash{Pairs: pairs}, nil
	default:
		return nil, errors.New(fmt.Sprintf("Cannot convert value of type %s to Object", reflect.TypeOf(in)))
	}
}

func FromObject(in Object) (interface{}, error) {
	switch in := in.(type) {
	case *String:
		return in.Value, nil
	case *Integer:
		return in.Value, nil
	case *Float:
		return in.Value, nil
	case *Boolean:
		return in.Value, nil
	case *Null:
		return nil, nil
	case *ReturnValue:
		return FromObject(in.Value)
	case *Hash:
		pairs := make(map[string]interface{})
		for k, v := range in.Pairs {
			val, err := FromObject(v.Value)
			if err != nil {
				return nil, err
			}
			pairs[string(k)] = val
		}
		return pairs, nil
	case *Array:
		elements := make([]interface{}, 0, len(in.Elements))
		for _, v := range in.Elements {
			val, err := FromObject(v)
			if err != nil {
				return nil, err
			}
			elements = append(elements, val)
		}
		return elements, nil
	default:
		return nil, errors.New(fmt.Sprintf("Cannot convert Object of type %s to a native type", reflect.TypeOf(in)))
	}
}
