package main

import (
	"github.com/ofux/deluge/deluge"
	"github.com/robertkrimen/otto"
	log "github.com/sirupsen/logrus"
	"time"
)

func main() {
	//log.SetLevel(log.DebugLevel)
	start := time.Now()
	vm := otto.New()
	script, err := vm.Compile("myscenario.js", nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	wd := deluge.NewRain("rain1", script, 10000, 1*time.Millisecond)
	log.Infof("Deluge initialized in %s", time.Now().Sub(start).String())
	wd.Run(1 * time.Second)
}
