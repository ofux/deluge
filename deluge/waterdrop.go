package deluge

import (
	"errors"
	"github.com/ofux/deluge-dsl/ast"
	"github.com/ofux/deluge-dsl/evaluator"
	"github.com/ofux/deluge-dsl/object"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var halt = errors.New("Halt water drop")

type WaterDropStatus int

const (
	Virgin WaterDropStatus = iota
	InProgress
	DoneSuccess
	DoneAssertionError
)

type WaterDrop struct {
	Name          string
	script        ast.Node
	evaluator     *evaluator.Evaluator
	client        *http.Client
	Status        WaterDropStatus
	SleepDuration time.Duration
}

func NewWaterDrop(name string, script ast.Node) *WaterDrop {
	wd := &WaterDrop{
		Name:      name,
		script:    script,
		evaluator: evaluator.NewEvaluator(),
		client:    http.DefaultClient,
		Status:    Virgin,
	}

	wd.evaluator.AddBuiltin("assert", wd.Assert)
	wd.evaluator.AddBuiltin("pause", wd.Pause)
	wd.evaluator.AddBuiltin("http", wd.DoHTTP)

	return wd
}

func (wd *WaterDrop) Run() {
	defer func() {
		if caught := recover(); caught != nil {
			if caught == halt {
				log.Debug(halt.Error())
				return
			}
			panic(caught) // Something else happened, repanic!
		}
	}()

	wd.Status = InProgress
	env := object.NewEnvironment()
	evaluated := wd.evaluator.Eval(wd.script, env)
	if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
		log.Errorln(evaluated.Inspect())
		if errObj, ok := evaluated.(*object.Error); ok {
			log.Fatal(errObj.Message)
		}
	}
	if wd.Status == InProgress {
		wd.Status = DoneSuccess
	}
}

//type BuiltinFunction func(node ast.Node, args ...Object) Object

func (wd *WaterDrop) Assert(node ast.Node, args ...object.Object) object.Object {
	checkArgsTypeFatal(node, args, 0, object.BOOLEAN_OBJ)
	result := args[0].(*object.Boolean)

	if !result.Value {
		log.Debugf("Assertion failed at %s", ast.PrintLocation(node))
		wd.Status = DoneAssertionError
		// TODO: exit/interrupt
	}
	return evaluator.NULL
}

func (wd *WaterDrop) Pause(node ast.Node, args ...object.Object) object.Object {
	checkArgsTypeFatal(node, args, 0, object.STRING_OBJ)

	dArg := args[0].(*object.String)
	d, err := time.ParseDuration(dArg.Value)
	checkFatal(node, err)
	wd.SleepDuration += d
	time.Sleep(d)
	return evaluator.NULL
}

func (wd *WaterDrop) DoHTTP(node ast.Node, args ...object.Object) object.Object {
	checkArgsTypeFatal(node, args, 0, object.HASH_OBJ)

	jsReq := args[0].(*object.Hash)

	jsUrl, ok := jsReq.Get("url")
	if !ok {
		log.Fatalf("Invalid HTTP request: missing 'url' field at %s", ast.PrintLocation(node))
	}
	checkTypeFatal(node, jsUrl.Value, object.STRING_OBJ)
	url := jsUrl.Value.(*object.String).Value

	var method = "GET"
	jsMethod, ok := jsReq.Get("method")
	if ok {
		checkTypeFatal(node, jsMethod.Value, object.STRING_OBJ)
		method = jsMethod.Value.(*object.String).Value
	}

	req, err := http.NewRequest(method, url, nil)
	checkFatal(node, err)

	log.Debugf("Performing HTTP request: %s %s", req.Method, req.URL.String())
	start := time.Now()
	//res, err := wd.client.Do(req)
	end := time.Now()
	duration := end.Sub(start)
	if err != nil {
		log.Debugf("Request error: %s", err.Error())
	} else {
		log.Debugf("Response status: %s in %s", "res.Status", duration.String())
	}

	return evaluator.NULL
}

func checkFatal(node ast.Node, err error) {
	if err != nil {
		log.Fatalf("%s at %s\n", err.Error(), ast.PrintLocation(node))
	}
}

func checkTypeFatal(node ast.Node, obj object.Object, expectedType object.ObjectType) {
	if obj == nil || obj.Type() != expectedType {
		log.Fatalf("%s: expected type %s, got %s\n", ast.PrintLocation(node), expectedType, obj.Type())
	}
}

func checkArgsTypeFatal(node ast.Node, args []object.Object, argIndex int, expectedType object.ObjectType) {
	if len(args) <= argIndex {
		log.Fatalf("Expected at least %d arguments at %s\n", argIndex+1, ast.PrintLocation(node))
	}
	if args[argIndex] == nil || args[argIndex].Type() != expectedType {
		log.Fatalf("%s: expected argument n°%d to be of type %s, got %s\n", ast.PrintLocation(node), argIndex+1, expectedType, args[argIndex].Type())
	}
}
