package core

import (
	"github.com/ofux/deluge/cleanhttp"
	"github.com/ofux/deluge/core/recording"
	"github.com/ofux/deluge/dsl/ast"
	"github.com/ofux/deluge/dsl/evaluator"
	"github.com/ofux/deluge/dsl/object"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
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
	session       *object.Hash

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
		session: &object.Hash{
			Pairs: make(map[object.HashKey]object.HashPair),
		},

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
	env := su.createEnvironment()
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

func (su *simUser) createEnvironment() *object.Environment {
	env := object.NewEnvironment()
	params := su.scenario.scriptParams
	// Add ARGS
	if len(params) >= 1 {
		env.Add(params[0].Value, su.scenario.scriptArgs)
	}
	// Add SESSION
	if len(params) >= 2 {
		env.Add(params[1].Value, su.session)
	}
	return env
}

func (su *simUser) execHTTPRequest(node ast.Node, args ...object.Object) object.Object {
	if oErr := evaluator.AssertArgsType(node, args, object.STRING_OBJ, object.HASH_OBJ); oErr != nil {
		return oErr
	}
	reqName := args[0].(*object.String).Value
	reqObj := args[1].(*object.Hash)

	url, _, err := reqObj.GetAsString("url")
	if err != nil {
		return evaluator.NewError(node, "invalid HTTP request: %s", err.Error())
	}

	var method = "GET"
	if m, ok, err := reqObj.GetAsString("method"); ok {
		if err != nil {
			return evaluator.NewError(node, "invalid HTTP request: %s", err.Error())
		}
		method = m.Value
	}

	var body io.Reader
	if b, ok, err := reqObj.GetAsString("body"); ok {
		if err != nil {
			return evaluator.NewError(node, "invalid HTTP request: %s", err.Error())
		}
		body = strings.NewReader(b.Value)
	}

	// Create request
	req, err := http.NewRequest(method, url.Value, body)
	if err != nil {
		return evaluator.NewError(node, err.Error())
	}

	if headers, ok, err := reqObj.GetAsHash("headers"); ok {
		if err != nil {
			return evaluator.NewError(node, "invalid HTTP request: %s", err.Error())
		}
		for headerKey, headerVal := range headers.Pairs {
			headerValStr, ok := headerVal.Value.(*object.String)
			if !ok {
				return evaluator.NewError(node, "invalid HTTP header '%s': should be of type %s but was %s", headerKey, object.STRING_OBJ, headerVal.Value.Type())
			}
			req.Header.Add(string(headerKey), headerValStr.Value)
		}
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
