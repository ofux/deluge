package deluge

import (
	"github.com/ofux/deluge-dsl/lexer"
	"github.com/ofux/deluge-dsl/parser"
	"testing"
)

func NewSimUserTest(t *testing.T, js string) *SimUser {
	l := lexer.New(js)
	p := parser.New(l)

	program, ok := p.ParseProgram()
	if !ok {
		PrintParserErrors(p.Errors())
		t.Fatal("Parsing error(s)")
	}

	return NewSimUser("1", program)
}

func checkSimUserStatus(t *testing.T, su *SimUser, status SimUserStatus) {
	if su.Status != status {
		t.Fatalf("Bad SimUser status %d, expected %d", su.Status, status)
	}
}

func TestAssert(t *testing.T) {
	t.Run("Assert true", func(t *testing.T) {
		su := NewSimUserTest(t, `
		assert(1+1 == 2)
		`)
		su.Run()
		checkSimUserStatus(t, su, DoneSuccess)
	})

	t.Run("Assert false", func(t *testing.T) {
		su := NewSimUserTest(t, `
		assert(1+1 == 3)
		`)
		su.Run()
		checkSimUserStatus(t, su, DoneAssertionError)
	})
}

func TestPause(t *testing.T) {
	t.Run("Pause valid duration", func(t *testing.T) {
		su := NewSimUserTest(t, `
		pause("10ms")
		`)
		su.Run()
		checkSimUserStatus(t, su, DoneSuccess)

		if su.SleepDuration.String() != "10ms" {
			t.Fatalf("Expected sleep duration to be %s but was %s", "10ms", su.SleepDuration.String())
		}
	})
}
