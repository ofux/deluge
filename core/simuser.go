package core

import (
	"github.com/ofux/deluge/cleanhttp"
	"github.com/ofux/deluge/core/recording"
	"github.com/ofux/deluge/dsl/evaluator"
	"github.com/ofux/deluge/dsl/object"
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
			Pairs: make(map[object.HashKey]object.Object),
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
	// Inject ARGS into first param
	if len(params) >= 1 {
		env.Add(params[0].Value, su.scenario.scriptArgs)
	}
	// Inject SESSION into second param
	if len(params) >= 2 {
		env.Add(params[1].Value, su.session)
	}
	return env
}
