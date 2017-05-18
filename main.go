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

	dlg := deluge.NewDeluge(program)
	log.Infof("Deluge initialized in %s", time.Now().Sub(start).String())
	dlg.Run()
}
