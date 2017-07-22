package core

import (
	"github.com/ofux/deluge/core/recording"
	"github.com/ofux/deluge/core/recording/recordingtest"
	"github.com/ofux/deluge/dsl/object"
	"github.com/ofux/docilemonkey/docilemonkey"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestScenario_Run(t *testing.T) {

	// discard DSL logs for testing
	logger := log.New()
	logger.Out = ioutil.Discard
	logTest := logger.WithField("test", true)

	t.Run("Run simple scenario with HTTP request", func(t *testing.T) {

		srv := httptest.NewServer(http.HandlerFunc(docilemonkey.Handler))
		defer srv.Close()

		const reqName = "My request"
		program, params := compileTest(t, `
		http("`+reqName+`", {
			"url": "`+srv.URL+`/hello/toto?s=201",
			"method": "POST"
		});
		`)

		scenario := newScenario("foo", 50, 50*time.Millisecond, program, params, nil, logTest)
		scenario.run(200*time.Millisecond, nil)

		records, err := scenario.httpRecorder.GetRecords()
		if err != nil {
			t.Fatalf(err.Error())
		}

		iterCount := len(records.PerIteration)
		if iterCount < 1 {
			t.Fatalf("Expected to have at least %d iterations, got %d", 1, iterCount)
		}
		assert.Equal(t, uint64(50), scenario.EffectiveUserCount)
		if scenario.EffectiveExecCount < 1 {
			t.Fatalf("Expected to have at least %d executions, got %d", 1, scenario.EffectiveExecCount)
		}
		recordingtest.CheckHTTPRecord(t, records.Global, reqName, int64(scenario.EffectiveExecCount), 201, recording.Ok)
		for i, record := range records.PerIteration {
			if i < iterCount-1 {
				recordingtest.CheckHTTPRecord(t, record, reqName, 1, 201, recording.Ok)
			}
		}
	})

	t.Run("Run scenario with session", func(t *testing.T) {

		srv := httptest.NewServer(http.HandlerFunc(docilemonkey.Handler))
		defer srv.Close()

		program, params := compileTest(t, `
		let c = session["count"];
		if (c == null) {
			c = 1;
		} else {
			c++;
		}
		session["count"] = c;
		assert(c < 3);
		`)

		scenario := newScenario("foo", 5, 10*time.Millisecond, program, params, nil, logTest)
		scenario.run(20000*time.Millisecond, nil)

		assert.Equal(t, uint64(5), scenario.EffectiveUserCount)
		assert.Equal(t, uint64(15), scenario.EffectiveExecCount)
	})

	t.Run("Run scenario without error, with too short iterations", func(t *testing.T) {

		srv := httptest.NewServer(http.HandlerFunc(docilemonkey.Handler))
		defer srv.Close()

		program, params := compileTest(t, `
		pause("50ms");
		`)

		scenario := newScenario("foo", 50, 1*time.Millisecond, program, params, nil, logTest)
		scenario.run(200*time.Millisecond, nil)

		assert.Equal(t, uint64(50), scenario.EffectiveUserCount)
		if scenario.EffectiveExecCount < 50 {
			t.Fatalf("Expected to have at least %d executions, got %d", 50, scenario.EffectiveExecCount)
		}
	})

	t.Run("Run scenario with args", func(t *testing.T) {

		srv := httptest.NewServer(http.HandlerFunc(docilemonkey.Handler))
		defer srv.Close()

		const reqName = "My request"
		program, params := compileTest(t, `
		http("`+reqName+`", {
			"url": args["baseUrl"] + "/hello/toto?s=500",
			"method": args["method"]
		});
		`)
		scriptArgs := &object.Hash{
			Pairs: map[object.HashKey]object.Object{
				"baseUrl": &object.String{srv.URL},
				"method":  &object.String{"PUT"},
			},
			IsImmutable: true,
		}

		scenario := newScenario("foo", 50, 50*time.Millisecond, program, params, scriptArgs, logTest)
		scenario.run(200*time.Millisecond, nil)

		records, err := scenario.httpRecorder.GetRecords()
		if err != nil {
			t.Fatalf(err.Error())
		}

		iterCount := len(records.PerIteration)
		if iterCount < 1 {
			t.Fatalf("Expected to have at least %d iterations, got %d", 1, iterCount)
		}
		assert.Equal(t, uint64(50), scenario.EffectiveUserCount)
		if scenario.EffectiveExecCount < 1 {
			t.Fatalf("Expected to have at least %d executions, got %d", 1, scenario.EffectiveExecCount)
		}
		recordingtest.CheckHTTPRecord(t, records.Global, reqName, int64(scenario.EffectiveExecCount), 500, recording.Ko)
		for i, record := range records.PerIteration {
			if i < iterCount-1 {
				recordingtest.CheckHTTPRecord(t, record, reqName, 1, 500, recording.Ko)
			}
		}
	})

	t.Run("Run scenario with args and try to modify it", func(t *testing.T) {
		program, params := compileTest(t, `
		args["method"] = "foobar"
		`)
		scriptArgs := &object.Hash{
			Pairs: map[object.HashKey]object.Object{
				"method": &object.String{"PUT"},
			},
			IsImmutable: true,
		}

		scenario := newScenario("foo", 50, 50*time.Millisecond, program, params, scriptArgs, logTest)
		scenario.run(200*time.Millisecond, nil)

		assert.Equal(t, ScenarioDoneError, scenario.Status)
		assert.Len(t, scenario.Errors, 50)
		assert.Equal(t, "hash is immutable, you cannot modify it", scenario.Errors[0].Message)
	})

	t.Run("Run scenario with error", func(t *testing.T) {

		srv := httptest.NewServer(http.HandlerFunc(docilemonkey.Handler))
		defer srv.Close()

		program, params := compileTest(t, `
		doesntexists();
		`)

		scenario := newScenario("foo", 50, 1*time.Millisecond, program, params, nil, logTest)
		scenario.run(200*time.Millisecond, nil)

		if len(scenario.Errors) != 50 {
			t.Fatalf("Expected to have %d errors, got %d", 50, len(scenario.Errors))
		}
		for _, err := range scenario.Errors {
			if err.Message != "identifier not found: doesntexists" {
				t.Errorf("Wrong error message. Got '%s'", err.Message)
			}
		}

		assert.Equal(t, uint64(50), scenario.EffectiveUserCount)
		assert.Equal(t, uint64(50), scenario.EffectiveExecCount)
	})
}
