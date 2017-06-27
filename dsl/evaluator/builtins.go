package evaluator

import (
	"errors"
	"fmt"
	"github.com/ofux/deluge/dsl/ast"
	"github.com/ofux/deluge/dsl/object"
	"github.com/ofux/deluge/dsl/token"
	"time"
)

func AddGlobalBuiltin(name string, fn object.BuiltinFunction) error {
	if _, ok := globalBuiltins[name]; ok {
		return errors.New(fmt.Sprintf("Global built-in function '%s' is already defined", name))
	}
	globalBuiltins[name] = &object.Builtin{Fn: fn}
	return nil
}

func AssertArgCount(node ast.Node, args []object.Object, count int) *object.Error {
	if len(args) != count {
		return NewError(node, "wrong number of arguments. got=%d, want=%d",
			len(args), count)
	}
	return nil
}

func AssertArgsType(node ast.Node, args []object.Object, types ...object.ObjectType) *object.Error {
	if len(args) != len(types) {
		return NewError(node, "wrong number of arguments. got=%d, want=%d",
			len(args), len(types))
	}
	for i, t := range types {
		if args[i].Type() != t {
			return NewError(node, "wrong type of argument n°%d. got=%s, want=%s",
				i+1, args[i].Type(), t)
		}
	}
	return nil
}

var globalBuiltins = map[string]*object.Builtin{
	"exit": {
		Fn: func(node ast.Node, args ...object.Object) object.Object {
			if len(args) > 0 {
				interrupt(args[0])
			} else {
				interrupt(nil)
			}
			return nil
		},
	},
	"assert": {
		Fn: func(node ast.Node, args ...object.Object) object.Object {
			if oErr := AssertArgsType(node, args, object.BOOLEAN_OBJ); oErr != nil {
				return oErr
			}

			if b := args[0].(*object.Boolean); !b.Value {
				interrupt(&object.Error{
					Message:    "Assertion failed",
					StackToken: []token.Token{node.TokenDetails()},
				})
			}
			return TRUE
		},
	},
	"pause": {
		Fn: func(node ast.Node, args ...object.Object) object.Object {
			if oErr := AssertArgsType(node, args, object.STRING_OBJ); oErr != nil {
				return oErr
			}

			dArg := args[0].(*object.String)
			d, err := time.ParseDuration(dArg.Value)
			if err != nil {
				return NewError(node, err.Error())
			}
			time.Sleep(d)

			return NULL
		},
	},
	"len": {
		Fn: func(node ast.Node, args ...object.Object) object.Object {
			if oErr := AssertArgCount(node, args, 1); oErr != nil {
				return oErr
			}

			switch arg := args[0].(type) {
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			default:
				return NewError(node, "wrong type of argument. got=%s, want=%s or %s",
					args[0].Type(), object.ARRAY_OBJ, object.STRING_OBJ)
			}
		},
	},
	"puts": {
		Fn: func(node ast.Node, args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}

			return NULL
		},
	},
	"first": {
		Fn: func(node ast.Node, args ...object.Object) object.Object {
			if oErr := AssertArgCount(node, args, 1); oErr != nil {
				return oErr
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return NewError(node, "argument to `first` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[0]
			}

			return NULL
		},
	},
	"last": {
		Fn: func(node ast.Node, args ...object.Object) object.Object {
			if oErr := AssertArgCount(node, args, 1); oErr != nil {
				return oErr
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return NewError(node, "argument to `last` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			if length > 0 {
				return arr.Elements[length-1]
			}

			return NULL
		},
	},
	"rest": {
		Fn: func(node ast.Node, args ...object.Object) object.Object {
			if oErr := AssertArgCount(node, args, 1); oErr != nil {
				return oErr
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return NewError(node, "argument to `rest` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			if length > 0 {
				newElements := make([]object.Object, length-1, length-1)
				copy(newElements, arr.Elements[1:length])
				return &object.Array{Elements: newElements}
			}

			return NULL
		},
	},
	"push": {
		Fn: func(node ast.Node, args ...object.Object) object.Object {
			if oErr := AssertArgCount(node, args, 2); oErr != nil {
				return oErr
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return NewError(node, "argument to `push` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)

			newElements := make([]object.Object, length+1, length+1)
			copy(newElements, arr.Elements)
			newElements[length] = args[1]

			return &object.Array{Elements: newElements}
		},
	},
}