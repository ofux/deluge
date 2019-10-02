package core

import (
	"github.com/ofux/deluge/core/recording"
	"github.com/ofux/deluge/core/recording/recordingtest"
	"github.com/ofux/deluge/core/status"
	"github.com/ofux/deluge/repov2"
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

		compileDeluge(t, `
		deluge("foo", "Some name", "200ms", {
			"myScenario": {
				"concurrent": 100,
				"delay": "100ms"
			}
		});`)

		dlg, err := NewRunnableDeluge("foo")
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

		compileDeluge(t, `
		deluge("foo", "Some name", "20s", {
			"myScenario": {
				"concurrent": 100,
				"delay": "500ms"
			}
		});`)

		dlg, err := NewRunnableDeluge("foo")
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

		assert.Equal(t, status.DelugeInterrupted, dlg.Status)

		// Should do nothing, should not panic, should not cause race condition
		go func() {
			dlg.Interrupt()
			assert.Equal(t, status.DelugeInterrupted, dlg.Status)
		}()
		dlg.Interrupt()
		assert.Equal(t, status.DelugeInterrupted, dlg.Status)
	})

	t.Run("Run deluge with args", func(t *testing.T) {
		clearScenarioRepo()

		compileScenario(t, `
		scenario("myScenario", "My scenario with args", function (args) {

			assert(args["foo"] == "bar");
			assert(args["x"] == null);

		});`)

		compileDeluge(t, `
		deluge("foo", "Some name", "100ms", {
			"myScenario": {
				"concurrent": 10,
				"delay": "1000ms",
				"args": {
					"foo": "bar"
				}
			}
		});`)

		dlg, err := NewRunnableDeluge("foo")
		assert.NoError(t, err)

		<-dlg.Run()

		assert.Equal(t, status.DelugeDoneSuccess, dlg.Status)
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

		compileDeluge(t, `
		deluge("foo", "Some name", "20s", {
			"myScenario": {
				"concurrent": 5,
				"delay": "10ms",
				"args": {
					"foo": "bar"
				}
			}
		});`)

		dlg, err := NewRunnableDeluge("foo")
		assert.NoError(t, err)

		<-dlg.Run()

		assert.Equal(t, status.DelugeDoneError, dlg.Status)
		assert.Len(t, dlg.Scenarios["myScenario"].Errors, 5)
		assert.Equal(t, "Assertion failed", dlg.Scenarios["myScenario"].Errors[0].Message)
		assert.Equal(t, uint64(5), dlg.Scenarios["myScenario"].EffectiveUserCount)
		assert.Equal(t, uint64(15), dlg.Scenarios["myScenario"].EffectiveExecCount)
	})
}

func TestDeluge_Run_With_Errors(t *testing.T) {

	t.Run("Missing scenario", func(t *testing.T) {
		clearScenarioRepo()

		compileDeluge(t, `
		deluge("foo", "Some name", "100ms", {
			"missingScenario": {
				"concurrent": 10,
				"delay": "10ms"
			}
		});`)

		_, err := NewRunnableDeluge("foo")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "scenario 'missingScenario' is configured but not defined")
	})

	t.Run("Assert fails", func(t *testing.T) {
		clearScenarioRepo()
		compileScenario(t, `
		scenario("myScenario", "My scenario", function () {
			assert(false);
		});`)

		compileDeluge(t, `
		deluge("foo", "Some name", "100ms", {
			"myScenario": {
				"concurrent": 10,
				"delay": "10ms"
			}
		});`)

		dlg, err := NewRunnableDeluge("foo")
		assert.NoError(t, err)

		<-dlg.Run()

		assert.Equal(t, status.DelugeDoneError, dlg.Status)

		assert.Equal(t, uint64(10), dlg.Scenarios["myScenario"].EffectiveExecCount)
		assert.Equal(t, uint64(10), dlg.Scenarios["myScenario"].EffectiveUserCount)
	})

	t.Run("Error trying to modify args hash", func(t *testing.T) {
		clearScenarioRepo()
		compileScenario(t, `
		scenario("sc1", "My scenario 1", function (args) {
			args["hello"] = "world";
		});`)

		compileScenario(t, `
		scenario("sc2", "My scenario 2", function (args) {
			args["x"]++;
		});`)

		compileDeluge(t, `
		deluge("foo", "Some name", "100ms", {
			"sc1": {
				"concurrent": 10,
				"delay": "10ms",
				"args": {
					"foo": "bar"
				}
			},
			"sc2": {
				"concurrent": 10,
				"delay": "10ms",
				"args": {
					"x": 1
				}
			}
		});`)

		dlg, err := NewRunnableDeluge("foo")
		assert.NoError(t, err)

		<-dlg.Run()

		assert.Equal(t, status.DelugeDoneError, dlg.Status)

		assert.Equal(t, status.ScenarioDoneError, dlg.Scenarios["sc1"].Status)
		require.Len(t, dlg.Scenarios["sc1"].Errors, 10)
		assert.Equal(t, "hash is immutable, you cannot modify it", dlg.Scenarios["sc1"].Errors[0].Message)
		assert.Equal(t, uint64(10), dlg.Scenarios["sc1"].EffectiveExecCount)
		assert.Equal(t, uint64(10), dlg.Scenarios["sc1"].EffectiveUserCount)

		assert.Equal(t, status.ScenarioDoneError, dlg.Scenarios["sc2"].Status)
		require.Len(t, dlg.Scenarios["sc2"].Errors, 10)
		assert.Equal(t, "hash is immutable, you cannot modify it", dlg.Scenarios["sc2"].Errors[0].Message)
		assert.Equal(t, uint64(10), dlg.Scenarios["sc2"].EffectiveExecCount)
		assert.Equal(t, uint64(10), dlg.Scenarios["sc2"].EffectiveUserCount)
	})

}

func BenchmarkNewDeluge(b *testing.B) {
	clearScenarioRepo()

	compileScenario(b, `
scenario("myScenario", "My scenario", function () {

    http("My request", {
        "url": "http://localhost:8080/hello/toto"
    });

});`)

	compileDeluge(b, `
		deluge("foo", "Some name", "200ms", {
			"myScenario": {
				"concurrent": 100,
				"delay": "100ms"
			}
		});`)

	for i := 0; i < b.N; i++ {
		_, err := NewRunnableDeluge("foo")

		if err != nil {
			b.Fatal(err)
		}
	}
}

func compileDeluge(t testing.TB, script string) *CompiledDeluge {
	compiled, err := CompileDeluge(script)
	if err != nil {
		t.Fatal(err)
	}
	err = repov2.Instance.SaveDeluge(compiled.MapToPersistedDeluge())
	if err != nil {
		t.Fatal(err)
	}
	return compiled
}
