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

func TestDeluge_NewRunnableDeluge(t *testing.T) {
	t.Run("Create runnable deluge", func(t *testing.T) {
		clearScenarioRepo()

		compileScenario(t, `
		scenario("myScenario1", "My scenario with args", function (args) {
		});`)

		compileScenario(t, `
		scenario("myScenario2", "My scenario with args", function (args) {
		});`)

		compileDeluge(t, `
		deluge("foo", "Some name", "100ms", {
			"myScenario1": {
				"concurrent": 10,
				"delay": "1000ms"
			},
			"myScenario2": {
				"concurrent": 10,
				"delay": "1000ms"
			}
		});`)

		dlg, err := NewRunnableDeluge("foo")
		assert.NoError(t, err)
		require.NotNil(t, dlg.Scenarios["myScenario1"])
		require.NotNil(t, dlg.Scenarios["myScenario2"])
		assert.Equal(t, "myScenario1", dlg.Scenarios["myScenario1"].GetScenarioDefinition().ID)
		assert.Equal(t, "myScenario2", dlg.Scenarios["myScenario2"].GetScenarioDefinition().ID)
	})

	t.Run("Create runnable deluge with unknown scenario", func(t *testing.T) {
		clearScenarioRepo()

		compileScenario(t, `
		scenario("myScenario1", "My scenario with args", function (args) {
		});`)

		compileDeluge(t, `
		deluge("foo", "Some name", "100ms", {
			"myScenario1": {
				"concurrent": 10,
				"delay": "1000ms"
			},
			"myScenario2": {
				"concurrent": 10,
				"delay": "1000ms"
			}
		});`)

		_, err := NewRunnableDeluge("foo")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "scenario 'myScenario2' is configured but not defined")
	})

	t.Run("Create runnable deluge with bad scenario", func(t *testing.T) {
		clearScenarioRepo()

		err := repov2.Instance.SaveScenario(&repov2.PersistedScenario{
			ID:     "myScenario1",
			Name:   "My scenario",
			Script: "bad script",
		})
		require.NoError(t, err)

		compileDeluge(t, `
		deluge("foo", "Some name", "100ms", {
			"myScenario1": {
				"concurrent": 10,
				"delay": "1000ms"
			}
		});`)

		_, err = NewRunnableDeluge("foo")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to recompile scenario myScenario1")
	})

	t.Run("Create runnable deluge with bad ID", func(t *testing.T) {
		clearScenarioRepo()

		_, err := NewRunnableDeluge("foo")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deluge with ID 'foo' does not exist")
	})

	t.Run("Create runnable deluge with bad deluge script", func(t *testing.T) {
		clearScenarioRepo()

		compileScenario(t, `
		scenario("myScenario1", "My scenario with args", function (args) {
		});`)

		err := repov2.Instance.SaveDeluge(&repov2.PersistedDeluge{
			ID:             "foo",
			Name:           "My deluge",
			Script:         "bad script",
			GlobalDuration: time.Minute,
		})
		require.NoError(t, err)

		_, err = NewRunnableDeluge("foo")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to recompile deluge with ID 'foo'")
	})
}

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

		assertStatuses(t, dlg, status.DelugeVirgin, status.DelugeInProgress, status.DelugeDoneSuccess)

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

		assert.Equal(t, status.DelugeInterrupted, dlg.runStatus)

		// Should do nothing, should not panic, should not cause race condition
		go func() {
			dlg.Interrupt()
			assert.Equal(t, status.DelugeInterrupted, dlg.runStatus)
		}()
		dlg.Interrupt()
		assert.Equal(t, status.DelugeInterrupted, dlg.runStatus)

		assertStatuses(t, dlg, status.DelugeVirgin, status.DelugeInProgress, status.DelugeInterrupted)
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

		assertStatuses(t, dlg, status.DelugeVirgin, status.DelugeInProgress, status.DelugeDoneSuccess)

		assert.Equal(t, status.DelugeDoneSuccess, dlg.runStatus)
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

		assertStatuses(t, dlg, status.DelugeVirgin, status.DelugeInProgress, status.DelugeDoneError)

		assert.Equal(t, status.DelugeDoneError, dlg.runStatus)
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

	t.Run("Run same deluge twice", func(t *testing.T) {
		clearScenarioRepo()
		compileScenario(t, `
		scenario("myScenario", "My scenario", function () {
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

		assertStatuses(t, dlg, status.DelugeVirgin, status.DelugeInProgress, status.DelugeDoneSuccess)
		assert.Equal(t, status.DelugeDoneSuccess, dlg.runStatus)

		<-dlg.Run()

		// no more status change
		assertStatuses(t, dlg)
		assert.Equal(t, status.DelugeDoneSuccess, dlg.runStatus)
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

		assertStatuses(t, dlg, status.DelugeVirgin, status.DelugeInProgress, status.DelugeDoneError)
		assert.Equal(t, status.DelugeDoneError, dlg.runStatus)

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

		assertStatuses(t, dlg, status.DelugeVirgin, status.DelugeInProgress, status.DelugeDoneError)
		assert.Equal(t, status.DelugeDoneError, dlg.runStatus)

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

func assertStatuses(t testing.TB, dlg *RunnableDeluge, expected ...status.DelugeStatus) {
	i := 0
	for st := range dlg.OnStatusChangeChan() {
		if i >= len(expected) {
			t.Errorf("Expected no more status in channel but got %s", st)
			return
		}
		if expected[i] != st {
			t.Errorf("Expected %dth status in channel to be %s but got %s", i+1, expected[i], st)
		}
		i++
	}
}
