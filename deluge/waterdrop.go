package deluge

import (
	"errors"
	"fmt"
	"github.com/robertkrimen/otto"
	log "github.com/sirupsen/logrus"
	"time"
)

var halt = errors.New("Halt water drop")

type WaterDrop struct {
	vm           *otto.Otto
	script       *otto.Script
	IsSuccessful bool
}

func NewWaterDrop(script *otto.Script) *WaterDrop {
	vm := otto.New()
	wd := &WaterDrop{
		vm:     vm,
		script: script,
	}
	vm.Interrupt = make(chan func(), 1) // The buffer prevents blocking

	vm.Set("assert", wd.Assert)
	vm.Set("pause", wd.Pause)
	vm.Set("doHTTP", wd.DoHTTP)

	return wd
}

func (wd *WaterDrop) Run() {
	defer func() {
		if caught := recover(); caught != nil {
			if caught == halt {
				return
			}
			panic(caught) // Something else happened, repanic!
		}
	}()

	wd.vm.Run(wd.script)
}

func (wd *WaterDrop) Assert(call otto.FunctionCall) otto.Value {
	//fmt.Printf("Assert is %s.\n", call.Argument(0).String())
	result, err := call.Argument(0).ToBoolean()
	if err != nil {
		log.Fatalf("%s at %s\n", err.Error(), call.CallerLocation())
	}
	wd.IsSuccessful = result
	if !result {
		wd.vm.Interrupt <- func() {
			panic(halt)
		}
	}
	return otto.Value{}
}

func (wd *WaterDrop) Pause(call otto.FunctionCall) otto.Value {
	d, err := time.ParseDuration(call.Argument(0).String())
	if err != nil {
		log.Fatalf("%s at %s\n", err.Error(), call.CallerLocation())
	}
	time.Sleep(d)
	return otto.Value{}
}

func (wd *WaterDrop) DoHTTP(call otto.FunctionCall) otto.Value {
	v, _ := call.Argument(0).Object().Get("url")

	fmt.Printf("doHTTP: %s\n", v.String())
	return otto.Value{}
}
