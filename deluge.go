package main

import (
	"github.com/robertkrimen/otto"
	log "github.com/sirupsen/logrus"
	"github.com/ofux/deluge/deluge"
)

func main() {
	vm := otto.New()
	script, err := vm.Compile("myscenario.js", nil)
	if err != nil {
		log.Fatal(err)
		return;
	}

	wd := deluge.NewRain(script, 10)
	wd.Run()
}
