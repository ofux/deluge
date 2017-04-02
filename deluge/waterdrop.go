package deluge

import (
	"errors"
	"github.com/robertkrimen/otto"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var halt = errors.New("Halt water drop")

type WaterDropStatus int

const (
	Virgin WaterDropStatus = iota
	InProgress
	DoneSuccess
	DoneAssertionError
)

type WaterDrop struct {
	vm            *otto.Otto
	script        *otto.Script
	client        *http.Client
	Status        WaterDropStatus
	SleepDuration time.Duration
}

func NewWaterDrop(script *otto.Script) *WaterDrop {
	vm := otto.New()
	wd := &WaterDrop{
		vm:     vm,
		script: script,
		client: http.DefaultClient,
		Status: Virgin,
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
				log.Debug(halt.Error())
				return
			}
			panic(caught) // Something else happened, repanic!
		}
	}()

	wd.Status = InProgress
	_, err := wd.vm.Run(wd.script)
	if err != nil {
		log.Fatal(err)
	}
	if wd.Status == InProgress {
		wd.Status = DoneSuccess
	}
}

func (wd *WaterDrop) Assert(call otto.FunctionCall) otto.Value {
	//fmt.Printf("Assert is %s.\n", call.Argument(0).String())
	result, err := call.Argument(0).ToBoolean()
	checkFatal(err, call)

	if !result {
		log.Debugf("Assertion failed at %s", call.CallerLocation())
		wd.Status = DoneAssertionError
		wd.vm.Interrupt <- func() {
			panic(halt)
		}
	}
	return otto.Value{}
}

func (wd *WaterDrop) Pause(call otto.FunctionCall) otto.Value {
	d, err := time.ParseDuration(call.Argument(0).String())
	checkFatal(err, call)
	wd.SleepDuration += d
	time.Sleep(d)
	return otto.Value{}
}

func (wd *WaterDrop) DoHTTP(call otto.FunctionCall) otto.Value {
	if !call.Argument(0).IsObject() {
		log.Fatalf("Expected HTTP request object at %s\n", call.CallerLocation())
	}
	jsReq := call.Argument(0).Object()

	jsUrl, err := jsReq.Get("url")
	checkFatal(err, call)
	if !jsUrl.IsDefined() {
		log.Fatalf("Invalid HTTP request: missing 'url' field at %s", call.CallerLocation())
	}

	var method = "GET"
	jsMethod, err := jsReq.Get("method")
	checkFatal(err, call)
	if jsMethod.IsDefined() {
		method = jsMethod.String()
	}

	req, err := http.NewRequest(method, jsUrl.String(), nil)
	checkFatal(err, call)

	log.Debugf("Performing HTTP request: %s %s", req.Method, req.URL.String())
	start := time.Now()
	res, err := wd.client.Do(req)
	end := time.Now()
	duration := end.Sub(start)
	if err != nil {
		log.Debugf("Request error: %s", err.Error())
	} else {
		log.Debugf("Response status: %s in %s", res.Status, duration.String())
	}

	return otto.Value{}
}

func checkFatal(err error, call otto.FunctionCall) {
	if err != nil {
		log.Fatalf("%s at %s\n", err.Error(), call.CallerLocation())
	}
}
