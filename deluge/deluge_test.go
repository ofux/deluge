package deluge

import (
	"github.com/ofux/deluge-dsl/lexer"
	"github.com/ofux/deluge-dsl/parser"
	"testing"
)

func BenchmarkNewDeluge(b *testing.B) {
	l := lexer.New(`
deluge("Some name", {
    "myScenario": {
        "concurrent": 100,
        "delay": "2s"
    }
});

scenario("myScenario", "My scenario", function () {

    http({
        "url": "http://localhost:8080/hello/toto"
    });

});
	`)
	p := parser.New(l)

	program, ok := p.ParseProgram()
	if !ok {
		PrintParserErrors(p.Errors())
		b.Fatal("Parsing error(s)")
	}

	for i := 0; i < b.N; i++ {
		NewDeluge(program)
	}
}
