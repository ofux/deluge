package deluge

import (
	"github.com/ofux/deluge-dsl/ast"
	"github.com/ofux/deluge-dsl/lexer"
	"github.com/ofux/deluge-dsl/parser"
	"github.com/ofux/docilemonkey/docilemonkey"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeluge_Run(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(docilemonkey.Handler))
	defer srv.Close()

	program := compileTest(t, `
deluge("Some name", "200ms", {
    "myScenario": {
        "concurrent": 100,
        "delay": "100ms"
    }
});

scenario("myScenario", "My scenario", function () {

    http("My request", {
        "url": "`+srv.URL+`/hello/toto?s=201",
        "method": "POST"
    });

});`)

	dlg := NewDeluge(program)
	dlg.Run()

	results, err := dlg.scenarios["myScenario"].httpRecorder.GetRecords()
	if err != nil {
		t.Fatalf(err.Error())
	}

	const reqRec = "My request"
	result, ok := results[reqRec]
	if !ok {
		t.Fatalf("Expected to have some records for '%s'", reqRec)
	}
	if len(result) != 2 {
		t.Fatalf("Expected to have %d records for '%s', got %d", 2, reqRec, len(result))
	}
	if result[0].TotalCount() != 100 {
		t.Errorf("Expected to have totalCount = %d, got %d", 100, result[0].TotalCount())
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
