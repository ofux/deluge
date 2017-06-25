package core

import (
	"github.com/ofux/deluge-dsl/ast"
	"github.com/ofux/deluge-dsl/lexer"
	"github.com/ofux/deluge-dsl/parser"
	"github.com/ofux/deluge/core/recording"
	"github.com/ofux/deluge/core/recording/recordingtest"
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

		dlg := NewDeluge("foo", program)
		<-dlg.Run()

		records, err := dlg.Scenarios["myScenario"].httpRecorder.GetRecords()
		if err != nil {
			t.Fatalf(err.Error())
		}

		if len(records.PerIteration) != 2 {
			t.Fatalf("Expected to have %d iterations, got %d", 2, len(records.PerIteration))
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

		dlg := NewDeluge("foo", program)
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
	})
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
