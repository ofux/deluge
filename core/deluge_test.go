package core

import (
	"github.com/ofux/deluge/core/recording"
	"github.com/ofux/deluge/core/recording/recordingtest"
	"github.com/ofux/deluge/dsl/ast"
	"github.com/ofux/deluge/dsl/lexer"
	"github.com/ofux/deluge/dsl/parser"
	"github.com/ofux/docilemonkey/docilemonkey"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDeluge_Run(t *testing.T) {

	t.Run("Run a simple deluge", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(docilemonkey.Handler))
		defer srv.Close()

		const reqName = "My request"
		program := compileTest(t, `
		deluge("Some name", "200ms", {
			"myScenario": {
				"concurrent": 100,
				"delay": "100ms"
			}
		});

		scenario("myScenario", "My scenario", function () {

			http("`+reqName+`", {
				"url": "`+srv.URL+`/hello/toto?s=201",
				"method": "POST"
			});

		});`)

		dlg, err := NewDeluge("foo", program)
		assert.NoError(t, err)
		<-dlg.Run()

		records, err := dlg.Scenarios["myScenario"].httpRecorder.GetRecords()
		if err != nil {
			t.Fatalf(err.Error())
		}

		if len(records.PerIteration) < 1 || len(records.PerIteration) > 2 {
			t.Fatalf("Expected to have 1 or 2 iterations, got %d", len(records.PerIteration))
		}
		recordingtest.CheckHTTPRecord(t, records.Global, reqName, int64(dlg.Scenarios["myScenario"].EffectiveExecCount), 201, recording.Ok)
	})

	t.Run("Run and interrupt a deluge", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(docilemonkey.Handler))
		defer srv.Close()

		const reqName = "My request"
		program := compileTest(t, `
		deluge("Some name", "20s", {
			"myScenario": {
				"concurrent": 100,
				"delay": "500ms"
			}
		});

		scenario("myScenario", "My scenario", function () {

			http("`+reqName+`", {
				"url": "`+srv.URL+`/hello/toto?s=200",
				"method": "PUT"
			});

		});`)

		dlg, err := NewDeluge("foo", program)
		assert.NoError(t, err)

		start := time.Now()
		done := dlg.Run()
		time.Sleep(100 * time.Millisecond)
		dlg.Interrupt()
		<-done
		elapsedTime := time.Now().Sub(start)
		if elapsedTime.Seconds() > 10 {
			t.Errorf("Looks like deluge was not interrupted")
		}

		assert.Equal(t, DelugeInterrupted, dlg.Status)

		// Should do nothing, should not panic, should not cause race condition
		go func() {
			dlg.Interrupt()
			assert.Equal(t, DelugeInterrupted, dlg.Status)
		}()
		dlg.Interrupt()
		assert.Equal(t, DelugeInterrupted, dlg.Status)
	})
}

func TestDeluge_Run_With_Deluge_Errors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			`
			deluge("Some name", "200ms", {
				"myScenario": {
					"concurrent": 100,
					"delay": "100ms"
				}
			});`,
			"Scenario 'myScenario' is configured but not defined.",
		},
		{
			`deluge("200ms", {
				"myScenario": {
					"concurrent": 100,
					"delay": "100ms"
				}
			});`,
			"RUNTIME ERROR: Expected 3 arguments at",
		},
		{
			`deluge(1, "200ms", {
				"myScenario": {
					"concurrent": 100,
					"delay": "100ms"
				}
			});`,
			"RUNTIME ERROR: Expected 1st argument to be a string at",
		},
		{
			`deluge("Some name", 200, {
				"myScenario": {
					"concurrent": 100,
					"delay": "100ms"
				}
			});`,
			"RUNTIME ERROR: Expected 2nd argument to be a string at",
		},
		{
			`deluge("Some name", "200", {
				"myScenario": {
					"concurrent": 100,
					"delay": "100ms"
				}
			});`,
			"RUNTIME ERROR: Expected 2nd argument to be a valid duration at",
		},
		{
			`deluge("Some name", "200ms", "bad");`,
			"RUNTIME ERROR: Expected 3rd argument to be an object at",
		},
		{
			`deluge("Some name", "200ms", {
				"myScenario": "bad"
			});`,
			"RUNTIME ERROR: Expected scenario configuration to be an object at",
		},
		{
			`deluge("Some name", "200ms", {
				"myScenario": {
					"delay": "100ms"
				}
			});`,
			"RUNTIME ERROR: Expected 'concurrent' value in configuration at",
		},
		{
			`deluge("Some name", "200ms", {
				"myScenario": {
					"concurrent": 100
				}
			});`,
			"RUNTIME ERROR: Expected 'delay' value in configuration at",
		},
		{
			`deluge("Some name", "200ms", {
				"myScenario": {
					"concurrent": "100",
					"delay": "100ms"
				}
			});`,
			"RUNTIME ERROR: Expected 'concurrent' value to be an integer in configuration at",
		},
		{
			`deluge("Some name", "200ms", {
				"myScenario": {
					"concurrent": 100,
					"delay": 100
				}
			});`,
			"RUNTIME ERROR: Expected 'delay' value to be a valid duration in configuration at",
		},
		{
			`deluge("Some name", "200ms", {
				"myScenario": {
					"concurrent": 100,
					"delay": "100"
				}
			});`,
			"RUNTIME ERROR: Expected 'delay' value to be a valid duration in configuration at",
		},
	}

	for _, tt := range tests {
		program := compileTest(t, tt.input)
		_, err := NewDeluge("foo", program)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), tt.expected)
	}
}

func TestDeluge_Run_With_Scenario_Errors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			`
			scenario("My scenario", function () {})`,
			"RUNTIME ERROR: Expected 3 arguments at",
		},
		{
			`
			scenario(1, "My scenario", function () {})`,
			"RUNTIME ERROR: Expected 1st argument to be a string at",
		},
		{
			`
			scenario("myScenario", 1, function () {})`,
			"RUNTIME ERROR: Expected 2nd argument to be a string at",
		},
		{
			`
			scenario("myScenario", "My scenario", "bad")`,
			"RUNTIME ERROR: Expected 3rd argument to be a function at",
		},
		{
			`
			scenario("myScenario", "My scenario 1", function () {});
			scenario("myScenario", "My scenario 2", function () {});`,
			"RUNTIME ERROR: Scenario 'myScenario' is already defined",
		},
	}

	for _, tt := range tests {
		program := compileTest(t, tt.input)
		_, err := NewDeluge("foo", program)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), tt.expected)
	}
}

func BenchmarkNewDeluge(b *testing.B) {

	program := compileTest(b, `
deluge("Some name", "200ms", {
    "myScenario": {
        "concurrent": 100,
        "delay": "100ms"
    }
});

scenario("myScenario", "My scenario", function () {

    http("My request", {
        "url": "http://localhost:8080/hello/toto"
    });

});`)

	for i := 0; i < b.N; i++ {
		NewDeluge("foo", program)
	}
}

func compileTest(t testing.TB, script string) *ast.Program {
	l := lexer.New(script)
	p := parser.New(l)

	program, ok := p.ParseProgram()
	if !ok {
		PrintParserErrors(p.Errors())
		t.Fatal("Parsing error(s)")
	}

	return program
}
