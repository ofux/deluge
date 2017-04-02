package deluge

import (
	"github.com/robertkrimen/otto"
	"testing"
)

func NewWaterDropTest(t *testing.T, js string) *WaterDrop {
	vm := otto.New()
	script, err := vm.Compile("", js)
	if err != nil {
		t.Fatal(err)
	}
	return NewWaterDrop(script)
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
