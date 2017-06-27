package object

func IsNumeric(object Object) bool {
	return object.Type() == INTEGER_OBJ || object.Type() == FLOAT_OBJ
}

func IsInteger(object Object) bool {
	return object.Type() == INTEGER_OBJ
}
