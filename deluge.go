package main

import (
	"github.com/ofux/deluge/deluge"
	"github.com/robertkrimen/otto"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)

	vm := otto.New()
	script, err := vm.Compile("myscenario.js", nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	wd := deluge.NewRain(script, 10)
	wd.Run()
}
