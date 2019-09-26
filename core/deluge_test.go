package core

import (
	"github.com/ofux/deluge/core/recording"
	"github.com/ofux/deluge/core/recording/recordingtest"
	"github.com/ofux/docilemonkey/docilemonkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDeluge_Run(t *testing.T) {

	t.Run("Run deluge with HTTP requests", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(docilemonkey.Handler))
		defer srv.Close()
		clearScenarioRepo()

		const reqName = "My request"
		compileScenario(t, `
		scenario("myScenario", "My scenario", function () {

			http("`+reqName+`", {
				"url": "`+srv.URL+`/hello/toto?s=201",
				"method": "POST"
			});

		});`)

		dlg, err := NewDeluge("foo", `
		deluge("Some name", "200ms", {
			"myScenario": {
				"concurrent": 100,
				"delay": "100ms"
			}
		});`)
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
		clearScenarioRepo()

		const reqName = "My request"
		compileScenario(t, `
		scenario("myScenario", "My scenario", function () {

			http("`+reqName+`", {
				"url": "`+srv.URL+`/hello/toto?s=200",
				"method": "PUT"
			});

		});`)

		dlg, err := NewDeluge("foo", `
		deluge("Some name", "20s", {
			"myScenario": {
				"concurrent": 100,
				"delay": "500ms"
			}
		});`)
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

	t.Run("Run deluge with args", func(t *testing.T) {
		clearScenarioRepo()

		compileScenario(t, `
		scenario("myScenario", "My scenario with args", function (args) {

			assert(args["foo"] == "bar");
			assert(args["x"] == null);

		});`)

		dlg, err := NewDeluge("foo", `
		deluge("Some name", "100ms", {
			"myScenario": {
				"concurrent": 10,
				"delay": "1000ms",
				"args": {
					"foo": "bar"
				}
			}
		});`)
		assert.NoError(t, err)

		<-dlg.Run()

		assert.Equal(t, DelugeDoneSuccess, dlg.Status)
		assert.Equal(t, uint64(10), dlg.Scenarios["myScenario"].EffectiveExecCount)
		assert.Equal(t, uint64(10), dlg.Scenarios["myScenario"].EffectiveUserCount)
	})

	t.Run("Run deluge with args and session", func(t *testing.T) {
		clearScenarioRepo()

		compileScenario(t, `
		scenario("myScenario", "My scenario", function (args, session) {

			assert(args["foo"] == "bar");

			if (session["count"] == null) {
				session["count"] = 0;
			}
			session["count"]++;
			assert(session["count"] < 3);

		});`)

		dlg, err := NewDeluge("foo", `
		deluge("Some name", "20s", {
			"myScenario": {
				"concurrent": 5,
				"delay": "10ms",
				"args": {
					"foo": "bar"
				}
			}
		});`)
		assert.NoError(t, err)

		<-dlg.Run()

		assert.Equal(t, DelugeDoneError, dlg.Status)
		assert.Len(t, dlg.Scenarios["myScenario"].Errors, 5)
		assert.Equal(t, "Assertion failed", dlg.Scenarios["myScenario"].Errors[0].Message)
		assert.Equal(t, uint64(5), dlg.Scenarios["myScenario"].EffectiveUserCount)
		assert.Equal(t, uint64(15), dlg.Scenarios["myScenario"].EffectiveExecCount)
	})
}

func TestDeluge_Run_With_Errors(t *testing.T) {

	t.Run("Assert fails", func(t *testing.T) {
		clearScenarioRepo()
		compileScenario(t, `
		scenario("myScenario", "My scenario", function () {
			assert(false);
		});`)

		dlg, err := NewDeluge("foo", `
		deluge("Some name", "100ms", {
			"myScenario": {
				"concurrent": 10,
				"delay": "10ms"
			}
		});`)
		assert.NoError(t, err)

		<-dlg.Run()

		assert.Equal(t, DelugeDoneError, dlg.Status)

		assert.Equal(t, uint64(10), dlg.Scenarios["myScenario"].EffectiveExecCount)
		assert.Equal(t, uint64(10), dlg.Scenarios["myScenario"].EffectiveUserCount)
	})

	t.Run("Error trying to modify args hash", func(t *testing.T) {
		clearScenarioRepo()
		compileScenario(t, `
		scenario("s1", "My scenario 1", function (args) {
			args["hello"] = "world";
		});`)

		compileScenario(t, `
		scenario("s2", "My scenario 2", function (args) {
			args["x"]++;
		});`)

		dlg, err := NewDeluge("foo", `
		deluge("Some name", "100ms", {
			"s1": {
				"concurrent": 10,
				"delay": "10ms",
				"args": {
					"foo": "bar"
				}
			},
			"s2": {
				"concurrent": 10,
				"delay": "10ms",
				"args": {
					"x": 1
				}
			}
		});`)
		assert.NoError(t, err)

		<-dlg.Run()

		assert.Equal(t, DelugeDoneError, dlg.Status)

		assert.Equal(t, ScenarioDoneError, dlg.Scenarios["s1"].Status)
		require.Len(t, dlg.Scenarios["s1"].Errors, 10)
		assert.Equal(t, "hash is immutable, you cannot modify it", dlg.Scenarios["s1"].Errors[0].Message)
		assert.Equal(t, uint64(10), dlg.Scenarios["s1"].EffectiveExecCount)
		assert.Equal(t, uint64(10), dlg.Scenarios["s1"].EffectiveUserCount)

		assert.Equal(t, ScenarioDoneError, dlg.Scenarios["s2"].Status)
		require.Len(t, dlg.Scenarios["s2"].Errors, 10)
		assert.Equal(t, "hash is immutable, you cannot modify it", dlg.Scenarios["s2"].Errors[0].Message)
		assert.Equal(t, uint64(10), dlg.Scenarios["s2"].EffectiveExecCount)
		assert.Equal(t, uint64(10), dlg.Scenarios["s2"].EffectiveUserCount)
	})

}

func TestDeluge_New_With_Deluge_Errors(t *testing.T) {
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
			"scenario 'myScenario' is configured but not defined",
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
		{
			`deluge("Some name", "200ms", {
				"myScenario": {
					"concurrent": 100,
					"delay": "100ms",
					"args": "foobar"
				}
			});`,
			"RUNTIME ERROR: Expected 'args' to be an object at",
		},
		{
			`deluge("Some name", "200ms", {}); deluge("Some other name", "200ms", {});`,
			"RUNTIME ERROR: Expected only one deluge definition at",
		},
	}

	for _, tt := range tests {
		clearScenarioRepo()
		_, err := NewDeluge("foo", tt.input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), tt.expected)
	}
}

func BenchmarkNewDeluge(b *testing.B) {
	clearScenarioRepo()

	compileScenario(b, `
scenario("myScenario", "My scenario", function () {

    http("My request", {
        "url": "http://localhost:8080/hello/toto"
    });

});`)

	for i := 0; i < b.N; i++ {
		_, err := NewDeluge("foo", `
		deluge("Some name", "200ms", {
			"myScenario": {
				"concurrent": 100,
				"delay": "100ms"
			}
		});`)

		if err != nil {
			b.Fatal(err)
		}
	}
}
