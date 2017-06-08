package deluge

import (
	"github.com/ofux/deluge-dsl/ast"
	"github.com/ofux/deluge-dsl/evaluator"
	"github.com/ofux/deluge-dsl/object"
	"github.com/ofux/deluge/deluge/recording"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type SimUserStatus int

const (
	Virgin SimUserStatus = iota
	InProgress
	DoneSuccess
	DoneError
)

type SimUser struct {
	Name          string
	scenario      *Scenario
	evaluator     *evaluator.Evaluator
	client        *http.Client
	Status        SimUserStatus
	SleepDuration time.Duration
	httpRecorder  *recording.HTTPRecorder
	log           *log.Entry
	iteration     int
}

func NewSimUser(name string, scenario *Scenario) *SimUser {
	su := &SimUser{
		Name:      name,
		Status:    Virgin,
		scenario:  scenario,
		evaluator: evaluator.NewEvaluator(),
		client:    http.DefaultClient,

		httpRecorder: scenario.httpRecorder,
		log: scenario.log.WithFields(log.Fields{
			"user": name,
		}),
	}

	if err := su.evaluator.AddBuiltin("http", su.ExecHTTPRequest); err != nil {
		log.Fatal(err.Error())
	}

	return su
}

func (su *SimUser) Run(iteration int) {
	su.iteration = iteration
	su.Status = InProgress
	env := object.NewEnvironment()
	evaluated := su.evaluator.Eval(su.scenario.script, env)

	if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
		su.log.Errorln(evaluated.Inspect())
		su.Status = DoneError
		return
	}

	if su.Status == InProgress {
		su.Status = DoneSuccess
	}
}

func (su *SimUser) ExecHTTPRequest(node ast.Node, args ...object.Object) object.Object {
	if oErr := evaluator.AssertArgsType(node, args, object.STRING_OBJ, object.HASH_OBJ); oErr != nil {
		return oErr
	}

	reqName := args[0].(*object.String).Value
	reqObj := args[1].(*object.Hash)

	jsUrl, ok := reqObj.Get("url")
	if !ok {
		return evaluator.NewError(node, "invalid HTTP request: missing 'url' field")
	}
	url, ok := jsUrl.Value.(*object.String)
	if !ok {
		return evaluator.NewError(node, "invalid HTTP request: 'url' should be a STRING")
	}

	var method = "GET"
	if methodField, ok := reqObj.Get("method"); ok {
		if methodFieldVal, ok := methodField.Value.(*object.String); ok {
			method = methodFieldVal.Value
		}
	}

	req, err := http.NewRequest(method, url.Value, nil)
	if err != nil {
		return evaluator.NewError(node, err.Error())
	}

	su.log.Debugf("Performing HTTP request: %s %s", req.Method, req.URL.String())
	start := time.Now()
	res, err := su.client.Do(req)
	end := time.Now()
	duration := end.Sub(start)

	if err != nil {
		su.log.Debugf("Request error: %s", err.Error())
		return evaluator.NewError(node, err.Error())
	} else {
		su.log.Debugf("Response status: %s in %s", "res.Status", duration.String())
		su.httpRecorder.Record(&recording.HTTPRecordEntry{
			Iteration:  su.iteration,
			Name:       reqName,
			Value:      duration.Nanoseconds() / 1000,
			StatusCode: res.StatusCode,
		})
	}

	return evaluator.NULL
}
