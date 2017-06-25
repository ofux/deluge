package core

import (
	"github.com/ofux/deluge/core/recording"
	"github.com/ofux/deluge/core/recording/recordingtest"
	"github.com/ofux/docilemonkey/docilemonkey"
	log "github.com/sirupsen/logrus"
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

	t.Run("Run scenario without error", func(t *testing.T) {

		srv := httptest.NewServer(http.HandlerFunc(docilemonkey.Handler))
		defer srv.Close()

		const reqName = "My request"
		program := compileTest(t, `
		http("`+reqName+`", {
			"url": "`+srv.URL+`/hello/toto?s=201",
			"method": "POST"
		});
		`)

		scenario := newScenario("foo", 50, 50*time.Millisecond, program, logTest)
		scenario.run(200*time.Millisecond, nil)

		records, err := scenario.httpRecorder.GetRecords()
		if err != nil {
			t.Fatalf(err.Error())
		}

		iterCount := len(records.PerIteration)
		if iterCount < 1 {
			t.Fatalf("Expected to have at least %d iterations, got %d", 1, iterCount)
		}
		if scenario.EffectiveUserCount != 50 {
			t.Fatalf("Expected to have %d simulated users, got %d", 50, scenario.EffectiveUserCount)
		}
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

	t.Run("Run scenario without error, with too short iterations", func(t *testing.T) {

		srv := httptest.NewServer(http.HandlerFunc(docilemonkey.Handler))
		defer srv.Close()

		program := compileTest(t, `
		pause("50ms");
		`)

		scenario := newScenario("foo", 50, 1*time.Millisecond, program, logTest)
		scenario.run(200*time.Millisecond, nil)

		if scenario.EffectiveUserCount != 50 {
			t.Fatalf("Expected to have %d simulated users, got %d", 50, scenario.EffectiveUserCount)
		}
		if scenario.EffectiveExecCount < 50 {
			t.Fatalf("Expected to have at least %d executions, got %d", 50, scenario.EffectiveExecCount)
		}
	})

	t.Run("Run scenario with error", func(t *testing.T) {

		srv := httptest.NewServer(http.HandlerFunc(docilemonkey.Handler))
		defer srv.Close()

		program := compileTest(t, `
		doesntexists();
		`)

		scenario := newScenario("foo", 50, 1*time.Millisecond, program, logTest)
		scenario.run(200*time.Millisecond, nil)

		if len(scenario.Errors) != 50 {
			t.Fatalf("Expected to have %d errors, got %d", 50, len(scenario.Errors))
		}
		for _, err := range scenario.Errors {
			if err.Message != "identifier not found: doesntexists" {
				t.Errorf("Wrong error message. Got '%s'", err.Message)
			}
		}

		if scenario.EffectiveUserCount != 50 {
			t.Fatalf("Expected to have %d simulated users, got %d", 50, scenario.EffectiveUserCount)
		}
		if scenario.EffectiveExecCount != 50 {
			t.Fatalf("Expected to have %d executions, got %d", 50, scenario.EffectiveExecCount)
		}
	})
}
