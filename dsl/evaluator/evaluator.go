package evaluator

import (
	"errors"
	"fmt"
	"github.com/ofux/deluge/dsl/ast"
	"github.com/ofux/deluge/dsl/object"
	"github.com/ofux/deluge/dsl/token"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

type Evaluator struct {
	builtins map[string]*object.Builtin
}

type evalInterruption struct {
	returnedVal object.Object
}

func NewEvaluator() *Evaluator {
	ev := &Evaluator{
		builtins: make(map[string]*object.Builtin),
	}
	return ev
}

func (e *Evaluator) AddBuiltin(name string, fn object.BuiltinFunction) error {
	if _, ok := e.builtins[name]; ok {
		return errors.New(fmt.Sprintf("Bult-in function '%s' is already defined", name))
	}
	e.builtins[name] = &object.Builtin{Fn: fn}
	return nil
}

func (e *Evaluator) Eval(node ast.Node, env *object.Environment) (returnedVal object.Object) {
	defer func() {
		if r := recover(); r != nil {
			if interruption, ok := r.(evalInterruption); ok {
				returnedVal = interruption.returnedVal
			} else {
				panic(r) // Something else happened, repanic!
			}
		}
	}()
	returnedVal = e.eval(node, env)
	return
}

func interrupt(arg object.Object) {
	var interruption evalInterruption
	if arg != nil {
		interruption = evalInterruption{
			returnedVal: arg,
		}
	} else {
		interruption = evalInterruption{
			returnedVal: NULL,
		}
	}
	panic(interruption)
}

func (e *Evaluator) eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	// Statements
	case *ast.Program:
		return e.evalProgram(node, env)

	case *ast.BlockStatement:
		env := object.NewEnclosedEnvironment(env)
		return e.evalBlockStatement(node, env)

	case *ast.ExpressionStatement:
		return e.eval(node.Expression, env)

	case *ast.ReturnStatement:
		val := e.eval(node.ReturnValue, env)
		if IsError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.LetStatement:
		val := e.eval(node.Value, env)
		if IsError(val) {
			return val
		}
		if !env.Add(node.Name.Value, val) {
			return NewError(node.Name, "variable %s redeclared in this block", node.Name.Value)
		}

	case *ast.AssignStatement:
		return e.evalAssignStatement(node, env)

	// Expressions
	case *ast.Null:
		return NULL

	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)

	case *ast.PrefixExpression:
		right := e.eval(node.Right, env)
		if IsError(right) {
			return right
		}
		return e.evalPrefixExpression(node, right)

	case *ast.InfixExpression:
		return e.evalInfixExpression(node, env)

	case *ast.IfStatement:
		return e.evalIfStatement(node, env)

	case *ast.ForStatement:
		return e.evalForStatement(node, env)

	case *ast.Identifier:
		return e.evalIdentifier(node, env)

	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}

	case *ast.CallExpression:
		function := e.eval(node.Function, env)
		if IsError(function) {
			return function
		}

		args := e.evalExpressions(node.Arguments, env)
		if len(args) == 1 && IsError(args[0]) {
			return args[0]
		}

		funcResult := e.applyFunction(node, function, args)
		if IsError(funcResult) {
			funcErr := funcResult.(*object.Error)
			funcErr.AddCallToStack(node)
		}
		return funcResult

	case *ast.ArrayLiteral:
		elements := e.evalExpressions(node.Elements, env)
		if len(elements) == 1 && IsError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}

	case *ast.IndexExpression:
		left := e.eval(node.Left, env)
		if IsError(left) {
			return left
		}
		index := e.eval(node.Index, env)
		if IsError(index) {
			return index
		}
		return e.evalIndexExpression(node, left, index)

	case *ast.HashLiteral:
		return e.evalHashLiteral(node, env)

	}

	return nil
}

func (e *Evaluator) evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = e.eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func (e *Evaluator) evalBlockStatement(
	block *ast.BlockStatement,
	env *object.Environment,
) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = e.eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func (e *Evaluator) evalPrefixExpression(prefixExp *ast.PrefixExpression, right object.Object) object.Object {
	operator := prefixExp.Operator
	switch operator {
	case "!":
		return e.evalBangOperatorExpression(right)
	case "-":
		return e.evalMinusPrefixOperatorExpression(prefixExp, right)
	default:
		return NewError(prefixExp, "unknown operator: %s%s", operator, right.Type())
	}
}

func (e *Evaluator) evalInfixExpression(
	node *ast.InfixExpression,
	env *object.Environment,
) object.Object {
	operator := node.Operator

	// Eval && and || operators first, because right and left must be e.evaluated under certain circumstances only
	if operator == "&&" || operator == "||" {
		return e.evalBooleanInfixExpression(node, env)
	}

	// Eval any other operators
	left := e.eval(node.Left, env)
	if IsError(left) {
		return left
	}
	right := e.eval(node.Right, env)
	if IsError(right) {
		return right
	}

	switch {
	case object.IsInteger(left) && object.IsInteger(right):
		return e.evalIntegerInfixExpression(node, left, right)
	case object.IsNumeric(left) && object.IsNumeric(right):
		return e.evalFloatInfixExpression(node, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return e.evalStringInfixExpression(node, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return NewError(node, "type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	default:
		return NewError(node, "unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalIntegerReassignStatement(
	node *ast.AssignStatement,
	env *object.Environment,
	curVal object.Object,
) object.Object {
	operator := node.Operator
	val := curVal.(*object.Integer)

	switch operator {
	case "++":
		val.Value++
		return nil
	case "--":
		val.Value--
		return nil
	}

	right := e.eval(node.Value, env)
	if IsError(right) {
		return right
	}
	rightInt, ok := right.(*object.Integer)
	if !ok {
		return NewError(node.Name, "unknown operator %s %s %s", val.Type(), operator, right.Type())
	}

	switch operator {
	case "+=":
		val.Value += rightInt.Value
	case "-=":
		val.Value -= rightInt.Value
	case "*=":
		val.Value *= rightInt.Value
	case "/=":
		val.Value /= rightInt.Value
	default:
		return NewError(node.Name, "unknown operator %s %s %s", val.Type(), operator, right.Type())
	}
	return nil
}

func (e *Evaluator) evalStringReassignStatement(
	node *ast.AssignStatement,
	env *object.Environment,
	curVal object.Object,
) object.Object {
	operator := node.Operator
	val := curVal.(*object.String)

	right := e.eval(node.Value, env)
	if IsError(right) {
		return right
	}
	rightStr, ok := right.(*object.String)
	if !ok {
		return NewError(node.Name, "unknown operator %s %s %s", val.Type(), operator, right.Type())
	}

	switch operator {
	case "+=":
		val.Value += rightStr.Value
	default:
		return NewError(node.Name, "unknown operator %s %s %s", val.Type(), operator, right.Type())
	}
	return nil
}

func (e *Evaluator) evalAssignStatement(
	node *ast.AssignStatement,
	env *object.Environment,
) object.Object {
	operator := node.Operator

	// Simple assignment => set variable's value directly
	if operator == "=" {
		val := e.eval(node.Value, env)
		if IsError(val) {
			return val
		}
		if !env.Set(node.Name.Value, val) {
			return NewError(node.Name, "identifier not found: %s", node.Name.Value)
		}
		return nil
	}

	// Re-assignment => get current variable's value to compute the new one
	val, ok := env.Get(node.Name.Value)
	if !ok {
		return NewError(node.Name, "identifier not found: %s", node.Name.Value)
	}
	if IsError(val) {
		return val
	}

	switch {
	case val.Type() == object.INTEGER_OBJ:
		return e.evalIntegerReassignStatement(node, env, val)
	case val.Type() == object.STRING_OBJ:
		return e.evalStringReassignStatement(node, env, val)
	}

	return NewError(node.Name, "unknown operator %s %s", val.Type(), operator)
}

func (e *Evaluator) evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func (e *Evaluator) evalMinusPrefixOperatorExpression(node ast.Node, right object.Object) object.Object {
	switch right := right.(type) {
	case *object.Integer:
		return &object.Integer{Value: -right.Value}
	case *object.Float:
		return &object.Float{Value: -right.Value}
	default:
		return NewError(node, "unknown operator: -%s", right.Type())
	}
}

func (e *Evaluator) evalIntegerInfixExpression(
	node *ast.InfixExpression,
	left, right object.Object,
) object.Object {
	operator := node.Operator
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "%":
		return &object.Integer{Value: leftVal % rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return NewError(node, "unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalFloatInfixExpression(
	node *ast.InfixExpression,
	left, right object.Object,
) object.Object {
	operator := node.Operator
	var leftVal, rightVal float64
	switch left := left.(type) {
	case *object.Integer:
		leftVal = float64(left.Value)
	case *object.Float:
		leftVal = left.Value
	}
	switch right := right.(type) {
	case *object.Integer:
		rightVal = float64(right.Value)
	case *object.Float:
		rightVal = right.Value
	}

	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "/":
		return &object.Float{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return NewError(node, "unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalStringInfixExpression(
	node *ast.InfixExpression,
	left, right object.Object,
) object.Object {
	operator := node.Operator
	if operator != "+" {
		return NewError(node, "unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}

	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}
}

func (e *Evaluator) evalBooleanInfixExpression(
	node *ast.InfixExpression,
	env *object.Environment,
) object.Object {
	operator := node.Operator
	switch {
	case operator == "&&":
		left := e.eval(node.Left, env)
		if IsError(left) {
			return left
		}
		if left.Type() != object.BOOLEAN_OBJ {
			return NewError(node, "unknown operator: %s %s %s",
				left.Type(), operator, object.BOOLEAN_OBJ)
		}
		if !left.(*object.Boolean).Value {
			return nativeBoolToBooleanObject(false)
		}
		// left is true, so let's eval right
		right := e.eval(node.Right, env)
		if IsError(right) {
			return right
		}
		if right.Type() != object.BOOLEAN_OBJ {
			return NewError(node, "unknown operator: %s %s %s",
				object.BOOLEAN_OBJ, operator, right.Type())
		}
		return nativeBoolToBooleanObject(right.(*object.Boolean).Value)

	case operator == "||":
		left := e.eval(node.Left, env)
		if IsError(left) {
			return left
		}
		if left.Type() != object.BOOLEAN_OBJ {
			return NewError(node, "unknown operator: %s %s %s",
				left.Type(), operator, object.BOOLEAN_OBJ)
		}
		if left.(*object.Boolean).Value {
			return nativeBoolToBooleanObject(true)
		}
		// left is false, so let's eval right
		right := e.eval(node.Right, env)
		if IsError(right) {
			return right
		}
		if right.Type() != object.BOOLEAN_OBJ {
			return NewError(node, "unknown operator: %s %s %s",
				object.BOOLEAN_OBJ, operator, right.Type())
		}
		return nativeBoolToBooleanObject(right.(*object.Boolean).Value)
	}

	panic(errors.New(fmt.Sprintf("evalBooleanInfixExpression has been called with operator %s", operator)))
}

func (e *Evaluator) evalIfStatement(
	ie *ast.IfStatement,
	env *object.Environment,
) object.Object {
	condition := e.eval(ie.Condition, env)
	if IsError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return e.eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return e.eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

func (e *Evaluator) evalForStatement(
	fo *ast.ForStatement,
	env *object.Environment,
) object.Object {
	env = object.NewEnclosedEnvironment(env)

	init := e.eval(fo.Initialization, env)
	if IsError(init) {
		return init
	}
	condition := e.eval(fo.Condition, env)
	if IsError(condition) {
		return condition
	}

	var loop object.Object = NULL

	for isTruthy(condition) {
		loop = e.eval(fo.Loop, env)
		if IsError(loop) {
			return loop
		}
		afterthought := e.eval(fo.Afterthought, env)
		if IsError(afterthought) {
			return afterthought
		}
		condition = e.eval(fo.Condition, env)
		if IsError(condition) {
			return condition
		}
	}

	return loop
}

func (e *Evaluator) evalIdentifier(
	node *ast.Identifier,
	env *object.Environment,
) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := e.builtins[node.Value]; ok {
		return builtin
	}

	if builtin, ok := globalBuiltins[node.Value]; ok {
		return builtin
	}

	return NewError(node, "identifier not found: "+node.Value)
}

func (e *Evaluator) evalExpressions(
	exps []ast.Expression,
	env *object.Environment,
) []object.Object {
	var result []object.Object

	for _, exp := range exps {
		evaluated := e.eval(exp, env)
		if IsError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func (e *Evaluator) applyFunction(node ast.Node, fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {

	case *object.Function:
		if len(fn.Parameters) != len(args) {
			return NewError(node, "wrong number of arguments: expected %d, got %d", len(fn.Parameters), len(args))
		}
		extendedEnv, err := extendFunctionEnv(fn, args)
		if err != nil {
			return err
		}
		evaluated := e.eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)

	case *object.Builtin:
		return fn.Fn(node, args...)

	default:
		return NewError(node, "not a function: %s", fn.Type())
	}
}

func (e *Evaluator) evalIndexExpression(node ast.Node, left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return e.evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return e.evalHashIndexExpression(node, left, index)
	default:
		return NewError(node, "index operator not supported: %s", left.Type())
	}
}

func (e *Evaluator) evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > max {
		return NULL
	}

	return arrayObject.Elements[idx]
}

func (e *Evaluator) evalHashLiteral(
	node *ast.HashLiteral,
	env *object.Environment,
) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := e.eval(keyNode, env)
		if IsError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return NewError(keyNode, "unusable as hash key: %s", key.Type())
		}

		value := e.eval(valueNode, env)
		if IsError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

func (e *Evaluator) evalHashIndexExpression(node ast.Node, hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return NewError(node, "unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func NewError(node ast.Node, format string, a ...interface{}) *object.Error {
	return &object.Error{
		Message: fmt.Sprintf(format, a...),
		StackToken: []token.Token{
			node.TokenDetails(),
		},
	}
}

func IsError(obj object.Object) bool {
	return obj != nil && obj.Type() == object.ERROR_OBJ
}

func extendFunctionEnv(
	fn *object.Function,
	args []object.Object,
) (*object.Environment, *object.Error) {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		if !env.Add(param.Value, args[paramIdx]) {
			return nil, NewError(param, "")
		}
	}

	return env, nil
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}
