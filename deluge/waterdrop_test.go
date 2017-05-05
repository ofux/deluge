package deluge

import (
	"github.com/ofux/deluge-dsl/lexer"
	"github.com/ofux/deluge-dsl/parser"
	"testing"
	"time"
)

func NewWaterDropTest(t *testing.T, js string) *WaterDrop {
	l := lexer.New(js)
	p := parser.New(l)

	program, ok := p.ParseProgram()
	if !ok {
		PrintParserErrors(p.Errors())
		t.Fatal("Parsing error(s)")
	}

	return NewWaterDrop("1", program)
}

func checkWDStatus(t *testing.T, wd *WaterDrop, status WaterDropStatus) {
	if wd.Status != status {
		t.Fatalf("Bad WaterDrop status %d, expected %d", wd.Status, status)
	}
}

func TestAssert(t *testing.T) {
	t.Run("Assert true", func(t *testing.T) {
		wd := NewWaterDropTest(t, `
		assert(1+1 == 2)
		`)
		wd.Run()
		checkWDStatus(t, wd, DoneSuccess)
	})

	t.Run("Assert false", func(t *testing.T) {
		wd := NewWaterDropTest(t, `
		assert(1+1 == 3)
		`)
		wd.Run()
		checkWDStatus(t, wd, DoneAssertionError)
	})
}

func TestPause(t *testing.T) {
	t.Run("Pause valid duration", func(t *testing.T) {
		wd := NewWaterDropTest(t, `
		pause("10ms")
		`)
		wd.Run()
		checkWDStatus(t, wd, DoneSuccess)

		if wd.SleepDuration.String() != "10ms" {
			t.Fatalf("Expected sleep duration to be %s but was %s", "10ms", wd.SleepDuration.String())
		}
	})
}

func BenchmarkNewRain(b *testing.B) {
	l := lexer.New(`
let updatedOrder = doHTTP({
    "url": "http://localhost:8080/hello/toto"
});
	`)
	p := parser.New(l)

	program, ok := p.ParseProgram()
	if !ok {
		PrintParserErrors(p.Errors())
		b.Fatal("Parsing error(s)")
	}

	for i := 0; i < b.N; i++ {
		NewRain("rain", program, 1000, time.Second)
	}
}
