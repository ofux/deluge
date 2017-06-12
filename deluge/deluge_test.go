package deluge

import (
	"github.com/ofux/deluge-dsl/ast"
	"github.com/ofux/deluge-dsl/lexer"
	"github.com/ofux/deluge-dsl/parser"
	"github.com/ofux/deluge/deluge/recording"
	"github.com/ofux/deluge/deluge/recording/recordingtest"
	"github.com/ofux/docilemonkey/docilemonkey"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeluge_Run(t *testing.T) {
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

	dlg := NewDeluge(program)
	dlg.Run()

	records, err := dlg.scenarios["myScenario"].httpRecorder.GetRecords()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(records.PerIteration) != 2 {
		t.Fatalf("Expected to have %d iterations, got %d", 2, len(records.PerIteration))
	}
	recordingtest.CheckHTTPRecord(t, records.Global, reqName, 200, 201, recording.Ok)
	for _, record := range records.PerIteration {
		recordingtest.CheckHTTPRecord(t, record, reqName, 100, 201, recording.Ok)
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
		NewDeluge(program)
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
