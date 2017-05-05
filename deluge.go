package main

import (
	"github.com/ofux/deluge-dsl/lexer"
	"github.com/ofux/deluge-dsl/parser"
	"github.com/ofux/deluge/deluge"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"time"
)

func main() {
	//log.SetLevel(log.DebugLevel)
	start := time.Now()

	prg, err := ioutil.ReadFile("minimalscenario.js")
	if err != nil {
		log.Fatalf("Could not read file %s", "minimalscenario.js")
	}
	l := lexer.New(string(prg))
	p := parser.New(l)

	program, ok := p.ParseProgram()
	if !ok {
		deluge.PrintParserErrors(p.Errors())
		return
	}

	wd := deluge.NewRain("rain1", program, 1, 2*time.Second)
	log.Infof("Deluge initialized in %s", time.Now().Sub(start).String())
	wd.Run(200 * time.Second)
}
