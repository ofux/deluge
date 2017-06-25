package core

import (
	"github.com/ofux/deluge-dsl/ast"
	"github.com/ofux/deluge-dsl/evaluator"
	"github.com/ofux/deluge-dsl/object"
	"github.com/ofux/deluge/cleanhttp"
	"github.com/ofux/deluge/core/recording"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type simUserStatus int

const (
	UserVirgin simUserStatus = iota
	UserInProgress
	UserDoneSuccess
	UserDoneError
	UserInterrupted
)

type simUser struct {
	name          string
	scenario      *Scenario
	evaluator     *evaluator.Evaluator
	client        *http.Client
	sleepDuration time.Duration
	httpRecorder  *recording.HTTPRecorder
	log           *log.Entry
	iteration     int

	status    simUserStatus
	execError *object.Error
}

func newSimUser(name string, scenario *Scenario) *simUser {
	su := &simUser{
		name:      name,
		status:    UserVirgin,
		scenario:  scenario,
		evaluator: evaluator.NewEvaluator(),
		client:    cleanhttp.DefaultClient(),

		httpRecorder: scenario.httpRecorder,
		log: scenario.log.WithFields(log.Fields{
			"user": name,
		}),
	}

	if err := su.evaluator.AddBuiltin("http", su.execHTTPRequest); err != nil {
		log.Fatal(err.Error())
	}

	return su
}

func (su *simUser) run(iteration int) {
	su.iteration = iteration
	su.status = UserInProgress
	env := object.NewEnvironment()
	evaluated := su.evaluator.Eval(su.scenario.script, env)

	su.client.Transport.(*http.Transport).CloseIdleConnections()

	if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
		su.status = UserDoneError
		su.execError = evaluated.(*object.Error)
		su.log.Errorln(evaluated.Inspect())
		return
	}

	if su.status == UserInProgress {
		su.status = UserDoneSuccess
	}
}

func (su *simUser) execHTTPRequest(node ast.Node, args ...object.Object) object.Object {
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
		defer res.Body.Close()
		su.log.Debugf("Response status: %s in %s", "res.Status", duration.String())
		su.httpRecorder.Record(&recording.HTTPRecordEntry{
			Iteration:  su.iteration,
			Name:       reqName,
			Value:      duration.Nanoseconds() / 100000,
			StatusCode: res.StatusCode,
		})
	}

	return evaluator.NULL
}
